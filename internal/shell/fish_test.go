package shell

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFishShellName(t *testing.T) {
	f := NewFishShell()
	if f.Name() != "fish" {
		t.Errorf("Name() = %q, want %q", f.Name(), "fish")
	}
}

func TestFishHistoryFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	f := NewFishShell()
	want := filepath.Join(tmpHome, ".local/share/fish/fish_history")
	if f.HistoryFile() != want {
		t.Errorf("HistoryFile() = %q, want %q", f.HistoryFile(), want)
	}
}

func TestFishReadHistoryMissingFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	f := NewFishShell()
	_, err := f.ReadHistory()
	if err == nil {
		t.Error("expected error for missing history file, got nil")
	}
}

func TestFishReadHistory(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	histDir := filepath.Join(tmpHome, ".local/share/fish")
	if err := os.MkdirAll(histDir, 0700); err != nil {
		t.Fatal(err)
	}

	content := "- cmd: git status\n" +
		"  when: 1700000000\n" +
		"- cmd: kubectl get pods\n" +
		"  when: 1700000100\n"

	histFile := filepath.Join(histDir, "fish_history")
	if err := os.WriteFile(histFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	f := NewFishShell()
	entries, err := f.ReadHistory()
	if err != nil {
		t.Fatalf("ReadHistory() returned error: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	expected := []struct {
		cmd string
		ts  int64
	}{
		{"git status", 1700000000},
		{"kubectl get pods", 1700000100},
	}

	for i, e := range expected {
		if entries[i].Command != e.cmd {
			t.Errorf("entry[%d].Command = %q, want %q", i, entries[i].Command, e.cmd)
		}
		if entries[i].Timestamp != time.Unix(e.ts, 0) {
			t.Errorf("entry[%d].Timestamp = %v, want %v", i, entries[i].Timestamp, time.Unix(e.ts, 0))
		}
	}
}

func TestFishReadHistoryMultipleEntries(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	histDir := filepath.Join(tmpHome, ".local/share/fish")
	if err := os.MkdirAll(histDir, 0700); err != nil {
		t.Fatal(err)
	}

	content := "- cmd: git status\n" +
		"  when: 1700000000\n" +
		"- cmd: kubectl get pods\n" +
		"  when: 1700000100\n" +
		"- cmd: docker ps\n" +
		"  when: 1700000200\n"

	histFile := filepath.Join(histDir, "fish_history")
	if err := os.WriteFile(histFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	f := NewFishShell()
	entries, err := f.ReadHistory()
	if err != nil {
		t.Fatalf("ReadHistory() returned error: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestFishReadHistoryInvalidTimestampSkipped(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	histDir := filepath.Join(tmpHome, ".local/share/fish")
	if err := os.MkdirAll(histDir, 0700); err != nil {
		t.Fatal(err)
	}

	// Fish parser only appends an entry when it sees "when:" — a cmd without "when:" won't produce an entry
	content := "- cmd: git status\n" +
		"  when: 1700000000\n" +
		"- cmd: orphan command without when\n" + // no "when:" line — not appended
		"- cmd: docker ps\n" +
		"  when: 1700000200\n"

	histFile := filepath.Join(histDir, "fish_history")
	if err := os.WriteFile(histFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	f := NewFishShell()
	entries, err := f.ReadHistory()
	if err != nil {
		t.Fatalf("ReadHistory() returned error: %v", err)
	}

	// Only 2 entries have "when:" lines
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}
