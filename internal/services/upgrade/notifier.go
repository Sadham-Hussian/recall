package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"recall/internal/config"
)

type checkState struct {
	LastCheckedAt time.Time `json:"last_checked_at"`
	LatestVersion string    `json:"latest_version"`
	ReleaseURL    string    `json:"release_url"`
}

func stateFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".recall", ".update-check.json"), nil
}

func loadState() (*checkState, error) {
	p, err := stateFilePath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var s checkState
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func saveState(s *checkState) error {
	p, err := stateFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0644)
}

// MaybeCheckInBackground fires a GitHub API probe in a goroutine if the
// last probe was longer ago than the configured interval. It never blocks
// the caller and never surfaces errors: a failed check just means the
// banner won't appear.
func MaybeCheckInBackground(cfg *config.Config) {
	if cfg == nil || !cfg.Upgrade.AutoCheckEnabled {
		return
	}
	interval := time.Duration(cfg.Upgrade.CheckIntervalHours) * time.Hour
	if interval <= 0 {
		return
	}

	if state, err := loadState(); err == nil && state != nil {
		if time.Since(state.LastCheckedAt) < interval {
			return
		}
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		rel, err := LatestRelease(ctx)
		if err != nil {
			return
		}
		_ = saveState(&checkState{
			LastCheckedAt: time.Now(),
			LatestVersion: rel.TagName,
			ReleaseURL:    rel.HTMLURL,
		})
	}()
}

// silentCommands suppress the banner. `record` fires on every shell prompt,
// `hook` prints shell integration to stdout, `daemon` runs long and is
// often backgrounded, `migrate` runs during setup, and `upgrade` is
// redundant (user is already upgrading).
var silentCommands = map[string]bool{
	"record":  true,
	"hook":    true,
	"daemon":  true,
	"migrate": true,
	"upgrade": true,
}

// PrintNoticeIfAvailable emits a single line to stderr at command exit
// when a cached newer version is on hand.
func PrintNoticeIfAvailable(cfg *config.Config, currentVersion, currentCmd string) {
	if cfg == nil || !cfg.Upgrade.AutoCheckEnabled {
		return
	}
	if silentCommands[currentCmd] {
		return
	}

	state, err := loadState()
	if err != nil || state == nil || state.LatestVersion == "" {
		return
	}
	if !IsNewer(currentVersion, state.LatestVersion) {
		return
	}

	exe, _ := os.Executable()
	if exe != "" {
		exe, _ = filepath.EvalSymlinks(exe)
	}
	cmdHint := "run `recall upgrade` to update"
	if mgr, ok := IsManagedInstall(exe); ok && mgr == "homebrew" {
		cmdHint = "run `brew upgrade Sadham-Hussian/recall/recall` to update"
	}
	fmt.Fprintf(os.Stderr,
		"\n─── recall %s available (you have %s) · %s ───\n",
		state.LatestVersion, currentVersion, cmdHint)
}
