package shell

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBashShellName(t *testing.T) {
	b := NewBashShell()
	if b.Name() != "bash" {
		t.Errorf("Name() = %q, want %q", b.Name(), "bash")
	}
}

func TestBashHistoryFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	b := NewBashShell()
	want := filepath.Join(tmpHome, ".bash_history")
	if b.HistoryFile() != want {
		t.Errorf("HistoryFile() = %q, want %q", b.HistoryFile(), want)
	}
}

func TestBashReadHistoryMissingFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	b := NewBashShell()
	_, err := b.ReadHistory()
	if err == nil {
		t.Error("expected error for missing history file, got nil")
	}
}

func TestBashReadHistory(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	content := "git status\nkubectl get pods\ndocker ps\n"

	histFile := filepath.Join(tmpHome, ".bash_history")
	if err := os.WriteFile(histFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	b := NewBashShell()
	entries, err := b.ReadHistory()
	if err != nil {
		t.Fatalf("ReadHistory() returned error: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	commands := []string{"git status", "kubectl get pods", "docker ps"}
	for i, cmd := range commands {
		if entries[i].Command != cmd {
			t.Errorf("entry[%d].Command = %q, want %q", i, entries[i].Command, cmd)
		}
		// bash history has no timestamps — expect zero value
		if entries[i].Timestamp != (time.Time{}) {
			t.Errorf("entry[%d].Timestamp = %v, want zero value", i, entries[i].Timestamp)
		}
	}
}

func TestBashReadHistorySkipsBlankLines(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	content := "git status\n\nkubectl get pods\n\ndocker ps\n"

	histFile := filepath.Join(tmpHome, ".bash_history")
	if err := os.WriteFile(histFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	b := NewBashShell()
	entries, err := b.ReadHistory()
	if err != nil {
		t.Fatalf("ReadHistory() returned error: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("expected 3 entries (blank lines skipped), got %d", len(entries))
	}
}

func writeFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
}
