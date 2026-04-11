package query

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"recall/internal/config"
	"recall/internal/format"
	"recall/internal/services/command_execution"
	"recall/internal/storage/models"

	"github.com/manifoldco/promptui"

	"github.com/spf13/cobra"
)

var searchLimit int
var defaultSearchLimit int
var compactOutput bool
var fullOutput bool
var pick bool
var interactive bool
var fuzzySearchLimit int

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search commands using full-text search",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		cfg := config.LoadConfig()

		commandExecutionSearchService, err := command_execution.NewCommandExecutionSearchService()
		if err != nil {
			log.Fatalf("failed to create command execution search service: %v", err)
		}

		results, err := commandExecutionSearchService.Search(cfg, args, searchLimit)
		if err != nil {
			log.Fatalf("search failed: %v", err)
		}

		fmt.Println("Search Results")
		fmt.Println("──────────────")
		fmt.Println()

		if fullOutput {
			compactOutput = false
		}

		if interactive {
			interactiveCommandPicker(results)
		} else if pick {
			commands := printSearchResult(results)
			if pick {
				pickCommand(commands)
			}
		} else {
			printSearchResult(results)
		}
	},
}

func printSearchResult(results []models.HybridSearchResult) []string {
	commands := make([]string, 0, len(results))

	for i, r := range results {

		cmd := strings.TrimSpace(r.Command)

		successRate := 0
		if r.Count > 0 {
			successRate = int((float64(r.SuccessCount) / float64(r.Count)) * 100)
		}

		if compactOutput {

			fmt.Printf(
				"%2d. %-42s %4d times   ✔%3d%%   last: %s\n",
				i+1,
				format.Truncate(cmd, 42),
				r.Count,
				successRate,
				format.RelativeTime(r.LastTimestamp),
			)

		} else {

			fmt.Printf(
				"%2d. %4d times   ✔%3d%%   last: %s\n",
				i+1,
				r.Count,
				successRate,
				format.RelativeTime(r.LastTimestamp),
			)

			fmt.Printf("    %s\n\n", cmd)
		}

		commands = append(commands, cmd)
	}

	return commands
}

func pickCommand(commands []string) {

	fmt.Print("\nSelect command number: ")

	var choice int
	_, err := fmt.Scanln(&choice)
	if err != nil || choice < 1 || choice > len(commands) {
		fmt.Println("Invalid selection")
		return
	}

	selected := commands[choice-1]

	fmt.Println("\nExecuting:")
	fmt.Println(selected)

	cmd := exec.Command("sh", "-c", selected)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Println("Command failed:", err)
	}

}

func interactiveCommandPicker(results []models.HybridSearchResult) {

	items := make([]string, len(results))
	for i, r := range results {
		items[i] = r.Command
	}

	prompt := promptui.Select{
		Label: "Select command",
		Items: items,
		Size:  10,
	}

	index, _, err := prompt.Run()
	if err != nil {
		fmt.Println("Selection cancelled")
		return
	}

	selected := items[index]

	fmt.Println("\nExecuting:")
	fmt.Println(selected)

	cmd := exec.Command("sh", "-c", selected)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Println("Command failed:", err)
	}

}

func GetSearchCmd() *cobra.Command {
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "n", 10, "Number of results")
	searchCmd.Flags().BoolVar(&compactOutput, "compact", true, "Compact single-line output")
	searchCmd.Flags().BoolVar(&fullOutput, "full", false, "Show full command output")
	searchCmd.Flags().BoolVar(&pick, "pick", false, "Interactively pick a command to run")
	searchCmd.Flags().BoolVar(&interactive, "interactive", false, "Interactive command picker")
	return searchCmd
}
