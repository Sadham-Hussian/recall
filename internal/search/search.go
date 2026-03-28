package search

import (
	"fmt"
	"math"
	"os/exec"
	"recall/internal/storage/models"
	"strings"
	"time"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

func ComputeHybridScore(r *models.HybridSearchResult, currentDir string, projectRoot string, currentSessionID string) float64 {

	now := time.Now().Unix()

	freqScore := math.Log(float64(r.Count) + 1)

	ageSeconds := now - r.LastTimestamp
	ageDays := float64(ageSeconds) / 86400
	recencyScore := 1.0 / (1.0 + ageDays)

	successRate := 0.0
	if r.Count > 0 {
		successRate = float64(r.SuccessCount) / float64(r.Count)
	}

	cwdScore := 0.0
	if r.Cwd == currentDir {
		cwdScore = 1.0
	}

	projectScore := 0.0

	if SameProject(r.Cwd, projectRoot) {
		projectScore = 1.0
	}

	fuzzyScore := 0.0
	if r.FuzzyScore > 0 {
		fuzzyScore = 1.0 / float64(r.FuzzyScore)
	}

	sessionScore := 0.0

	if r.SessionID == currentSessionID {
		sessionScore = 1.0
	}

	return (0.25 * freqScore) +
		(0.20 * recencyScore) +
		(0.10 * successRate) +
		(0.20 * fuzzyScore) +
		(0.10 * cwdScore) +
		(0.10 * projectScore) +
		(0.05 * sessionScore)
}

func FindFuzzyScore(r *models.HybridSearchResult, query string) int {
	return fuzzy.RankMatchFold(query, r.Command)
}

func BuildFTSQuery(query string) string {

	parts := strings.Fields(query)

	for i, p := range parts {

		p = strings.ReplaceAll(p, `"`, ``)

		parts[i] = fmt.Sprintf(`"%s"*`, p)
	}

	return strings.Join(parts, " ")
}

func FuzzySearchFilter(hsr []models.HybridSearchResult, query string) []models.HybridSearchResult {

	query = strings.ToLower(query)
	queryTokens := strings.Fields(query)

	filtered := make([]models.HybridSearchResult, 0)

	for _, r := range hsr {

		cmd := strings.ToLower(r.Command)
		cmdTokens := strings.Fields(cmd)

		matchCount := 0

		for _, q := range queryTokens {

			for _, t := range cmdTokens {

				if fuzzy.Match(q, t) {
					matchCount++
					break
				}
			}
		}

		if matchCount > 0 {
			filtered = append(filtered, r)
		}
	}

	return filtered
}

func GetProjectRoot(cwd string) string {

	cmd := exec.Command("git", "-C", cwd, "rev-parse", "--show-toplevel")

	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}

func SameProject(cmdCwd string, projectRoot string) bool {

	if projectRoot == "" {
		return false
	}

	return strings.HasPrefix(cmdCwd, projectRoot)
}

func BuildFTSQueryForSemanticSearch(q string) string {
	q = strings.ToLower(q)

	// Replace characters that break FTS parsing
	replacer := strings.NewReplacer(
		":", " ",
		"-", " ",
		".", " ",
		",", " ",
		"(", " ",
		")", " ",
		"\"", " ",
		"'", " ",
	)
	q = replacer.Replace(q)

	words := strings.Fields(q)

	// Optional: remove very short / useless tokens
	var terms []string
	for _, w := range words {
		if len(w) < 2 {
			continue
		}

		// skip common stopwords (optional but useful)
		switch w {
		case "the", "is", "if", "on", "in", "at", "to", "for", "a", "an", "of", "using", "check":
			continue
		}

		terms = append(terms, w+"*") // prefix match
	}

	// fallback: if everything removed, return empty
	if len(terms) == 0 {
		return ""
	}

	// OR-based query (better recall)
	return strings.Join(terms, " OR ")
}
