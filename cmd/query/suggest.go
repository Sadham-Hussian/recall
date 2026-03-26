package query

import (
	"fmt"
	"log"
	"recall/internal/config"
	"recall/internal/services/command_execution"
	"strings"

	"github.com/spf13/cobra"
)

var suggestLimit int

var suggestCmd = &cobra.Command{
	Use:   "suggest [command]",
	Short: "Suggest next commands based on history",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		cfg := config.LoadConfig()

		command := strings.Join(args, " ")

		commandChainService, err := command_execution.NewCommandChainService()
		if err != nil {
			log.Fatalf("failed to create command chain service: %v", err)
		}

		results, err := commandChainService.GetNextCommands(cfg, command, suggestLimit)
		if err != nil {
			log.Fatalf("failed to fetch suggestions: %v", err)
		}

		if len(results) == 0 {
			fmt.Println("No suggestions found.")
			return
		}

		fmt.Println("Suggested next commands")
		fmt.Println("──────────────────────")
		fmt.Println()

		for i, r := range results {
			fmt.Printf("%2d. %s  (%d times)\n",
				i+1,
				r.NextCommand,
				r.OccurrenceCount,
			)
		}
	},
}

func GetSuggestCmd() *cobra.Command {
	suggestCmd.Flags().IntVarP(&suggestLimit, "limit", "n", 5, "Number of suggestions")
	return suggestCmd
}
