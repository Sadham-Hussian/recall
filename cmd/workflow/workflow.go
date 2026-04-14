package workflow

import (
	"fmt"
	"log"

	"recall/internal/config"
	"recall/internal/executor"
	workflow_svc "recall/internal/services/workflow"

	"github.com/spf13/cobra"
)

var (
	fromSession string
	description string
)

var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Manage reusable command workflows",
}

var runWorkflowCmd = &cobra.Command{
	Use:               "run <name>",
	Short:             "Execute a saved workflow",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: workflowNameCompletion,
	Run: func(cmd *cobra.Command, args []string) {
		config.LoadConfig()
		svc, err := workflow_svc.NewWorkflowService()
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		w, steps, err := svc.Show(args[0])
		if err != nil {
			log.Fatalf("workflow not found: %v", err)
		}

		// Display
		fmt.Println()
		fmt.Printf("Workflow: %s\n", w.Name)
		fmt.Println("────────────────")
		cmds := make([]string, len(steps))
		for i, s := range steps {
			cmds[i] = s.Command
			fmt.Printf("%2d. %s\n", i+1, s.Command)
		}
		fmt.Println()

		mode := executor.AskExecutionMode()
		switch mode {
		case 1:
			executor.RunAll(cmds)
		case 2:
			executor.RunStepByStep(cmds)
		case 3:
			executor.RunInteractive(cmds)
		default:
			fmt.Println("Exiting.")
		}
	},
}

func GetWorkflowCmd() *cobra.Command {
	saveCmd.Flags().StringVar(&fromSession, "from-session", "", "session ID to pick commands from")
	saveCmd.Flags().StringVar(&description, "description", "", "workflow description")

	workflowCmd.AddCommand(saveCmd)
	workflowCmd.AddCommand(listCmd)
	workflowCmd.AddCommand(showCmd)
	workflowCmd.AddCommand(runWorkflowCmd)
	workflowCmd.AddCommand(deleteCmd)
	return workflowCmd
}
