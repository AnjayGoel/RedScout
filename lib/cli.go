package lib

import (
	"flag"
	"redscout/models"
	"regexp"
	"strings"
	"time"
)

// ParseFlags parses command line flags and returns a config
func ParseFlags() models.Config {
	config := models.DefaultConfig()

	// Redis connection flags (standard redis-cli shorthands)
	flag.StringVar(&config.RedisHost, "h", config.RedisHost, "Redis host address")
	flag.IntVar(&config.RedisPort, "p", config.RedisPort, "Redis port number")
	flag.StringVar(&config.RedisUser, "u", config.RedisUser, "Redis username")
	flag.StringVar(&config.RedisPassword, "a", config.RedisPassword, "Redis password")
	flag.IntVar(&config.RedisDB, "n", config.RedisDB, "Redis database number")

	flag.BoolVar(&config.UseTLS, "tls", config.UseTLS, "Use TLS for Redis connection")

	// Application-specific flags (long form only)
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
