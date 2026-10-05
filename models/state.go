package models

import (
	"fmt"
	"time"
)

type State struct {
	//Current execution state variables
	DB            int
	CurrentPrefix Key

	LastInfoCheck        time.Time
	TotalMonitorDuration time.Duration
	Cursor               uint64
	ScannedKeys          int64

	// Redis Info
	RedisInfo *RedisInfo

	//Current Prefix and its analysis
	NamespaceStats NamespaceMetricList
	NamespaceSort  string

	// Special Keys for Debugging and Analysis
	SlowLogs SlowLogList
	HotKeys  HotKeyList
	BigKeys  BigKeyList

	// Key Value Viewer
	KeyValue *KeyValueInfo

	//Chan to send updates
	Updates chan *State

	//Last Status Message
	Status       string
	ScanComplete bool

	// Tracking progress of operations
	ScanProgress    float64 // 0-100
	MonitorProgress float64 // 0-100
}

func NewState(db int) *State {
	return &State{
		DB:            db,
		RedisInfo:     &RedisInfo{},
		NamespaceSort: "Keys",
		Updates:       make(chan *State, 1), // Buffered channel for updates
		Status:        "Initializing",
	}
}

func (s *State) Keyspace() KeyspaceInfo {
	return s.RedisInfo.Keyspace[fmt.Sprintf("db%d", s.DB)]
}
