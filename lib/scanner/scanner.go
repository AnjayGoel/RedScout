package scanner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/Masterminds/semver/v3"
	"github.com/redis/go-redis/v9"

	"redscout/lib"
	"redscout/models"
)

var ErrBusy = errors.New("operation already in progress")

var minRedisVersion = semver.MustParse("4.0.0")

type Scanner struct {
	Config *models.Config
	State  *models.State

	kp      *models.KeyParser
	redis   *redis.Client
	store   *store
	ctx     context.Context
	cancel  context.CancelFunc
	logFile *os.File

	muInfo       sync.Mutex
	muScanRun    sync.Mutex
	muMonitorRun sync.Mutex
	muNamespace  sync.Mutex
}

func NewScanner(cfg *models.Config) (*Scanner, error) {
	client, err := lib.RedisClientFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	logFile, err := os.CreateTemp(cfg.LogsDir, "redscout_log_")
	if err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}
	log.SetOutput(logFile)

	ctx, cancel := context.WithCancel(context.Background())
	return &Scanner{
		Config:  cfg,
		State:   models.NewState(cfg.RedisDB),
		kp:      models.NewKeyParser(cfg.Delimiter, cfg.IDPatterns),
		redis:   client,
		store:   newStore(),
		ctx:     ctx,
		cancel:  cancel,
		logFile: logFile,
	}, nil
}

func (s *Scanner) Close() {
	s.cancel()
	_ = s.redis.Close()
	log.Printf("Scanner closed")
	log.SetOutput(io.Discard)
	_ = s.logFile.Close()
}

func (s *Scanner) publish() {
	select {
	case s.State.Updates <- s.State:
	default:
	}
}

func (s *Scanner) updateStatus(status string) {
	s.State.Status = status
	s.publish()
}

func (s *Scanner) Start() {
	s.updateStatus("Fetching redis info")
	if err := s.FetchRedisInfo(); err != nil {
		s.updateStatus(fmt.Sprintf("Error fetching Redis info: %v", err))
		return
	}
	version := s.State.RedisInfo.Server.RedisVersion
	if v, err := semver.NewVersion(version); err == nil && v.LessThan(minRedisVersion) {
		s.updateStatus(fmt.Sprintf("Unsupported Redis version %s, need at least %s", version, minRedisVersion))
		return
	}

	//Redis stats info
	go s.InfoUpdates()

	//Scan to analyse memory usage & keys
	if err := s.ScanMemory(); err != nil {
		s.updateStatus(fmt.Sprintf("Error scanning memory: %v", err))
		return
	}

	// Start Monitor to analyze operations
	if err := s.MonitorOps(); err != nil {
		s.updateStatus(fmt.Sprintf("Error monitoring operations: %v", err))
		return
	}

	s.updateStatus("Computing statistics")

	if err := s.FetchSlowLog(); err != nil {
		s.updateStatus(fmt.Sprintf("Error fetching slow log: %v", err))
	}
	s.ComputeNamespaceStats()
	s.ComputeBigKeys()
	s.ComputeHotKeys()

	s.State.ScanComplete = true
	s.updateStatus("Initial data load complete")
}

func (s *Scanner) Rescan() {
	if err := s.ScanMemory(); err != nil {
		if !errors.Is(err, ErrBusy) {
			s.updateStatus(fmt.Sprintf("Error scanning memory: %v", err))
		}
		return
	}
	s.ComputeNamespaceStats()
	s.ComputeBigKeys()
}

func (s *Scanner) Remonitor() {
	if err := s.MonitorOps(); err != nil {
		if !errors.Is(err, ErrBusy) {
			s.updateStatus(fmt.Sprintf("Error monitoring operations: %v", err))
		}
		return
	}
	s.ComputeNamespaceStats()
	s.ComputeHotKeys()
}

func (s *Scanner) DrillDownNamespace(namespace string) {
	s.muNamespace.Lock()
	defer s.muNamespace.Unlock()

	newPrefix, err := s.kp.Append(s.State.CurrentPrefix, namespace, true)
	if err != nil {
		return
	}
	log.Printf("Drilling down into namespace %s, prefix %s", namespace, newPrefix)

	s.State.CurrentPrefix = newPrefix
	s.computeNamespaceStats()

	// If no sub-namespaces, find a matching key and fetch its value
	if len(s.State.NamespaceStats) == 0 {
		if key, ok := s.store.firstKeyWithPrefix(s.kp, newPrefix); ok {
			go s.FetchKeyValue(key)
		}
	}
}

func (s *Scanner) LevelUpNamespace() {
	s.muNamespace.Lock()
	defer s.muNamespace.Unlock()

	if s.State.CurrentPrefix.IsEmpty() {
		return
	}
	newPrefix, err := s.State.CurrentPrefix.Pop()
	if err != nil {
		return
	}
	log.Printf("Going up from prefix %s to %s", s.State.CurrentPrefix, newPrefix)

	s.State.KeyValue = nil
	s.State.CurrentPrefix = newPrefix
	s.computeNamespaceStats()
}
