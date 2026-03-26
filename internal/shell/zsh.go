package shell

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ZshShell struct{}

func NewZshShell() Shell {
	return &ZshShell{}
}

func (z *ZshShell) Name() string {
	return "zsh"
}

func (z *ZshShell) HistoryFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".zsh_history")
}

func (z *ZshShell) ReadHistory() ([]Entry, error) {
	file, err := os.Open(z.HistoryFile())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []Entry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// Format: ": 1709112242:0;kubectl get pods"
		if !strings.HasPrefix(line, ":") {
			continue
		}

		parts := strings.SplitN(line, ";", 2)
		if len(parts) != 2 {
			continue
		}

		meta := parts[0]
		cmd := parts[1]

		metaParts := strings.Split(meta, ":")
		if len(metaParts) < 2 {
			continue
		}

		ts, err := strconv.ParseInt(strings.TrimSpace(metaParts[1]), 10, 64)
		if err != nil {
			continue
		}

		entries = append(entries, Entry{
			Command:   cmd,
			Timestamp: time.Unix(ts, 0),
		})
	}

	return entries, scanner.Err()
}
