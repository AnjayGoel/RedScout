package models

import (
	"github.com/redis/go-redis/v9"
	"sort"
)

type SlowLogList []redis.SlowLog

func (d SlowLogList) Sort(key string) {
	sort.Slice(d, func(i, j int) bool {
		switch key {
		case "ID":
			return d[i].ID > d[j].ID
		case "Timestamp":
			return d[i].Time.After(d[j].Time)
		case "Duration":
			return d[i].Duration > d[j].Duration
		case "Command":
			return d[i].Args[0] > d[j].Args[0]
		default:
			return d[i].ID > d[j].ID
		}
	})
}
