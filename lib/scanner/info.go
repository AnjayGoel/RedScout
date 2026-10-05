package scanner

import (
	"log"
	"time"

	"redscout/models"
)

func (s *Scanner) FetchSlowLog() error {
	slowLog, err := s.redis.SlowLogGet(s.ctx, s.Config.TopK).Result()
	if err != nil {
		return err
	}

	models.SlowLogList(slowLog).Sort("Timestamp")
	s.State.SlowLogs = slowLog
	s.publish()
	return nil
}

func (s *Scanner) FetchRedisInfo() error {
	s.muInfo.Lock()
	defer s.muInfo.Unlock()

	infoStr, err := s.redis.Info(s.ctx).Result()
	if err != nil {
		return err
	}

	parsed := models.ParseInfo(infoStr)
	prev := s.State.RedisInfo
	now := time.Now()

	dHits := parsed.Stats.KeyspaceHits - prev.Stats.KeyspaceHits
	dMisses := parsed.Stats.KeyspaceMisses - prev.Stats.KeyspaceMisses
	if dHits+dMisses > 0 {
		parsed.Computed.HitRate = float64(dHits) / float64(dHits+dMisses)
	} else {
		parsed.Computed.HitRate = prev.Computed.HitRate
	}

	if !s.State.LastInfoCheck.IsZero() {
		cpuDelta := (parsed.CPU.SystemTime + parsed.CPU.UserTime) - (prev.CPU.SystemTime + prev.CPU.UserTime)
		parsed.Computed.CPUUsage = cpuDelta / now.Sub(s.State.LastInfoCheck).Seconds()
	}

	s.State.LastInfoCheck = now
	s.State.RedisInfo = &parsed
	s.publish()
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
			if err := s.FetchRedisInfo(); err != nil {
				log.Printf("Error fetching Redis info: %v", err)
			}
		}
	}
}
