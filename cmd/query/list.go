package query

import (
	"fmt"
	"log"

	"recall/internal/config"
	"recall/internal/format"
	"recall/internal/services/command_execution"

	"github.com/spf13/cobra"
)

var listLimit int

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent commands stored in recall",
	Run: func(cmd *cobra.Command, args []string) {

		config.LoadConfig()

		commandExecutionService, err := command_execution.NewCommandExecutionService()
		if err != nil {
			log.Fatalf("failed to get last command: %v", err)
		}

		executions, err := commandExecutionService.ListRecent(listLimit)
		if err != nil {
			log.Fatalf("failed to fetch executions: %v", err)
		}

		for i, e := range executions {
			relative := format.RelativeTime(e.Timestamp)
			symbol := format.ExitSymbol(e.ExitCode)

			fmt.Printf(
				"%2d. %-10s %s  %s\n",
				i+1,
				relative,
				symbol,
				e.Command,
			)
		}
	},
}

func GetListCmd() *cobra.Command {
	listCmd.Flags().IntVarP(&listLimit, "limit", "n", 20, "Number of commands to list")
	return listCmd
}
