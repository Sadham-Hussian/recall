package ignore

import (
	"regexp"
	"strings"

	"recall/internal/config"
)

type Matcher struct {
	commands map[string]bool
	patterns []*regexp.Regexp
}

// NewMatcher compiles the ignore config into a reusable matcher.
// Invalid regex patterns are silently skipped so a bad pattern
// doesn't break recording entirely.
func NewMatcher(cfg *config.Config) *Matcher {
	m := &Matcher{
		commands: make(map[string]bool, len(cfg.Ignore.Commands)),
	}

	for _, cmd := range cfg.Ignore.Commands {
		m.commands[strings.ToLower(cmd)] = true
	}

	for _, p := range cfg.Ignore.Patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			continue
		}
		m.patterns = append(m.patterns, re)
	}

	return m
}

// ShouldIgnore returns true if the command should not be recorded.
// It checks exact first-token match against commands, then regex
// match against patterns. Returns on first match.
func (m *Matcher) ShouldIgnore(command string) bool {
	if command == "" {
		return true
	}

	// Extract the first token (binary name)
	firstToken := command
	if i := strings.IndexByte(command, ' '); i > 0 {
		firstToken = command[:i]
	}

	if m.commands[strings.ToLower(firstToken)] {
		return true
	}

	for _, re := range m.patterns {
		if re.MatchString(command) {
			return true
		}
	}

	return false
}
