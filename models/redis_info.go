package models

import (
	"strconv"
	"strings"
)

type RedisInfo struct {
	Server   ServerInfo
	Clients  ClientsInfo
	Memory   MemoryInfo
	Stats    StatsInfo
	CPU      CPUInfo
	Keyspace map[string]KeyspaceInfo
	Computed ComputedStats
}

func NewRedisInfo() RedisInfo {
	return RedisInfo{
		Keyspace: make(map[string]KeyspaceInfo),
	}
}

type ServerInfo struct {
	RedisVersion string
	OS           string
	ArchBits     int
	Uptime       int64
	HZ           int
}

type ClientsInfo struct {
	ConnectedClients int
	BlockedClients   int
}

type MemoryInfo struct {
	UsedMemory         int64
	UsedMemoryHuman    string
	MaxMemory          int64
	MaxMemoryHuman     string
	MemoryPolicy       string
	UsedMemoryPeakPerc float64
}

type CPUInfo struct {
	UserTime   float64
	SystemTime float64
}

type StatsInfo struct {
	TotalConnections int
	OpsPerSec        int64
	KeyspaceHits     int64
	KeyspaceMisses   int64
}

type KeyspaceInfo struct {
	Keys    int64
	Expires int64
	AvgTTL  int64
}

type ComputedStats struct {
	CPUUsage float64
	HitRate  float64
}

func ParseInfo(info string) RedisInfo {
	lines := strings.Split(info, "\n")
	result := NewRedisInfo()

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := parts[0], parts[1]

		switch key {
		// Server
		case "redis_version":
			result.Server.RedisVersion = val
		case "os":
			result.Server.OS = val
		case "arch_bits":
			result.Server.ArchBits, _ = strconv.Atoi(val)
		case "uptime_in_seconds":
			result.Server.Uptime, _ = strconv.ParseInt(val, 10, 64)
		case "hz":
			result.Server.HZ, _ = strconv.Atoi(val)

		// Clients
		case "connected_clients":
			result.Clients.ConnectedClients, _ = strconv.Atoi(val)
		case "blocked_clients":
			result.Clients.BlockedClients, _ = strconv.Atoi(val)

		// CPU
		case "used_cpu_sys":
			result.CPU.SystemTime, _ = strconv.ParseFloat(val, 64)
		case "used_cpu_user":
			result.CPU.UserTime, _ = strconv.ParseFloat(val, 64)

		// Memory
		case "used_memory":
			result.Memory.UsedMemory, _ = strconv.ParseInt(val, 10, 64)
		case "used_memory_human":
			result.Memory.UsedMemoryHuman = val
		case "maxmemory":
			result.Memory.MaxMemory, _ = strconv.ParseInt(val, 10, 64)
		case "maxmemory_human":
			result.Memory.MaxMemoryHuman = val
		case "used_memory_peak_perc":
			result.Memory.UsedMemoryPeakPerc, _ = strconv.ParseFloat(strings.TrimSuffix(val, "%"), 64)
		case "maxmemory_policy":
			result.Memory.MemoryPolicy = val

		// RawNamespaceStats
		case "total_connections_received":
			result.Stats.TotalConnections, _ = strconv.Atoi(val)
		case "instantaneous_ops_per_sec":
			result.Stats.OpsPerSec, _ = strconv.ParseInt(val, 10, 64)
		case "keyspace_hits":
			result.Stats.KeyspaceHits, _ = strconv.ParseInt(val, 10, 64)
		case "keyspace_misses":
			result.Stats.KeyspaceMisses, _ = strconv.ParseInt(val, 10, 64)

		// Keyspace
		default:
			if strings.HasPrefix(key, "db") {
				dbParts := strings.Split(val, ",")
				ks := KeyspaceInfo{}
				for _, p := range dbParts {
					kv := strings.SplitN(p, "=", 2)
					if len(kv) != 2 {
						continue
					}
					switch kv[0] {
					case "keys":
						ks.Keys, _ = strconv.ParseInt(kv[1], 10, 64)
					case "expires":
						ks.Expires, _ = strconv.ParseInt(kv[1], 10, 64)
					case "avg_ttl":
						ks.AvgTTL, _ = strconv.ParseInt(kv[1], 10, 64)
						ks.AvgTTL /= 1000
					}
				}
				result.Keyspace[key] = ks
			}
		}
	}

	queries := result.Stats.KeyspaceHits + result.Stats.KeyspaceMisses
	if queries == 0 {
		result.Computed.HitRate = 1.0
	} else {
		result.Computed.HitRate = float64(result.Stats.KeyspaceHits) / float64(queries)
	}

	cpuTime := result.CPU.UserTime + result.CPU.SystemTime
	if cpuTime > 0 && result.Server.Uptime > 0 {
		result.Computed.CPUUsage = cpuTime / float64(result.Server.Uptime)
	} else {
		result.Computed.CPUUsage = 0.0
	}

	return result
}
