package scanner

import (
	"context"
	"log"
	"time"

	"redscout/lib"
	"redscout/models"
)

const monitorBufferSize = 10000

func (s *Scanner) MonitorOps() error {
	if !s.muMonitorRun.TryLock() {
		return ErrBusy
	}
	defer s.muMonitorRun.Unlock()

	s.updateStatus("Monitoring operations")
	log.Printf("Ops monitor started for %v", s.Config.MonitorDuration)

	start := time.Now()
	s.State.MonitorProgress = 0

	client, err := lib.RedisClientFromConfig(s.Config)
	if err != nil {
		return err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(s.ctx, s.Config.MonitorDuration)
	defer cancel()

	//Buffered channel to handle Redis monitor output
	lines := make(chan string, monitorBufferSize)
	monitor := client.Monitor(ctx, lines)
	monitor.Start()
	defer func() {
		monitor.Stop()
		go drainUntilIdle(lines)
	}()

	progress := time.NewTicker(100 * time.Millisecond)
	defer progress.Stop()

	pending := opCounts{}
	flush := func() {
		s.store.addOps(pending)
		pending = opCounts{}
	}

	for {
		select {
		case line := <-lines:
			entry, ok := models.ParseMonitorLine(line)
			if !ok || entry.DB != s.Config.RedisDB {
				continue
			}
			for _, key := range entry.Keys {
				pending[keyCommand{key: key, command: entry.Command}]++
			}
		case <-progress.C:
			flush()
			elapsed := time.Since(start)
			s.State.MonitorProgress = min(float64(elapsed)/float64(s.Config.MonitorDuration)*100, 100)
			s.publish()
		case <-ctx.Done():
			flush()
			s.State.MonitorProgress = 100
			s.State.TotalMonitorDuration += s.Config.MonitorDuration
			s.updateStatus("Monitoring completed")
			log.Printf("Monitoring completed")
			return nil
		}
	}
}

func drainUntilIdle(lines <-chan string) {
	for {
		select {
		case <-lines:
		case <-time.After(time.Second):
			return
		}
	}
}
