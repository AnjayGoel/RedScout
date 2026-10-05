package scanner

import (
	"log"
	"slices"
	"strings"

	"redscout/models"
)

func (s *Scanner) ComputeNamespaceStats() {
	s.muNamespace.Lock()
	defer s.muNamespace.Unlock()
	s.computeNamespaceStats()
}

func (s *Scanner) computeNamespaceStats() {
	prefix := s.State.CurrentPrefix
	log.Printf("Computing namespace stats for prefix: %s", prefix)

	snapshots := map[string]*models.NamespaceSnapshot{}
	snapshot := func(name string) *models.NamespaceSnapshot {
		sn, ok := snapshots[name]
		if !ok {
			sn = models.NewNamespaceSnapshot(name)
			snapshots[name] = sn
		}
		return sn
	}

	var sampled int64
	s.store.eachSample(func(ks keySample) {
		sampled++
		if name, ok := s.namespaceOf(ks.key, prefix); ok {
			snapshot(name).AddSample(ks.memory, ks.ttl, ks.typ)
		}
	})
	s.store.eachOp(func(kc keyCommand, count int64) {
		if name, ok := s.namespaceOf(kc.key, prefix); ok {
			snapshot(name).OpsFrequency[kc.command] += count
		}
	})

	totalKeys := s.State.Keyspace().Keys
	metrics := make(models.NamespaceMetricList, 0, len(snapshots))
	for _, sn := range snapshots {
		metrics = append(metrics, sn.ToMetric(totalKeys, sampled, s.State.TotalMonitorDuration))
	}
	metrics.Sort(s.State.NamespaceSort)

	s.State.NamespaceStats = metrics
	s.publish()
}

func (s *Scanner) namespaceOf(key string, prefix models.Key) (string, bool) {
	name, err := s.kp.Namespace(s.kp.NewKey(key, false), prefix, true)
	return name, err == nil
}

func (s *Scanner) ComputeBigKeys() {
	topK := int(s.Config.TopK)
	var bigKeys models.BigKeyList
	s.store.eachSample(func(ks keySample) {
		bigKeys = append(bigKeys, models.BigKey{Key: ks.key, Size: ks.memory, Type: ks.typ, TTL: ks.ttl})
		if len(bigKeys) >= 2*topK {
			bigKeys.Sort()
			bigKeys = bigKeys[:topK]
		}
	})
	bigKeys.Sort()

	s.State.BigKeys = bigKeys[:min(len(bigKeys), topK)]
	s.publish()
}

func (s *Scanner) ComputeHotKeys() {
	type keyStats struct {
		total      int64
		topCommand string
		topCount   int64
	}

	stats := map[string]*keyStats{}
	s.store.eachOp(func(kc keyCommand, count int64) {
		st, ok := stats[kc.key]
		if !ok {
			st = &keyStats{}
			stats[kc.key] = st
		}
		st.total += count
		if count > st.topCount {
			st.topCommand, st.topCount = kc.command, count
		}
	})

	seconds := s.State.TotalMonitorDuration.Seconds()
	if seconds == 0 {
		seconds = 1
	}

	hotKeys := make(models.HotKeyList, 0, len(stats))
	for key, st := range stats {
		hotKeys = append(hotKeys, models.HotKey{
			Key:     key,
			Ops:     float64(st.total) / seconds,
			Command: strings.ToUpper(st.topCommand),
		})
	}
	hotKeys.Sort()

	s.State.HotKeys = slices.Clone(hotKeys[:min(len(hotKeys), int(s.Config.TopK))])
	s.publish()
}
