package models

import (
	"sort"
)

type HotKey struct {
	Key Key
	Ops float64
}

type HotKeyList []HotKey

func (h HotKeyList) Sort() {
	sort.Slice(h, func(i, j int) bool {
		return h[i].Ops > h[j].Ops
	})
}

type BigKey struct {
	Key  Key
	Size int64
}

type BigKeyList []BigKey

func (b BigKeyList) Sort() {
	sort.Slice(b, func(i, j int) bool {
		return b[i].Size > b[j].Size
	})
}

// BigKeyMinHeap is a min-heap for BigKey by Size (for top-K selection)
type BigKeyMinHeap []BigKey

func (h *BigKeyMinHeap) Len() int           { return len(*h) }
func (h *BigKeyMinHeap) Less(i, j int) bool { return (*h)[i].Size < (*h)[j].Size }
func (h *BigKeyMinHeap) Swap(i, j int)      { (*h)[i], (*h)[j] = (*h)[j], (*h)[i] }
func (h *BigKeyMinHeap) Push(x interface{}) {
	*h = append(*h, x.(BigKey))
}
func (h *BigKeyMinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// HotKeyMinHeap is a min-heap for HotKey by Ops (for top-K selection)
type HotKeyMinHeap []HotKey

func (h *HotKeyMinHeap) Len() int           { return len(*h) }
func (h *HotKeyMinHeap) Less(i, j int) bool { return (*h)[i].Ops < (*h)[j].Ops }
func (h *HotKeyMinHeap) Swap(i, j int)      { (*h)[i], (*h)[j] = (*h)[j], (*h)[i] }
func (h *HotKeyMinHeap) Push(x interface{}) {
	*h = append(*h, x.(HotKey))
}
func (h *HotKeyMinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
