package models

import (
	"time"
)

type State struct {
	//Current execution state variables
	CurrentPrefix Key

	LastInfoCheck        time.Time
	TotalMonitorDuration time.Duration
	Cursor               uint64
	ScannedKeys          int64

	// Redis Info
	RedisInfo *RedisInfo

	//Current Prefix and its analysis
	NamespaceStats NamespaceMetricList

	// Special Keys for Debugging and Analysis
	SlowLogs SlowLogList
	HotKeys  HotKeyList
	BigKeys  BigKeyList

	//Chan to send updates
	Updates chan *State

	//Last Status Message
	Status       string
	ScanComplete bool
}

func NewState() *State {
	return &State{
		CurrentPrefix:        Key{},
		LastInfoCheck:        time.Unix(0, 0),
		TotalMonitorDuration: 0,
		Cursor:               0,
		ScannedKeys:          0,
		RedisInfo:            &RedisInfo{},
		NamespaceStats:       NamespaceMetricList{},
		SlowLogs:             SlowLogList{},
		HotKeys:              HotKeyList{},
		BigKeys:              BigKeyList{},
		Updates:              make(chan *State, 100), // Buffered channel for updates
		Status:               "Initializing",
		ScanComplete:         false,
	}
}
