package query

import (
	"fmt"
	"log"
	"time"

	"recall/internal/config"
	"recall/internal/format"
	"recall/internal/services/command_execution"

	"github.com/spf13/cobra"
)

var lastCmd = &cobra.Command{
	Use:   "last",
	Short: "Show the last executed command",
	Run: func(cmd *cobra.Command, args []string) {

		config.LoadConfig()

		commandExecutionService, err := command_execution.NewCommandExecutionService()
		if err != nil {
			log.Fatalf("failed to get last command: %v", err)
		}

		execution, err := commandExecutionService.Last()
		if err != nil {
			log.Fatalf("no commands found")
		}

		t := time.Unix(execution.Timestamp, 0)

		fmt.Println("Last Command")
		fmt.Println("────────────")
		fmt.Printf("Time      : %s (%s)\n",
			t.Format("2006-01-02 15:04:05"),
			format.RelativeTime(execution.Timestamp),
		)
		fmt.Printf("Command   : %s\n", execution.Command)
		fmt.Printf("Exit Code : %d %s\n",
			execution.ExitCode,
			format.ExitSymbol(execution.ExitCode),
		)
		fmt.Printf("Directory : %s\n", execution.CWD)
	},
}

func GetLastCmd() *cobra.Command {
	return lastCmd
}
