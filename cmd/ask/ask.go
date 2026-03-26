package ask

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"recall/internal/config"
	"recall/internal/embedding"
	"recall/internal/services/command_embedding"

	"github.com/spf13/cobra"
)

var askCmd = &cobra.Command{
	Use:   "ask",
	Short: "Search commands using semantic similarity",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			fmt.Println("Please provide a query.")
			return
		}

		cfg := config.LoadConfig()

		if !cfg.Embedding.IsEmbedEnabled {
			fmt.Println("Embedding is disabled in config.")
			fmt.Println("Enable it to use semantic search.")
			return
		}

		var embedder embedding.Embedder

		if cfg.Embedding.EmbeddingProvider == "ollama" {
			embedder = embedding.NewOllamaEmbedder(
				cfg.Embedding.OllamaEmbeddingBaseURL,
				cfg.Embedding.OllamaEmbeddingModel,
				cfg.Embedding.OllamaHttpTimeoutInSec,
			)
		} else {
			log.Fatalf("unsupported embedding provider: %s", cfg.Embedding.EmbeddingProvider)
		}

		query := strings.Join(args, " ")

		searchService, err := command_embedding.NewEmbeddingProcessor(embedder, cfg.Embedding.OllamaEmbeddingModel)
		if err != nil {
			log.Fatalf("failed to create search service: %v", err)
		}

		results, err := searchService.Search(query, cfg.Search.TopK)
		if err != nil {
			log.Fatalf("search failed: %v", err)
		}

		if len(results) == 0 {
			fmt.Println("No results found.")
			return
		}

		fmt.Println("Top matching commands")
		fmt.Println("──────────────────────")

		for i, r := range results {
			fmt.Printf("%d. %s (%.4f)\n", i+1, r.Command, r.Score)
		}

		fmt.Println()
		fmt.Print("Run a command? (enter number or n): ")

		var input string
		fmt.Scanln(&input)

		if input == "n" {
			fmt.Println("Skipped.")
			return
		}

		idx := int(input[0] - '1')

		if idx < 0 || idx >= len(results) {
			fmt.Println("Invalid selection.")
			return
		}

		selected := results[idx].Command

		fmt.Println()
		fmt.Println("Executing:", selected)

		err = runCommand(selected)
		if err != nil {
			fmt.Println("Command failed:", err)
		}
	},
}

func runCommand(command string) error {

	cmd := exec.Command("sh", "-c", command)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func GetAskCmd() *cobra.Command {
	return askCmd
}
