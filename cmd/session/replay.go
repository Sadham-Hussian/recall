package session

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"recall/internal/config"
	"recall/internal/services/session"
	"recall/internal/storage/models"

	"github.com/manifoldco/promptui"
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
			runAll(commands)

		case 2:
			runStepByStep(commands)

		case 3:
			runInteractive(commands)

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

func runAll(commands []models.CommandExecution) {

	for _, c := range commands {

		fmt.Println()
		fmt.Println("Executing:", c.Command)

		err := runCommand(c.Command)
		if err != nil {
			fmt.Println("Command failed:", err)
			return
		}
	}
}

func runStepByStep(commands []models.CommandExecution) {

	for _, c := range commands {

		fmt.Printf("\nRun command: %s ? (y/n/exit): ", c.Command)

		var input string
		fmt.Scanln(&input)

		switch input {

		case "y":
			fmt.Println("Executing:", c.Command)
			err := runCommand(c.Command)
			if err != nil {
				fmt.Println("Command failed:", err)
			}

		case "exit":
			fmt.Println("Exiting replay.")
			return
		}
	}
}

func runInteractive(commands []models.CommandExecution) {

	items := make([]string, len(commands))

	for i, c := range commands {
		items[i] = c.Command
	}

	for {

		prompt := promptui.Select{
			Label: "Select command to execute (Ctrl+C to exit)",
			Items: items,
			Size:  10,
		}

		index, _, err := prompt.Run()
		if err != nil {
			fmt.Println("Exiting interactive mode.")
			return
		}

		command := items[index]

		fmt.Println()
		fmt.Println("Executing:", command)

		err = runCommand(command)
		if err != nil {
			fmt.Println("Command failed:", err)
		}
	}
}

func runCommand(command string) error {

	cmd := exec.Command("sh", "-c", command)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
