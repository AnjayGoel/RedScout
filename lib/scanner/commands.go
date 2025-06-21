package scanner

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"io"
	"log"
	"redmon/lib"
	"redmon/models"
	"strings"
	"time"
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

	if s.State.LastInfoCheck.Second() != 0 {
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
		if next == 0 || scanned >= s.Config.KeysScanSize {
			break
		}
		s.State.Cursor = next
	}

	s.State.ScannedKeys += scanned
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

	keys, err := s.scanKeys()
	if err != nil {
		return err
	}

	if _, err := s.scanFile.Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("failed to seek scan file: %w", err)
	}

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

		if _, err := pipe.Exec(s.ctx); err != nil {
			return err
		}

		for _, tr := range trips {
			xMem, e1 := tr.mem.Result()
			xTtl, e2 := tr.ttl.Result()
			xType, e3 := tr.typeCmd.Result()
			if e1 != nil || e2 != nil || e3 != nil {
				continue
			}
			_, _ = s.scanFile.WriteString(fmt.Sprintf("%s %d %d %s\n", tr.key, xMem, int64(xTtl.Seconds()), xType))
		}
	}

	log.Printf("Memory scan completed; scanned %d keys", len(keys))
	s.updateStatus("Memory scan completed")
	return nil
}

func (s *Scanner) MonitorOps() error {
	s.muRedis.Lock()
	defer s.muRedis.Unlock()
	s.updateStatus("Monitoring operations")

	log.Printf("Ops monitor started for %v", s.Config.MonitorDuration)

	//Buffered channel to handle Redis monitor output
	ch := make(chan string, 10000)
	defer close(ch)

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
	defer monitor.Stop()

	if _, err := s.monitorFile.Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("failed to seek monitor file: %w", err)
	}

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
			var key string
			if len(parts) >= 4 {
				key = parts[3]
			}
			if cmd == "eval" {
				continue
			}
			_, _ = s.monitorFile.WriteString(fmt.Sprintf("%s %s\n", key, cmd))
		case <-ctxTimeout.Done():
			s.State.TotalMonitorDuration += s.Config.MonitorDuration
			s.updateStatus("Monitoring completed")
			log.Printf("Monitoring completed")
			return nil
		}
	}
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
