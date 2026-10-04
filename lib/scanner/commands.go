package scanner

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"redscout/lib"
	"redscout/models"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

func (s *Scanner) FetchSlowLog() error {
	s.muRedis.Lock()
	defer s.muRedis.Unlock()

	slowLog, err := s.redis.SlowLogGet(s.ctx, s.Config.TopK).Result()

	models.SlowLogList(slowLog).Sort("Timestamp")
	s.State.SlowLogs = slowLog
	s.State.Updates <- s.State

	return err
}

func (s *Scanner) FetchRedisInfo() error {
	s.muRedis.Lock()
	defer s.muRedis.Unlock()

	infoStr, err := s.redis.Info(s.ctx).Result()
	if err != nil {
		return err
	}

	parsed := models.ParseInfo(infoStr)

	dHits := parsed.Stats.KeyspaceHits - s.State.RedisInfo.Stats.KeyspaceHits
	dMisses := parsed.Stats.KeyspaceMisses - s.State.RedisInfo.Stats.KeyspaceMisses

	if (dHits + dMisses) > 0 {
		parsed.Computed.HitRate = float64(dHits) / float64(dHits+dMisses)
	} else {
		parsed.Computed.HitRate = s.State.RedisInfo.Computed.HitRate
	}

	if !s.State.LastInfoCheck.IsZero() {
		currCPUTime := parsed.CPU.SystemTime + parsed.CPU.UserTime
		prevCPUTime := s.State.RedisInfo.CPU.UserTime + s.State.RedisInfo.CPU.SystemTime
		parsed.Computed.CPUUsage = (currCPUTime - prevCPUTime) * 1000 / float64(time.Now().UnixMilli()-s.State.LastInfoCheck.UnixMilli())
	}

	s.State.LastInfoCheck = time.Now()
	s.State.RedisInfo = &parsed

	s.State.Updates <- s.State

	return nil
}

func (s *Scanner) scanKeys() ([]string, error) {
	//Batch size for scanning keys
	scanSize := int64(lib.ScanBatchSize)

	var (
		collected []string
		scanned   int64
	)

	for {
		res, next, err := s.redis.Scan(s.ctx, s.State.Cursor, "*", scanSize).Result()
		if err != nil {
			log.Printf("scan error: %v", err)
			return nil, err
		}

		collected = append(collected, res...)
		scanned += int64(len(res))
		s.State.Cursor = next
		s.State.ScannedKeys += int64(len(res))
		s.State.ScanProgress = min(float64(scanned)/float64(s.Config.KeysScanSize)*50, 50)
		s.State.Updates <- s.State
		if next == 0 || scanned >= s.Config.KeysScanSize {
			break
		}
	}
	return collected, nil
}

type trip struct {
	key     string
	mem     *redis.IntCmd
	ttl     *redis.DurationCmd
	typeCmd *redis.StatusCmd
}

func (s *Scanner) ScanMemory() error {
	s.updateStatus("Scanning memory")

	s.muRedis.Lock()
	defer s.muRedis.Unlock()

	log.Printf("Memory scan started")

	s.State.TotalKeysToScan = s.Config.KeysScanSize
	s.updateStatus("Collecting keys")

	keys, err := s.scanKeys()
	if err != nil {
		return err
	}

	s.State.TotalKeysToScan = int64(len(keys))
	s.updateStatus("Scanning memory")

	var buf bytes.Buffer
	processedKeys := int64(0)
	for i := 0; i < len(keys); i += lib.MemoryPipeBatchSize {
		pipe := s.redis.Pipeline()

		keyBatch := keys[i:min(i+lib.MemoryPipeBatchSize, len(keys))]

		var trips []trip
		for _, key := range keyBatch {
			tr := trip{key: key}
			tr.mem = pipe.MemoryUsage(s.ctx, key)
			tr.ttl = pipe.TTL(s.ctx, key)
			tr.typeCmd = pipe.Type(s.ctx, key)
			trips = append(trips, tr)
		}

		_, _ = pipe.Exec(s.ctx)

		for _, tr := range trips {
			xMem, e1 := tr.mem.Result()
			xTtl, e2 := tr.ttl.Result()
			xType, e3 := tr.typeCmd.Result()
			if e1 != nil || e2 != nil || e3 != nil {
				continue
			}
			fmt.Fprintf(&buf, "%s %d %d %s\n", tr.key, xMem, int64(xTtl.Seconds()), xType)
		}
		if err := appendLog(&s.muScan, s.scanFile, &buf); err != nil {
			return fmt.Errorf("failed to write scan file: %w", err)
		}

		processedKeys += int64(len(keyBatch))
		s.State.ScanProgress = 50 + min(float64(processedKeys)/float64(len(keys))*50, 49)
		s.State.Updates <- s.State
	}

	s.State.ScanProgress = 100
	log.Printf("Memory scan completed; scanned %d keys", len(keys))
	s.updateStatus("Memory scan completed")
	return nil
}

func (s *Scanner) MonitorOps() error {
	s.muRedis.Lock()
	defer s.muRedis.Unlock()
	s.updateStatus("Monitoring operations")

	log.Printf("Ops monitor started for %v", s.Config.MonitorDuration)

	s.State.MonitorStartTime = time.Now()
	s.State.MonitorDurationTotal = s.Config.MonitorDuration
	s.State.MonitorProgress = 0

	//Buffered channel to handle Redis monitor output
	ch := make(chan string, 10000)

	client, err := lib.RedisClientFromConfig(s.Config)
	if err != nil {
		log.Printf("Error creating Redis redis for monitoring: %v", err)
		return err
	}
	defer client.Close()

	ctxTimeout, cancel := context.WithTimeout(s.ctx, s.Config.MonitorDuration)
	defer cancel()

	monitor := client.Monitor(ctxTimeout, ch)
	monitor.Start()
	defer func() {
		monitor.Stop()
		go drainUntilIdle(ch)
	}()

	var buf bytes.Buffer

	progressTicker := time.NewTicker(100 * time.Millisecond)
	defer progressTicker.Stop()

	for {
		select {
		case line, ok := <-ch:
			if !ok {
				continue
			}
			parts := strings.Split(line, "\"")
			if len(parts) < 2 {
				continue
			}
			cmd := strings.ToLower(parts[1])
			if cmd == "eval" || cmd == "evalsha" {
				continue
			}

			// Extract keys from quoted args (parts[3], parts[5], parts[7], ...)
			var keys []string
			switch cmd {
			case "mset", "msetnx":
				// MSET key1 val1 key2 val2 ... — every other arg is a key
				for j := 3; j < len(parts); j += 4 {
					keys = append(keys, parts[j])
				}
			case "mget", "del", "unlink", "exists", "touch", "watch":
				// All args are keys
				for j := 3; j < len(parts); j += 2 {
					keys = append(keys, parts[j])
				}
			default:
				if len(parts) >= 4 {
					keys = append(keys, parts[3])
				}
			}

			for _, key := range keys {
				buf.WriteString(key)
				buf.WriteByte(' ')
				buf.WriteString(cmd)
				buf.WriteByte('\n')
			}
		case <-progressTicker.C:
			if err := appendLog(&s.muMonitor, s.monitorFile, &buf); err != nil {
				return fmt.Errorf("failed to write monitor file: %w", err)
			}
			elapsed := time.Since(s.State.MonitorStartTime)
			s.State.MonitorProgress = min(float64(elapsed)/float64(s.Config.MonitorDuration)*100, 100)
			s.State.Updates <- s.State
		case <-ctxTimeout.Done():
			if err := appendLog(&s.muMonitor, s.monitorFile, &buf); err != nil {
				return fmt.Errorf("failed to write monitor file: %w", err)
			}
			s.State.MonitorProgress = 100
			s.State.TotalMonitorDuration += s.Config.MonitorDuration
			s.State.Updates <- s.State
			s.updateStatus("Monitoring completed")
			log.Printf("Monitoring completed")
			return nil
		}
	}
}

func appendLog(mu *sync.Mutex, f *os.File, buf *bytes.Buffer) error {
	if buf.Len() == 0 {
		return nil
	}
	mu.Lock()
	defer mu.Unlock()
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return err
	}
	_, err := buf.WriteTo(f)
	buf.Reset()
	return err
}

func drainUntilIdle(ch chan string) {
	for {
		select {
		case <-ch:
		case <-time.After(time.Second):
			return
		}
	}
}

// FindKeysWithPrefix returns keys from the scan log matching the current prefix
func (s *Scanner) FindKeysWithPrefix(prefix models.Key, limit int) []string {
	s.muScan.Lock()
	defer s.muScan.Unlock()

	_, _ = s.scanFile.Seek(0, io.SeekStart)
	scanner := bufio.NewScanner(s.scanFile)

	var keys []string
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) != 4 {
			continue
		}
		key := s.kp.NewKey(parts[0], false)
		if s.kp.IsA(key, prefix) {
			keys = append(keys, parts[0])
			if len(keys) >= limit {
				break
			}
		}
	}
	return keys
}

func (s *Scanner) FetchKeyValue(key string) error {
	s.muRedis.Lock()
	defer s.muRedis.Unlock()

	ctx := s.ctx

	// Get type
	keyType, err := s.redis.Type(ctx, key).Result()
	if err != nil {
		return err
	}

	info := &models.KeyValueInfo{
		Key:  key,
		Type: keyType,
	}

	// Get TTL
	ttl, err := s.redis.TTL(ctx, key).Result()
	if err == nil {
		info.TTL = int64(ttl.Seconds())
	}

	// Get memory usage
	mem, err := s.redis.MemoryUsage(ctx, key).Result()
	if err == nil {
		info.Size = mem
	}

	// Get value based on type
	switch keyType {
	case "string":
		val, err := s.redis.Get(ctx, key).Result()
		if err == nil {
			info.Length = int64(len(val))
			info.Value = val
		}
	case "list":
		length, err := s.redis.LLen(ctx, key).Result()
		if err == nil {
			info.Length = length
		}
		vals, err := s.redis.LRange(ctx, key, 0, 9).Result()
		if err == nil {
			info.Value = strings.Join(vals, "\n")
			if length > 10 {
				info.Value += fmt.Sprintf("\n... (%d more)", length-10)
			}
		}
	case "set":
		length, err := s.redis.SCard(ctx, key).Result()
		if err == nil {
			info.Length = length
		}
		vals, err := s.redis.SRandMemberN(ctx, key, 10).Result()
		if err == nil {
			info.Value = strings.Join(vals, "\n")
			if length > 10 {
				info.Value += fmt.Sprintf("\n... (%d more)", length-10)
			}
		}
	case "hash":
		length, err := s.redis.HLen(ctx, key).Result()
		if err == nil {
			info.Length = length
		}
		// Use HScan to avoid loading entire hash into memory
		var cursor uint64
		scanKeys, _, err := s.redis.HScan(ctx, key, cursor, "*", 10).Result()
		if err == nil {
			lines := make([]string, 0, 10)
			// HScan returns alternating key, value pairs
			for i := 0; i+1 < len(scanKeys) && len(lines) < 10; i += 2 {
				lines = append(lines, fmt.Sprintf("%s: %s", scanKeys[i], scanKeys[i+1]))
			}
			if length > int64(len(lines)) {
				lines = append(lines, fmt.Sprintf("... (%d more)", length-int64(len(lines))))
			}
			info.Value = strings.Join(lines, "\n")
		}
	case "zset":
		length, err := s.redis.ZCard(ctx, key).Result()
		if err == nil {
			info.Length = length
		}
		vals, err := s.redis.ZRangeWithScores(ctx, key, 0, 9).Result()
		if err == nil {
			lines := make([]string, 0, len(vals))
			for _, z := range vals {
				lines = append(lines, fmt.Sprintf("%.0f: %v", z.Score, z.Member))
			}
			if length > 10 {
				lines = append(lines, fmt.Sprintf("... (%d more)", length-10))
			}
			info.Value = strings.Join(lines, "\n")
		}
	case "stream":
		length, err := s.redis.XLen(ctx, key).Result()
		if err == nil {
			info.Length = length
		}
		info.Value = fmt.Sprintf("stream with %d entries", info.Length)
	default:
		info.Value = fmt.Sprintf("(%s type)", keyType)
	}

	s.State.KeyValue = info
	s.State.Updates <- s.State
	return nil
}

func (s *Scanner) InfoUpdates() {
	ticker := time.NewTicker(s.Config.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			err := s.FetchRedisInfo()
			if err != nil {
				log.Printf("Error fetching Redis info: %v", err)
				continue
			}
		}
	}
}
