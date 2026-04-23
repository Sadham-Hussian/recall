package explain

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"recall/internal/config"
	"recall/internal/generation"
	"recall/internal/services/command_execution"
	explainsvc "recall/internal/services/explain"

	"github.com/spf13/cobra"
)

var last int

var explainCmd = &cobra.Command{
	Use:   "explain [command]",
	Short: "Explain a shell command using AI",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.LoadConfig()

		if !cfg.Explain.IsExplainEnabled {
			return fmt.Errorf("explain is disabled — set explain.is_explain_enabled: true in ~/.recall/config.yaml")
		}

		generator := createGenerator(cfg)

		svc := explainsvc.NewExplainService(generator)

		ctx, cancel := context.WithTimeout(cmd.Context(),
			time.Duration(cfg.Explain.TimeoutSeconds)*time.Second)
		defer cancel()

		if last > 0 {
			return explainLast(ctx, svc, last)
		}

		if len(args) == 0 {
			return fmt.Errorf("provide a command to explain, or use --last")
		}

		return explainCommand(ctx, svc, args[0])
	},
}

func createGenerator(cfg *config.Config) generation.Generator {
	switch cfg.Explain.Provider {
	case "ollama":
		return generation.NewOllamaGenerator(
			cfg.Explain.BaseURL,
			cfg.Explain.Model,
			cfg.Explain.TimeoutSeconds,
		)
	default:
		log.Fatalf("unsupported explain provider: %s", cfg.Explain.Provider)
		return nil
	}
}

func explainCommand(ctx context.Context, svc *explainsvc.ExplainService, command string) error {
	fmt.Printf("$ %s\n\n", command)

	stream, err := svc.Explain(ctx, command)
	if err != nil {
		return err
	}
	defer stream.Close()

	if _, err := io.Copy(os.Stdout, stream); err != nil {
		return err
	}

	fmt.Println()
	return nil
}

func explainLast(ctx context.Context, svc *explainsvc.ExplainService, count int) error {
	config.LoadConfig()

	execSvc, err := command_execution.NewCommandExecutionService()
	if err != nil {
		return err
	}

	commands, err := execSvc.ListRecent(count)
	if err != nil {
		return err
	}

	if len(commands) == 0 {
		fmt.Println("No commands found.")
		return nil
	}

	for i, cmd := range commands {
		if i > 0 {
			fmt.Println("\n────────────────\n")
		}
		if err := explainCommand(ctx, svc, cmd.Command); err != nil {
			fmt.Fprintf(os.Stderr, "error explaining '%s': %v\n", cmd.Command, err)
		}
	}

	return nil
}

func GetExplainCmd() *cobra.Command {
	explainCmd.Flags().IntVar(&last, "last", 0, "explain the last N commands")
	return explainCmd
}
