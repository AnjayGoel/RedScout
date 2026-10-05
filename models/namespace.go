package models

import (
	"slices"
	"sort"
	"time"
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

func NewNamespaceSnapshot(namespace string) *NamespaceSnapshot {
	return &NamespaceSnapshot{Namespace: namespace, OpsFrequency: make(map[string]int64)}
}

func (r *NamespaceSnapshot) AddSample(memory, ttl int64, keyType string) {
	r.Keys++
	r.TotalMemory += memory
	if ttl > 0 {
		r.KeysWithTTL++
		r.TotalTTL += ttl
	}
	if !slices.Contains(r.Types, keyType) {
		r.Types = append(r.Types, keyType)
	}
}

func (r *NamespaceSnapshot) ToMetric(totalKeys, sampledKeys int64, monitored time.Duration) *NamespaceMetrics {
	m := &NamespaceMetrics{
		Namespace: r.Namespace,
		Types:     r.Types,
		Ops:       make(map[OpType]float64),
	}

	if r.Keys > 0 && sampledKeys > 0 {
		m.EstKeys = totalKeys * r.Keys / sampledKeys
		m.MemPerKey = float64(r.TotalMemory) / float64(r.Keys)
		m.EstMemory = int64(float64(m.EstKeys) * m.MemPerKey)
		m.TTLPercent = float64(r.KeysWithTTL) / float64(r.Keys)
	}
	if r.KeysWithTTL > 0 {
		m.AvgTTL = r.TotalTTL / r.KeysWithTTL
	}

	if seconds := monitored.Seconds(); seconds > 0 {
		for command, count := range r.OpsFrequency {
			m.Ops[GetOpType(command)] += float64(count) / seconds
		}
	}
	m.Ops[TotalOp] = m.Ops[GetOp] + m.Ops[SetOp] + m.Ops[DelOp] + m.Ops[EvalOp]

	return m
}

type NamespaceMetricList []*NamespaceMetrics

func (d NamespaceMetricList) Sort(sortBy string) {
	sort.Slice(d, func(i, j int) bool {
		switch sortBy {
		case "Keys":
			return d[i].EstKeys > d[j].EstKeys
		case "TTL":
			return d[i].AvgTTL > d[j].AvgTTL
		case "% TTL":
			return d[i].TTLPercent > d[j].TTLPercent
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
