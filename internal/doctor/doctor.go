package doctor

import (
	"fmt"
	"recall/internal/config"
)

type CheckResult struct {
	Name    string
	OK      bool
	Message string
	Fix     string
}

func RunDoctor(cfg *config.Config) {
	results := []CheckResult{
		checkConfig(cfg),
		checkDB(cfg),
		checkTables(cfg),
	}

	// Embedding checks only if enabled
	if cfg.Embedding.IsEmbedEnabled {
		results = append(results,
			checkEmbeddingProvider(cfg),
			checkEmbeddingModel(cfg),
		)
	}

	printResults(results)
}

func printResults(results []CheckResult) {
	allGood := true

	for _, r := range results {
		if r.OK {
			fmt.Printf("✔ %s\n", r.Name)
		} else {
			allGood = false
			fmt.Printf("✘ %s\n", r.Name)

			if r.Message != "" {
				fmt.Printf("  → %s\n", r.Message)
			}
			if r.Fix != "" {
				fmt.Printf("  → %s\n", r.Fix)
			}
		}
	}

	if allGood {
		fmt.Println("\nAll good 🚀")
	}
}
