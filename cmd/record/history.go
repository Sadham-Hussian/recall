package record

import (
	"fmt"
	"log"

	"recall/internal/config"
	"recall/internal/services/command_execution"

	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Import shell history into recall database",
	Run: func(cmd *cobra.Command, args []string) {

		config.LoadConfig()

		commandExecutionService, err := command_execution.NewCommandExecutionService()
		if err != nil {
			log.Fatalf("failed to create command execution service: %v", err)
		}
		imported, err := commandExecutionService.RecordCommandHistory()
		if err != nil {
			log.Fatalf("failed to import history: %v", err)
		}

		fmt.Printf("Imported %d history entries into database\n", imported)
	},
}

func GetHistoryCmd() *cobra.Command {
	return historyCmd
}
