package shell

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type FishShell struct{}

func NewFishShell() Shell {
	return &FishShell{}
}

func (f *FishShell) Name() string {
	return "fish"
}

func (f *FishShell) HistoryFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local/share/fish/fish_history")
}

func (f *FishShell) ReadHistory() ([]Entry, error) {
	file, err := os.Open(f.HistoryFile())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []Entry
	var currentCmd string
	var currentTime int64

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "- cmd:") {
			currentCmd = strings.TrimSpace(strings.TrimPrefix(line, "- cmd:"))
		}

		if strings.HasPrefix(line, "when:") {
			tsStr := strings.TrimSpace(strings.TrimPrefix(line, "when:"))
			currentTime, _ = strconv.ParseInt(tsStr, 10, 64)

			entries = append(entries, Entry{
				Command:   currentCmd,
				Timestamp: time.Unix(currentTime, 0),
			})
		}
	}

	return entries, scanner.Err()
}
