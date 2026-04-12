package session

import (
	"fmt"
	"log"
	"recall/internal/config"
	"recall/internal/executor"
	"recall/internal/services/session"
	"recall/internal/storage/models"

	"github.com/spf13/cobra"
)

var replayCmd = &cobra.Command{
	Use:   "replay [session_id]",
	Short: "Replay commands from a session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		config.LoadConfig()

		sessionID := args[0]

		sessionService, err := session.NewSessionService()
		if err != nil {
			log.Fatalf("failed to create session service: %v", err)
		}

		commands, err := sessionService.GetCommandsBySessionID(sessionID)
		if err != nil {
			log.Fatalf("failed to fetch session commands: %v", err)
		}

		if len(commands) == 0 {
			fmt.Println("No commands found for this session.")
			return
		}

		printWorkflow(sessionID, commands)

		mode := askExecutionMode()

		switch mode {

		case 1:
			executor.RunAll(toStrings(commands))

		case 2:
			executor.RunStepByStep(toStrings(commands))

		case 3:
			executor.RunInteractive(toStrings(commands))

		default:
			fmt.Println("Exiting.")
		}
	},
}

func printWorkflow(sessionID string, commands []models.CommandExecution) {

	fmt.Println()
	fmt.Println("Session Workflow")
	fmt.Println("────────────────")
	fmt.Println("Session ID:", sessionID)
	fmt.Println()

	for i, c := range commands {
		fmt.Printf("%2d. %s\n", i+1, c.Command)
	}

	fmt.Println()
}

func askExecutionMode() int {

	fmt.Println("Choose execution mode:")
	fmt.Println("1. Run all commands")
	fmt.Println("2. Step-by-step execution")
	fmt.Println("3. Interactive selection")
	fmt.Println("4. Exit")
	fmt.Println()

	fmt.Print("Enter choice: ")

	var choice int
	fmt.Scanln(&choice)

	return choice
}

func toStrings(commands []models.CommandExecution) []string {
	s := make([]string, len(commands))
	for i, c := range commands {
		s[i] = c.Command
	}
	return s
}
