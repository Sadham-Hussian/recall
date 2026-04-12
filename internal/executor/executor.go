package executor

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/manifoldco/promptui"
)

func AskExecutionMode() int {
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

func RunAll(commands []string) {
	for _, c := range commands {
		fmt.Println()
		fmt.Println("Executing:", c)
		if err := RunCommand(c); err != nil {
			fmt.Println("Command failed:", err)
			return
		}
	}
}

func RunStepByStep(commands []string) {
	for _, c := range commands {
		fmt.Printf("\nRun command: %s ? (y/n/exit): ", c)
		var input string
		fmt.Scanln(&input)
		switch input {
		case "y":
			fmt.Println("Executing:", c)
			if err := RunCommand(c); err != nil {
				fmt.Println("Command failed:", err)
			}
		case "exit":
			fmt.Println("Exiting.")
			return
		}
	}
}

func RunInteractive(commands []string) {
	for {
		prompt := promptui.Select{
			Label: "Select command to execute (Ctrl+C to exit)",
			Items: commands,
			Size:  10,
		}
		index, _, err := prompt.Run()
		if err != nil {
			fmt.Println("Exiting interactive mode.")
			return
		}
		fmt.Println()
		fmt.Println("Executing:", commands[index])
		if err := RunCommand(commands[index]); err != nil {
			fmt.Println("Command failed:", err)
		}
	}
}

func RunCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
