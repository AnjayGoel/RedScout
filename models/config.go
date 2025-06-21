package models

import (
	"os"
	"regexp"
	"strconv"
	"time"
)

type Config struct {
	RedisHost       string
	RedisPort       int
	RedisPassword   string
	RedisDB         int
	UseTLS          bool
	KeysScanSize    int64
	MonitorDuration time.Duration
	RefreshInterval time.Duration
	Delimiter       string
	LogsDir         string
	TopK            int64
	IDPatterns      []*regexp.Regexp
}

func DefaultConfig() Config {
	host := "localhost"
	if os.Getenv("REDIS_HOST") != "" {
		host = os.Getenv("REDIS_HOST")
	}

	port := 6379
	if os.Getenv("REDIS_PORT") != "" {
		var err error
		port, err = strconv.Atoi(os.Getenv("REDIS_PORT"))
		if err != nil {
			port = 6379 // default port
		}
	}

	password := os.Getenv("REDIS_PASSWORD")
	if password == "" {
		password = ""
	}

	return Config{
		RedisHost:       host,
		RedisPort:       port,
		RedisPassword:   password,
		RedisDB:         0,
		UseTLS:          true,
		KeysScanSize:    5000,
		MonitorDuration: 10 * time.Second,
		RefreshInterval: 5 * time.Second,
		Delimiter:       ":",
		LogsDir:         os.TempDir(),
		TopK:            100,
		IDPatterns:      []*regexp.Regexp{},
	}
}
