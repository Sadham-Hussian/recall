package search

import (
	"recall/internal/storage/models"
	"testing"
	"time"
)

func TestBuildFTSQuery(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty string", "", ""},
		{"single word", "git", `"git"*`},
		{"two words", "git push", `"git"* "push"*`},
		{"three words", "kubectl get pods", `"kubectl"* "get"* "pods"*`},
		{"strips double quotes", `say "hello"`, `"say"* "hello"*`},
		{"extra spaces handled by Fields", "git  push", `"git"* "push"*`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := BuildFTSQuery(tc.input)
			if got != tc.want {
				t.Errorf("BuildFTSQuery(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestBuildFTSQueryForSemanticSearch(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"basic words", "git push origin", "git* OR push* OR origin*"},
		{"stopwords filtered", "check the status", "status*"},
		{"only punctuation", ":", ""},
		{"mixed punctuation and words", "ls -la /home", "ls* OR la* OR /home*"},
		{"uppercase lowercased", "GIT PUSH", "git* OR push*"},
		{"single char tokens removed", "a b cd", "cd*"},
		{"empty string", "", ""},
		{"colon separated", "kubectl:get", "kubectl* OR get*"},
		{"dot separated", "v1.0.0", "v1*"}, // "0" is single-char, filtered out
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := BuildFTSQueryForSemanticSearch(tc.input)
			if got != tc.want {
				t.Errorf("BuildFTSQueryForSemanticSearch(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSameProject(t *testing.T) {
	tests := []struct {
		name        string
		cmdCwd      string
		projectRoot string
		want        bool
	}{
		{"subdir of project", "/home/user/proj/src", "/home/user/proj", true},
		{"exact project root", "/home/user/proj", "/home/user/proj", true},
		{"different project", "/home/user/other", "/home/user/proj", false},
		{"empty project root", "/home/user/proj", "", false},
		{"both empty", "", "", false},
		{"prefix but not path prefix", "/home/user/project", "/home/user/proj", true}, // HasPrefix match
		{"unrelated paths", "/tmp/work", "/home/user/proj", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := SameProject(tc.cmdCwd, tc.projectRoot)
			if got != tc.want {
				t.Errorf("SameProject(%q, %q) = %v, want %v", tc.cmdCwd, tc.projectRoot, got, tc.want)
			}
		})
	}
}

func TestFuzzySearchFilter(t *testing.T) {
	results := []models.HybridSearchResult{
		{Command: "git push origin"},
		{Command: "kubectl get pods"},
		{Command: "docker build ."},
	}

	t.Run("matches single token", func(t *testing.T) {
		got := FuzzySearchFilter(results, "git")
		if len(got) != 1 || got[0].Command != "git push origin" {
			t.Errorf("expected [git push origin], got %v", got)
		}
	})

	t.Run("matches multi-token query", func(t *testing.T) {
		got := FuzzySearchFilter(results, "kube pod")
		if len(got) != 1 || got[0].Command != "kubectl get pods" {
			t.Errorf("expected [kubectl get pods], got %v", got)
		}
	})

	t.Run("no match returns empty", func(t *testing.T) {
		got := FuzzySearchFilter(results, "zzz")
		if len(got) != 0 {
			t.Errorf("expected empty, got %v", got)
		}
	})

	t.Run("empty query matches nothing", func(t *testing.T) {
		got := FuzzySearchFilter(results, "")
		if len(got) != 0 {
			t.Errorf("expected empty for blank query, got %v", got)
		}
	})

	t.Run("empty input returns empty", func(t *testing.T) {
		got := FuzzySearchFilter([]models.HybridSearchResult{}, "git")
		if len(got) != 0 {
			t.Errorf("expected empty for empty input, got %v", got)
		}
	})
}

func TestComputeHybridScore(t *testing.T) {
	now := time.Now().Unix()

	highResult := &models.HybridSearchResult{
		Command:       "git push origin",
		Count:         100,
		LastTimestamp: now - 60, // 1 minute ago
		SuccessCount:  100,
		Cwd:           "/home/user/proj",
		SessionID:     "session-1",
		FuzzyScore:    1,
	}

	lowResult := &models.HybridSearchResult{
		Command:       "rm -rf /",
		Count:         1,
		LastTimestamp: now - 30*24*3600, // 30 days ago
		SuccessCount:  0,
		Cwd:           "/tmp",
		SessionID:     "session-99",
		FuzzyScore:    0,
	}

	currentDir := "/home/user/proj"
	projectRoot := "/home/user/proj"
	currentSession := "session-1"

	highScore := ComputeHybridScore(highResult, currentDir, projectRoot, currentSession)
	lowScore := ComputeHybridScore(lowResult, "/other/dir", "/other", "session-other")

	t.Run("high score result scores higher than low score result", func(t *testing.T) {
		if highScore <= lowScore {
			t.Errorf("expected highScore (%f) > lowScore (%f)", highScore, lowScore)
		}
	})

	t.Run("all scores are non-negative", func(t *testing.T) {
		if highScore < 0 {
			t.Errorf("highScore is negative: %f", highScore)
		}
		if lowScore < 0 {
			t.Errorf("lowScore is negative: %f", lowScore)
		}
	})

	t.Run("zero fuzzy score does not cause division by zero", func(t *testing.T) {
		r := &models.HybridSearchResult{
			Command:       "ls",
			Count:         1,
			LastTimestamp: now,
			SuccessCount:  1,
			FuzzyScore:    0,
		}
		score := ComputeHybridScore(r, "", "", "")
		if score < 0 {
			t.Errorf("score with FuzzyScore=0 should be non-negative, got %f", score)
		}
	})

	t.Run("session match adds score", func(t *testing.T) {
		base := &models.HybridSearchResult{
			Command: "git status", Count: 1, LastTimestamp: now, SuccessCount: 1,
		}
		withSession := &models.HybridSearchResult{
			Command: "git status", Count: 1, LastTimestamp: now, SuccessCount: 1,
			SessionID: "sess-abc",
		}
		scoreBase := ComputeHybridScore(base, "", "", "sess-abc")
		scoreWith := ComputeHybridScore(withSession, "", "", "sess-abc")
		if scoreWith <= scoreBase {
			t.Errorf("session match should increase score: with=%f base=%f", scoreWith, scoreBase)
		}
	})

	t.Run("cwd match adds score", func(t *testing.T) {
		withCwd := &models.HybridSearchResult{
			Command: "make build", Count: 1, LastTimestamp: now, SuccessCount: 1,
			Cwd: "/home/user/proj",
		}
		withoutCwd := &models.HybridSearchResult{
			Command: "make build", Count: 1, LastTimestamp: now, SuccessCount: 1,
			Cwd: "/tmp",
		}
		scoreWith := ComputeHybridScore(withCwd, "/home/user/proj", "", "")
		scoreWithout := ComputeHybridScore(withoutCwd, "/home/user/proj", "", "")
		if scoreWith <= scoreWithout {
			t.Errorf("cwd match should increase score: with=%f without=%f", scoreWith, scoreWithout)
		}
	})
}
