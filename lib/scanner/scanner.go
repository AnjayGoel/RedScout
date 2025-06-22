package scanner

import (
	"context"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"redscout/lib"
	"redscout/models"
	"sync"
)

type Scanner struct {
	Config *models.Config
	kp     *models.KeyParser

	ctx     context.Context
	cancel  context.CancelFunc
	logFile *os.File

	State *models.State

	redis   *redis.Client
	muRedis sync.Mutex

	monitorFile *os.File
	muMonitor   sync.Mutex

	scanFile *os.File
	muScan   sync.Mutex
}

func NewScanner(cfg *models.Config) (*Scanner, error) {
	ctx, cancel := context.WithCancel(context.Background())

	client, err := lib.RedisClientFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	logFile, err := os.CreateTemp(cfg.LogsDir, "redscout_log_")
	if err != nil {
		return nil, fmt.Errorf("failed to create logFile file: %w", err)
	}
	log.SetOutput(logFile)

	monitorFile, err := os.CreateTemp(cfg.LogsDir, "redscout_monitor_")
	if err != nil {
		return nil, err
	}

	scanFile, err := os.CreateTemp(cfg.LogsDir, "redscout_scan_")

	if err != nil {
		return nil, err
	}

	s := &Scanner{
		Config: cfg,
		kp:     models.NewKeyParser(cfg.Delimiter, cfg.IDPatterns),

		ctx:     ctx,
		cancel:  cancel,
		logFile: logFile,

		redis:   client,
		muRedis: sync.Mutex{},

		State: models.NewState(),

		monitorFile: monitorFile,
		muMonitor:   sync.Mutex{},

		scanFile: scanFile,
		muScan:   sync.Mutex{},
	}

	return s, nil
}

func (s *Scanner) Close() {
	if s.redis != nil {
		_ = s.redis.Close()
	}
	s.cancel()
	log.Printf("Scanner closed")
}

func (s *Scanner) updateStatus(status string) {
	s.State.Status = status
	s.State.Updates <- s.State
}

func (s *Scanner) Start() {
	s.updateStatus("Fetching redis info")
	err := s.FetchRedisInfo()
	if err != nil {
		s.updateStatus(fmt.Sprintf("Error fetching Redis info: %v", err))
		return
	}
	ver := semver.MustParse(s.State.RedisInfo.Server.RedisVersion)
	if ver.LessThan(semver.MustParse("4.0.0")) {
		log.Fatalf("unsupported Redis version: %s, must be at least v4.0.0", s.State.RedisInfo.Server.RedisVersion)
	}

	//Redis stats info
	go s.InfoUpdates()

	//Scan to analyse memory usage & keys
	err = s.ScanMemory()
	if err != nil {
		s.updateStatus(fmt.Sprintf("Error scanning memory: %v", err))
		return
	}

	// Start Monitor to analyze operations
	err = s.MonitorOps()
	if err != nil {
		s.updateStatus(fmt.Sprintf("Error monitoring operations: %v", err))
		return
	}

	s.updateStatus("Computing statistics")

	err = s.FetchSlowLog()
	if err != nil {
		s.updateStatus(fmt.Sprintf("Error fetching slow logFile: %v", err))
	}

	err = s.ComputeNamespaceStats()
	if err != nil {
		s.updateStatus(fmt.Sprintf("Error generating namespace stats: %v", err))
	}

	err = s.ComputeBigKeysFromScanLog()
	if err != nil {
		s.updateStatus(fmt.Sprintf("Error computing big keys from scan log: %v", err))
		return
	}
	err = s.ComputeHotKeysFromMonitorLog()
	if err != nil {
		s.updateStatus(fmt.Sprintf("Error computing keys from monitor log: %v", err))
		return
	}

	s.State.ScanComplete = true
	s.updateStatus("Initial data load complete")
}

func (s *Scanner) DrillDownNamespace(namespace string) {
	currentPrefix := s.State.CurrentPrefix

	newPrefix, err := s.kp.Append(currentPrefix, namespace, true)
	if err != nil {
		return
	}

	log.Printf("Drilling down into namespace: %s with new prefix: %s\n", namespace, newPrefix)

	s.State.CurrentPrefix = newPrefix
	_ = s.ComputeNamespaceStats()
}

func (s *Scanner) LevelUpNamespace() {
	currentPrefix := s.State.CurrentPrefix
	if currentPrefix.IsEmpty() {
		return
	}

	newPrefix, err := currentPrefix.Pop()
	if err != nil {
		return
	}

	log.Printf("Going one level up from prefix: %s to new prefix: %s\n", currentPrefix, newPrefix)

	s.State.CurrentPrefix = newPrefix
	_ = s.ComputeNamespaceStats()
}
