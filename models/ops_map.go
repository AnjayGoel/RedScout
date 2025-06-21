package models

import (
	"strings"
)

type OpType string

const (
	GetOp     OpType = "GET"
	SetOp     OpType = "SET"
	DelOp     OpType = "DEL"
	EvalOp    OpType = "Eval"
	TotalOp   OpType = "Total"
	UnknownOp OpType = "Unknown"
)

// Map of Redis commands to their operation category
var redisCommandType = map[string]OpType{
	// GET-like
	"GET":       GetOp,
	"MGET":      GetOp,
	"HGET":      GetOp,
	"HMGET":     GetOp,
	"HGETALL":   GetOp,
	"ZRANGE":    GetOp,
	"ZREVRANGE": GetOp,
	"LRANGE":    GetOp,
	"SCARD":     GetOp,
	"SISMEMBER": GetOp,
	"ZCARD":     GetOp,
	"ZRANK":     GetOp,
	"GETBIT":    GetOp,
	"EXISTS":    GetOp,
	"TTL":       GetOp,
	"PTTL":      GetOp,
	"TYPE":      GetOp,
	"KEYS":      GetOp,
	"SCAN":      GetOp,

	// SET-like
	"SET":          SetOp,
	"MSET":         SetOp,
	"HSET":         SetOp,
	"HMSET":        SetOp,
	"LPUSH":        SetOp,
	"RPUSH":        SetOp,
	"SADD":         SetOp,
	"ZADD":         SetOp,
	"SETBIT":       SetOp,
	"INCR":         SetOp,
	"DECR":         SetOp,
	"INCRBY":       SetOp,
	"APPEND":       SetOp,
	"SETEX":        SetOp,
	"PSETEX":       SetOp,
	"SETNX":        SetOp,
	"ZINCRBY":      SetOp,
	"EXPIRE":       SetOp,
	"PEXPIRE":      SetOp,
	"EXPIREAT":     SetOp,
	"PERSIST":      SetOp,
	"RENAME":       SetOp,
	"RENAMENX":     SetOp,
	"MOVE":         SetOp,
	"LSET":         SetOp,
	"LINSERT":      SetOp,
	"HINCRBY":      SetOp,
	"HINCRBYFLOAT": SetOp,

	// DEL-like
	"DEL":      DelOp,
	"UNLINK":   DelOp,
	"FLUSHDB":  DelOp,
	"FLUSHALL": DelOp,
	"LPOP":     DelOp,
	"RPOP":     DelOp,
	"SPOP":     DelOp,
	"ZREM":     DelOp,
	"HDEL":     DelOp,
	"SREM":     DelOp,
	"LTRIM":    DelOp,

	// EVAL-like
	"EVAL": EvalOp,
}

func GetOpType(command string) OpType {
	upper := strings.ToUpper(command)
	if op, ok := redisCommandType[upper]; ok {
		return op
	}
	return UnknownOp
}
