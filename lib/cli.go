package lib

import (
	"flag"
	"redmon/models"
	"regexp"
	"strings"
	"time"
)

// ParseFlags parses command line flags and returns a config
func ParseFlags() models.Config {
	config := models.DefaultConfig()

	flag.StringVar(&config.RedisHost, "host", config.RedisHost, "Redis host address")
	flag.IntVar(&config.RedisPort, "port", config.RedisPort, "Redis port number")
	flag.StringVar(&config.RedisPassword, "password", config.RedisPassword, "Redis password")
	flag.IntVar(&config.RedisDB, "db", config.RedisDB, "Redis database number")

	var useTLS string
	flag.StringVar(&useTLS, "tls", "", "Use TLS for Redis connection")

	flag.Int64Var(&config.KeysScanSize, "scan-size", config.KeysScanSize, "Number of keys to scan per iteration")

	var monitorDuration int
	flag.IntVar(&monitorDuration, "monitor-duration", int(config.MonitorDuration.Seconds()), "Duration in seconds to monitor Redis operations")

	var refreshInterval int
	flag.IntVar(&refreshInterval, "refresh-interval", int(config.RefreshInterval.Seconds()), "Interval in seconds between Redis info refreshes")

	flag.StringVar(&config.Delimiter, "delimiter", config.Delimiter, "Delimiter for separating redis keys")

	flag.StringVar(&config.LogsDir, "logs-dir", config.LogsDir, "Directory to store logs")

	idRegexInput := ""
	flag.StringVar(&idRegexInput, "id-regex", "", "space seperated list of regex to infer IDs from keys")

	flag.Parse()

	if useTLS != "" {
		config.UseTLS = strings.ToLower(useTLS) == "true"
	}

	if monitorDuration > 0 {
		config.MonitorDuration = time.Duration(monitorDuration) * time.Second
	}

	if refreshInterval > 0 {
		config.RefreshInterval = time.Duration(refreshInterval) * time.Second
	}

	for _, pattern := range strings.Split(idRegexInput, " ") {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		regex, err := regexp.Compile("^" + pattern + "$")
		if err != nil {
			panic("Invalid regex pattern: " + pattern)
		}

		config.IDPatterns = append(config.IDPatterns, regex)
	}

	return config
}
