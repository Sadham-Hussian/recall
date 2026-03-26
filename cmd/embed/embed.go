package embed

import (
	"fmt"
	"log"

	"recall/internal/config"
	"recall/internal/embedding"
	"recall/internal/services/command_embedding"

	"github.com/spf13/cobra"
)

var embedCmd = &cobra.Command{
	Use:   "embed",
	Short: "Process embedding queue and generate embeddings",
	Run: func(cmd *cobra.Command, args []string) {

		cfg := config.LoadConfig()

		if !cfg.Embedding.IsEmbedEnabled {
			fmt.Println("Embedding is disabled in config.")
			fmt.Println("Enable it to use semantic search.")
			return
		}

		var embedder embedding.Embedder

		var model string

		if cfg.Embedding.EmbeddingProvider == "ollama" {
			embedder = embedding.NewOllamaEmbedder(
				cfg.Embedding.OllamaEmbeddingBaseURL,
				cfg.Embedding.OllamaEmbeddingModel,
				cfg.Embedding.OllamaHttpTimeoutInSec,
			)
			model = cfg.Embedding.OllamaEmbeddingModel
		} else {
			log.Fatalf("unsupported embedding provider: %s", cfg.Embedding.EmbeddingProvider)
		}

		fmt.Println("Starting embedding processor...")
		fmt.Println("Batch size: 500")
		fmt.Println()

		processor, err := command_embedding.NewEmbeddingProcessor(embedder, model)
		if err != nil {
			log.Fatalf("failed to create embedding processor: %v", err)
		}

		err = processor.Process(cfg.Processor.BatchSize)
		if err != nil {
			log.Fatalf("embedding process failed: %v", err)
		}

		fmt.Println()
		fmt.Println("Embedding completed successfully.")
	},
}

func GetEmbedCmd() *cobra.Command {
	return embedCmd
}
