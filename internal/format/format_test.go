package format

import (
	"fmt"
	"testing"
	"time"
)

func TestExitSymbol(t *testing.T) {
	tests := []struct {
		code int
		want string
	}{
		{0, "✔"},
		{1, "✖"},
		{127, "✖"},
		{-1, "✖"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("code_%d", tc.code), func(t *testing.T) {
			got := ExitSymbol(tc.code)
			if got != tc.want {
				t.Errorf("ExitSymbol(%d) = %q, want %q", tc.code, got, tc.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		max   int
		want  string
	}{
		{"max zero", "hello", 0, ""},
		{"max one", "hello", 1, "h"},
		{"max two", "hello", 2, "he"},
		{"max three", "hello", 3, "hel"},
		{"short string fits", "hello", 8, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncate with ellipsis", "hello world", 5, "he..."},
		{"truncate longer string", "kubectl get pods -n default", 10, "kubectl..."},
		{"empty string", "", 5, ""},
		{"unicode safe", "héllo wörld", 5, "hé..."},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Truncate(tc.input, tc.max)
			if got != tc.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tc.input, tc.max, got, tc.want)
			}
		})
	}
}

func TestNormalizeCommand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no-op", "git push", "git push"},
		{"trailing spaces", "  git push  ", "git push"},
		{"internal newline", "git\npush", "gitpush"},
		{"carriage return newline", "git\r\npush", "gitpush"},
		{"escaped newline literal", `git\npush`, "gitpush"},
		{"multiple spaces", "git   push   origin", "git push origin"},
		{"mixed whitespace and newlines", "  git\n  push  \n  origin  ", "git push origin"},
		{"empty string", "", ""},
		{"only spaces", "   ", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeCommand(tc.input)
			if got != tc.want {
				t.Errorf("NormalizeCommand(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestRelativeTime(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"30 seconds ago", -30 * time.Second, "30s ago"},
		{"1 minute ago", -1 * time.Minute, "1m ago"},
		{"5 minutes ago", -5 * time.Minute, "5m ago"},
		{"1 hour ago", -1 * time.Hour, "1h ago"},
		{"3 hours ago", -3 * time.Hour, "3h ago"},
		{"1 day ago", -24 * time.Hour, "1d ago"},
		{"2 days ago", -48 * time.Hour, "2d ago"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := time.Now().Add(tc.duration).Unix()
			got := RelativeTime(ts)
			if got != tc.want {
				t.Errorf("RelativeTime(%v ago) = %q, want %q", -tc.duration, got, tc.want)
			}
		})
	}
}
