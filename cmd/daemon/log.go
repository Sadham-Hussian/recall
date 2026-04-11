package daemon

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	maxLogSizeMB = 10
	maxBackups   = 1
)

func openLogFile(logPath string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return nil, err
	}
	rotateIfNeeded(logPath)
	return os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

// rotateIfNeeded renames daemon.log → daemon.log.1 → daemon.log.2 … up to maxBackups.
// The oldest backup (maxBackups) is overwritten/dropped.
func rotateIfNeeded(logPath string) {
	info, err := os.Stat(logPath)
	if err != nil || info.Size() <= int64(maxLogSizeMB)*1024*1024 {
		return
	}
	// shift backups: .log.2 → .log.3 (drop), .log.1 → .log.2, .log → .log.1
	for i := maxBackups; i >= 1; i-- {
		var src string
		if i == 1 {
			src = logPath
		} else {
			src = fmt.Sprintf("%s.%d", logPath, i-1)
		}
		dst := fmt.Sprintf("%s.%d", logPath, i)
		os.Rename(src, dst) // ignore error — src may not exist yet
	}
}

// checkRotation checks the current log size and rotates + reopens if needed.
// Returns the (possibly new) file handle.
func checkRotation(logPath string, current *os.File) *os.File {
	info, err := os.Stat(logPath)
	if err != nil || info.Size() <= int64(maxLogSizeMB)*1024*1024 {
		return current
	}
	current.Close()
	rotateIfNeeded(logPath)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// failed to reopen — fall back to stderr so we don't lose output
		log.SetOutput(os.Stderr)
		return current
	}
	log.SetOutput(f)
	return f
}
