package scanner

import (
	"log"

	"github.com/redis/go-redis/v9"

	"redscout/lib"
)

type keyProbe struct {
	key    string
	memory *redis.IntCmd
	ttl    *redis.DurationCmd
	typ    *redis.StatusCmd
}

func (s *Scanner) ScanMemory() error {
	if !s.muScanRun.TryLock() {
		return ErrBusy
	}
	defer s.muScanRun.Unlock()

	log.Printf("Memory scan started")

	if s.State.Cursor == 0 && s.State.ScannedKeys > 0 {
		s.store.resetSamples()
		s.State.ScannedKeys = 0
	}

	s.updateStatus("Collecting keys")

	keys, err := s.scanKeys()
	if err != nil {
		return err
	}

	s.updateStatus("Scanning memory")

	for i := 0; i < len(keys); i += lib.MemoryPipeBatchSize {
		batch := keys[i:min(i+lib.MemoryPipeBatchSize, len(keys))]
		samples, err := s.probeKeys(batch)
		if err != nil {
			return err
		}
		s.store.addSamples(samples)

		processed := i + len(batch)
		s.State.ScanProgress = 50 + min(float64(processed)/float64(len(keys))*50, 49)
		s.publish()
	}

	s.State.ScanProgress = 100
	log.Printf("Memory scan completed; scanned %d keys", len(keys))
	s.updateStatus("Memory scan completed")
	return nil
}

func (s *Scanner) scanKeys() ([]string, error) {
	var keys []string
	for {
		batch, next, err := s.redis.Scan(s.ctx, s.State.Cursor, "*", lib.ScanBatchSize).Result()
		if err != nil {
			log.Printf("scan error: %v", err)
			return nil, err
		}

		keys = append(keys, batch...)
		s.State.Cursor = next
		s.State.ScannedKeys += int64(len(batch))
		s.State.ScanProgress = min(float64(len(keys))/float64(s.Config.KeysScanSize)*50, 50)
		s.publish()

		if next == 0 || int64(len(keys)) >= s.Config.KeysScanSize {
			return keys, nil
		}
	}
}

func (s *Scanner) probeKeys(keys []string) ([]keySample, error) {
	pipe := s.redis.Pipeline()
	probes := make([]keyProbe, len(keys))
	for i, key := range keys {
		probes[i] = keyProbe{
			key:    key,
			memory: pipe.MemoryUsage(s.ctx, key),
			ttl:    pipe.TTL(s.ctx, key),
			typ:    pipe.Type(s.ctx, key),
		}
	}
	if _, err := pipe.Exec(s.ctx); err != nil && s.ctx.Err() != nil {
		return nil, err
	}

	samples := make([]keySample, 0, len(probes))
	for _, p := range probes {
		memory, err1 := p.memory.Result()
		ttl, err2 := p.ttl.Result()
		typ, err3 := p.typ.Result()
		if err1 != nil || err2 != nil || err3 != nil {
			continue
		}
		samples = append(samples, keySample{key: p.key, memory: memory, ttl: int64(ttl.Seconds()), typ: typ})
	}
	return samples, nil
}
