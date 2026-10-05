package models

import (
	"sort"
)

type HotKey struct {
	Key     string
	Ops     float64
	Command string
}

type HotKeyList []HotKey

func (h HotKeyList) Sort() {
	sort.Slice(h, func(i, j int) bool {
		return h[i].Ops > h[j].Ops
	})
}

type BigKey struct {
	Key  string
	Size int64
	Type string
	TTL  int64
}

type BigKeyList []BigKey

func (b BigKeyList) Sort() {
	sort.Slice(b, func(i, j int) bool {
		return b[i].Size > b[j].Size
	})
}
