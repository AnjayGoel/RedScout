package models

import (
	"sort"
)

type NamespaceSnapshot struct {
	Namespace    string
	Keys         int64
	KeysWithTTL  int64
	TotalMemory  int64
	TotalTTL     int64
	OpsFrequency map[string]int64
	Types        []string
}

type NamespaceMetrics struct {
	Namespace  string
	EstKeys    int64
	EstMemory  int64
	TTLPercent float64
	AvgTTL     int64
	MemPerKey  float64
	Ops        map[OpType]float64
	Types      []string
}

func (r *NamespaceSnapshot) ToMetric(s *State) *NamespaceMetrics {
	if s.ScannedKeys == 0 || r.Keys == 0 {
		return &NamespaceMetrics{
			Namespace: r.Namespace,
			Types:     r.Types,
			Ops:       make(map[OpType]float64),
		}
	}

	processed := &NamespaceMetrics{}

	totalKeys := s.RedisInfo.Keyspace["db0"].Keys

	processed.Namespace = r.Namespace
	processed.EstKeys = (totalKeys * r.Keys) / s.ScannedKeys
	processed.MemPerKey = float64(r.TotalMemory) / float64(r.Keys)
	processed.EstMemory = int64(float64(processed.EstKeys) * processed.MemPerKey)
	processed.TTLPercent = float64(r.KeysWithTTL) / float64(r.Keys)

	if r.KeysWithTTL == 0 {
		processed.AvgTTL = 0
	} else {
		processed.AvgTTL = r.TotalTTL / r.KeysWithTTL
	}

	processed.Ops = make(map[OpType]float64)
	for op, count := range r.OpsFrequency {
		processed.Ops[GetOpType(op)] += float64(count)
	}

	if s.TotalMonitorDuration > 0 {
		for opType, count := range processed.Ops {
			processed.Ops[opType] = count / s.TotalMonitorDuration.Seconds()
			processed.Ops[TotalOp] += processed.Ops[opType]
		}
	}

	processed.Types = r.Types
	return processed
}

type NamespaceMetricList []*NamespaceMetrics

func (d NamespaceMetricList) Sort(sortBy string) {
	sort.Slice(d, func(i, j int) bool {
		switch sortBy {
		case "Keys":
			return d[i].EstKeys > d[j].EstKeys
		case "TTL":
			return d[i].AvgTTL > d[j].AvgTTL
		case "Get":
			return d[i].Ops[GetOp] > d[j].Ops[GetOp]
		case "Set":
			return d[i].Ops[SetOp] > d[j].Ops[SetOp]
		case "Del":
			return d[i].Ops[DelOp] > d[j].Ops[DelOp]
		case "Total Ops":
			return d[i].Ops[TotalOp] > d[j].Ops[TotalOp]
		default:
			return d[i].EstMemory > d[j].EstMemory
		}
	})
}
