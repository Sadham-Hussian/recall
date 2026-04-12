package session

import (
	"fmt"
	"log"
	"os"
	"recall/internal/config"
	"recall/internal/executor"
	"recall/internal/services/session"

	"github.com/spf13/cobra"
)

var continueCmd = &cobra.Command{
	Use:   "continue",
	Short: "Suggest next command based on previous workflows",
	Run: func(cmd *cobra.Command, args []string) {

		config.LoadConfig()

		shellPID := os.Getppid()

		sessionService, err := session.NewSessionService()
		if err != nil {
			log.Fatalf("failed to create session service: %v", err)
		}

		suggestion, err := sessionService.
			GetNextCommandSuggestion(shellPID)

		if err != nil {
			log.Fatalf("failed to fetch suggestion: %v", err)
		}

		if suggestion == nil {
			fmt.Println("No suggestion found.")
			return
		}

		fmt.Println("Suggested next command")
		fmt.Println("──────────────────────")
		fmt.Println(suggestion.NextCommand)
		fmt.Println()

		fmt.Print("Run this command? (y/n): ")

		var input string
		fmt.Scanln(&input)

		if input != "y" {
			fmt.Println("Skipped.")
			return
		}

		fmt.Println()
		fmt.Println("Executing:", suggestion.NextCommand)

		err = executor.RunCommand(suggestion.NextCommand)
		if err != nil {
			fmt.Println("Command failed:", err)
		}
	},
}

func GetContinueCmd() *cobra.Command {
	return continueCmd
}
