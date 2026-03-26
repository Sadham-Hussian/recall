package shell

import (
	"bufio"
	"os"
	"path/filepath"
	"time"
)

type BashShell struct{}

func NewBashShell() Shell {
	return &BashShell{}
}

func (b *BashShell) Name() string {
	return "bash"
}

func (b *BashShell) HistoryFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".bash_history")
}

func (b *BashShell) ReadHistory() ([]Entry, error) {
	file, err := os.Open(b.HistoryFile())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []Entry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		cmd := scanner.Text()
		if cmd == "" {
			continue
		}

		entries = append(entries, Entry{
			Command:   cmd,
			Timestamp: time.Time{}, // bash doesn’t store timestamp by default
		})
	}

	return entries, scanner.Err()
}
