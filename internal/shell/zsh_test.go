package shell

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestZshShellName(t *testing.T) {
	z := NewZshShell()
	if z.Name() != "zsh" {
		t.Errorf("Name() = %q, want %q", z.Name(), "zsh")
	}
}

func TestZshHistoryFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	z := NewZshShell()
	want := filepath.Join(tmpHome, ".zsh_history")
	if z.HistoryFile() != want {
		t.Errorf("HistoryFile() = %q, want %q", z.HistoryFile(), want)
	}
}

func TestZshReadHistoryMissingFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	z := NewZshShell()
	_, err := z.ReadHistory()
	if err == nil {
		t.Error("expected error for missing history file, got nil")
	}
}

func TestZshReadHistory(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	content := ": 1700000000:0;git status\n" +
		": 1700000100:0;kubectl get pods\n" +
		": 1700000200:0;docker ps\n"

	histFile := filepath.Join(tmpHome, ".zsh_history")
	if err := os.WriteFile(histFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	z := NewZshShell()
	entries, err := z.ReadHistory()
	if err != nil {
		t.Fatalf("ReadHistory() returned error: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	expected := []struct {
		cmd string
		ts  int64
	}{
		{"git status", 1700000000},
		{"kubectl get pods", 1700000100},
		{"docker ps", 1700000200},
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

func TestZshReadHistorySkipsInvalidLines(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	content := ": 1700000000:0;git status\n" +
		": badline\n" +           // no semicolon
		"notahistoryline\n" +     // no leading colon
		": notanumber:0;ls\n" +   // non-numeric timestamp
		": 1700000200:0;docker ps\n"

	histFile := filepath.Join(tmpHome, ".zsh_history")
	if err := os.WriteFile(histFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	z := NewZshShell()
	entries, err := z.ReadHistory()
	if err != nil {
		t.Fatalf("ReadHistory() returned error: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 valid entries, got %d: %v", len(entries), entries)
	}
}
