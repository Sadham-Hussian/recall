package daemon

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"recall/internal/config"
	"recall/internal/embedding"
	"recall/internal/services/command_embedding"

	"github.com/spf13/cobra"
)

var pollInterval int

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Manage the background embedding daemon",
	Long:  "Run or manage the recall background embedding daemon.",
	// No Run — calling `recall daemon` alone prints help
}

var daemonRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the daemon in the foreground",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.LoadConfig()

		// Open log file — daemon owns it directly for rotation support
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("could not determine home dir: %v", err)
		}
		logPath := filepath.Join(home, ".recall", "daemon.log")
		logFile, err := openLogFile(logPath)
		if err != nil {
			log.Fatalf("could not open log file: %v", err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)
		log.SetFlags(0) // no timestamps prefix — messages include context already

		if !cfg.Embedding.IsEmbedEnabled {
			log.Println("[recall daemon] embedding is disabled in config — exiting")
			return
		}

		if pollInterval <= 0 {
			pollInterval = cfg.Daemon.PollIntervalSeconds
		}
		if pollInterval <= 0 {
			pollInterval = 30
		}

		var emb embedding.Embedder
		var model string

		if cfg.Embedding.EmbeddingProvider == "ollama" {
			emb = embedding.NewOllamaEmbedder(
				cfg.Embedding.OllamaEmbeddingBaseURL,
				cfg.Embedding.OllamaEmbeddingModel,
				cfg.Embedding.OllamaHttpTimeoutInSec,
			)
			model = cfg.Embedding.OllamaEmbeddingModel
		} else {
			log.Fatalf("unsupported embedding provider: %s", cfg.Embedding.EmbeddingProvider)
		}

		processor, err := command_embedding.NewEmbeddingProcessor(emb, model)
		if err != nil {
			log.Fatalf("failed to create embedding processor: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
		defer ticker.Stop()

		log.Printf("[recall daemon] started — poll interval: %ds", pollInterval)

		logFile = runTick(cfg, processor, logPath, logFile)

		for {
			select {
			case <-ticker.C:
				logFile = runTick(cfg, processor, logPath, logFile)
			case <-sigs:
				log.Println("[recall daemon] shutting down")
				return
			case <-ctx.Done():
				return
			}
		}
	},
}

var tickCount int

func runTick(cfg *config.Config, processor *command_embedding.CommandEmbeddingProcessor, logPath string, logFile *os.File) *os.File {
	tickCount++

	// Check log rotation every 10 ticks
	if tickCount%10 == 0 {
		logFile = checkRotation(logPath, logFile)
	}

	// Reload config so runtime changes (e.g. disabling embedding) take effect
	reloaded, err := config.ReloadConfig()
	if err != nil {
		log.Printf("[recall daemon] config reload error: %v (using previous config)", err)
	} else {
		cfg = reloaded
	}

	if !cfg.Embedding.IsEmbedEnabled {
		if tickCount%10 == 0 {
			log.Printf("[recall daemon] tick %d — embedding disabled", tickCount)
		}
		return logFile
	}

	hasPending, err := processor.HasPendingEmbeddings()
	if err != nil {
		log.Printf("[recall daemon] queue check error: %v", err)
		return logFile
	}

	if !hasPending {
		if tickCount%10 == 0 {
			log.Printf("[recall daemon] tick %d — queue empty", tickCount)
		}
		return logFile
	}

	log.Printf("[recall daemon] tick %d — processing queue", tickCount)
	if err := processor.Process(cfg.Processor.BatchSize); err != nil {
		log.Printf("[recall daemon] error: %v", err)
	}
	return logFile
}

func GetDaemonCmd() *cobra.Command {
	daemonRunCmd.Flags().IntVarP(&pollInterval, "interval", "i", 0, "poll interval in seconds (overrides config daemon.poll_interval_seconds)")

	daemonCmd.AddCommand(daemonRunCmd)
	daemonCmd.AddCommand(installCmd)
	daemonCmd.AddCommand(startCmd)
	daemonCmd.AddCommand(stopCmd)
	daemonCmd.AddCommand(statusCmd)

	return daemonCmd
}
