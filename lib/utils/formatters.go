package utils

import (
	"fmt"
	"strings"
)

func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func FormatNumber(n float64) string {
	switch {
	case n >= 1_000_000_000:
		return fmt.Sprintf("%.1fB", n/1_000_000_000)
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", n/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", n/1_000)
	default:
		return fmt.Sprintf("%.0f", n)
	}
}

func FormatOpsPerSec(n float64) string {
	return fmt.Sprintf("%s/s", FormatNumber(n))
}

func FormatDuration(seconds int64) string {
	if seconds < 0 {
		return "-" + FormatDuration(-seconds)
	}

	d := seconds / 86400
	h := (seconds % 86400) / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60

	result := ""
	if d > 0 {
		result += fmt.Sprintf("%dd ", d)
	}
	if h > 0 || d > 0 {
		result += fmt.Sprintf("%dh ", h)
	}
	if m > 0 || (h > 0 && s > 0) || (d > 0 && (h > 0 || s > 0)) {
		result += fmt.Sprintf("%dm ", m)
	}
	if s > 0 || (d == 0 && h == 0 && m == 0) {
		result += fmt.Sprintf("%ds", s)
	}
	return strings.TrimSpace(result)
}
