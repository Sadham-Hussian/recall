package workflow

import (
	"bufio"
	"fmt"
	"os"
	"recall/internal/config"
	workflow_svc "recall/internal/services/workflow"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:   "save <name>",
	Short: "Save a new workflow",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config.LoadConfig()
		name := args[0]
		svc, err := workflow_svc.NewWorkflowService()
		if err != nil {
			return err
		}

		var selectedCmds []string

		if fromSession != "" {
			// From session: show commands, let user pick by number
			cmds, err := svc.GetSessionCommands(fromSession)
			if err != nil {
				return err
			}

			fmt.Printf("\nSession: %s\n", fromSession)
			for i, c := range cmds {
				fmt.Printf("  %d. %s\n", i+1, c)
			}
			fmt.Println()
			fmt.Print("Select commands (e.g. 1,3,4): ")

			var input string
			fmt.Scanln(&input)

			parts := strings.Split(input, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				idx, err := strconv.Atoi(p)
				if err != nil || idx < 1 || idx > len(cmds) {
					return fmt.Errorf("invalid selection: %s", p)
				}
				selectedCmds = append(selectedCmds, cmds[idx-1])
			}
		} else {
			// Interactive: type commands one by one
			fmt.Println("\nEnter commands (type 'save' to finish):")
			reader := bufio.NewReader(os.Stdin)
			step := 1
			for {
				fmt.Printf("  %d: ", step)
				line, _ := reader.ReadString('\n')
				line = strings.TrimSpace(line)
				if line == "save" {
					break
				}
				if line == "" {
					continue
				}
				selectedCmds = append(selectedCmds, line)
				step++
			}
		}

		if len(selectedCmds) == 0 {
			return fmt.Errorf("no commands selected")
		}

		if err := svc.SaveFromCommands(name, description, selectedCmds); err != nil {
			return err
		}
		fmt.Printf("\n✔ Workflow '%s' saved (%d steps)\n", name, len(selectedCmds))
		return nil
	},
}
