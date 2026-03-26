package format

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func RelativeTime(ts int64) string {
	t := time.Unix(ts, 0)
	diff := time.Since(t)

	switch {
	case diff < time.Minute:
		return fmt.Sprintf("%ds ago", int(diff.Seconds()))
	case diff < time.Hour:
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	}
}

func ShortenPath(path string) string {
	home, _ := filepath.Abs("~")
	if strings.HasPrefix(path, home) {
		return strings.Replace(path, home, "~", 1)
	}
	return path
}

func ExitSymbol(code int) string {
	if code == 0 {
		return "✔"
	}
	return "✖"
}

func Truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}

	runes := []rune(s)

	if len(runes) <= max {
		return s
	}

	if max <= 3 {
		return string(runes[:max])
	}

	return string(runes[:max-3]) + "..."
}
