package scanner

import (
	"fmt"
	"strconv"
	"strings"

	"redscout/models"
)

const (
	maxStringPreviewBytes = 4096
	previewItems          = 10
)

func (s *Scanner) FetchKeyValue(key string) {
	keyType, err := s.redis.Type(s.ctx, key).Result()
	if err != nil {
		s.State.KeyValue = &models.KeyValueInfo{Key: key, Type: "error", Value: fmt.Sprintf("Failed to fetch: %v", err)}
		s.publish()
		return
	}

	info := &models.KeyValueInfo{Key: key, Type: keyType, TTL: -1}
	if ttl, err := s.redis.TTL(s.ctx, key).Result(); err == nil && ttl > 0 {
		info.TTL = int64(ttl.Seconds())
	}
	if mem, err := s.redis.MemoryUsage(s.ctx, key).Result(); err == nil {
		info.Size = mem
	}
	info.Length, info.Value = s.previewValue(key, keyType)

	s.State.KeyValue = info
	s.publish()
}

func (s *Scanner) previewValue(key, keyType string) (int64, string) {
	ctx := s.ctx
	switch keyType {
	case "string":
		length := s.redis.StrLen(ctx, key).Val()
		value := s.redis.GetRange(ctx, key, 0, maxStringPreviewBytes-1).Val()
		if remaining := length - int64(len(value)); remaining > 0 {
			value += fmt.Sprintf("\n... (%d more bytes)", remaining)
		}
		return length, value
	case "list":
		length := s.redis.LLen(ctx, key).Val()
		return length, previewLines(s.redis.LRange(ctx, key, 0, previewItems-1).Val(), length)
	case "set":
		length := s.redis.SCard(ctx, key).Val()
		return length, previewLines(s.redis.SRandMemberN(ctx, key, previewItems).Val(), length)
	case "hash":
		length := s.redis.HLen(ctx, key).Val()
		// Use HScan to avoid loading entire hash into memory
		fields, _ := s.redis.HScan(ctx, key, 0, "*", previewItems).Val()
		lines := make([]string, 0, previewItems)
		// HScan returns alternating key, value pairs
		for i := 0; i+1 < len(fields) && len(lines) < previewItems; i += 2 {
			lines = append(lines, fields[i]+": "+fields[i+1])
		}
		return length, previewLines(lines, length)
	case "zset":
		length := s.redis.ZCard(ctx, key).Val()
		members := s.redis.ZRangeWithScores(ctx, key, 0, previewItems-1).Val()
		lines := make([]string, 0, len(members))
		for _, z := range members {
			lines = append(lines, fmt.Sprintf("%s: %v", strconv.FormatFloat(z.Score, 'f', -1, 64), z.Member))
		}
		return length, previewLines(lines, length)
	case "stream":
		length := s.redis.XLen(ctx, key).Val()
		return length, fmt.Sprintf("stream with %d entries", length)
	default:
		return 0, fmt.Sprintf("(%s type)", keyType)
	}
}

func previewLines(lines []string, total int64) string {
	text := strings.Join(lines, "\n")
	if remaining := total - int64(len(lines)); remaining > 0 {
		text += fmt.Sprintf("\n... (%d more)", remaining)
	}
	return text
}
