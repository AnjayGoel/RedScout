package models

import (
	"strconv"
	"strings"
)

type MonitorEntry struct {
	DB      int
	Command string
	Keys    []string
}

var keylessCommands = map[string]bool{
	"auth": true, "client": true, "command": true, "config": true, "dbsize": true,
	"echo": true, "flushall": true, "flushdb": true, "hello": true, "info": true,
	"keys": true, "ping": true, "publish": true, "scan": true, "script": true,
	"select": true, "slowlog": true, "subscribe": true, "time": true,
}

func ParseMonitorLine(line string) (MonitorEntry, bool) {
	open := strings.IndexByte(line, '[')
	if open < 0 {
		return MonitorEntry{}, false
	}
	rest := line[open+1:]
	space := strings.IndexByte(rest, ' ')
	if space < 0 {
		return MonitorEntry{}, false
	}
	db, err := strconv.Atoi(rest[:space])
	if err != nil {
		return MonitorEntry{}, false
	}

	args := quotedArgs(rest[space:])
	if len(args) == 0 {
		return MonitorEntry{}, false
	}
	cmd := strings.ToLower(args[0])
	return MonitorEntry{DB: db, Command: cmd, Keys: commandKeys(cmd, args[1:])}, true
}

func quotedArgs(s string) []string {
	var args []string
	for {
		start := strings.IndexByte(s, '"')
		if start < 0 {
			return args
		}
		end := start + 1
		for end < len(s) && s[end] != '"' {
			if s[end] == '\\' {
				end++
			}
			end++
		}
		if end >= len(s) {
			return args
		}
		arg, err := strconv.Unquote(s[start : end+1])
		if err != nil {
			arg = s[start+1 : end]
		}
		args = append(args, arg)
		s = s[end+1:]
	}
}

func commandKeys(cmd string, args []string) []string {
	if keylessCommands[cmd] || len(args) == 0 {
		return nil
	}
	switch cmd {
	case "mset", "msetnx":
		// MSET key1 val1 key2 val2 ... — every other arg is a key
		keys := make([]string, 0, (len(args)+1)/2)
		for i := 0; i < len(args); i += 2 {
			keys = append(keys, args[i])
		}
		return keys
	case "mget", "del", "unlink", "exists", "touch", "watch":
		// All args are keys
		return args
	case "eval", "evalsha", "eval_ro", "evalsha_ro", "fcall", "fcall_ro":
		if len(args) < 2 {
			return nil
		}
		n, err := strconv.Atoi(args[1])
		if err != nil || n < 0 || 2+n > len(args) {
			return nil
		}
		return args[2 : 2+n]
	default:
		return args[:1]
	}
}
