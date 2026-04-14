package workflow

import (
	"fmt"
	"log"
	"recall/internal/config"
	workflow_svc "recall/internal/services/workflow"

	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:               "show <name>",
	Short:             "Show steps in a workflow",
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
		fmt.Println()
		fmt.Printf("Workflow: %s\n", w.Name)
		if w.Description != "" {
			fmt.Printf("Description: %s\n", w.Description)
		}
		fmt.Println("────────────────")
		for _, s := range steps {
			fmt.Printf("%2d. %s\n", s.StepOrder, s.Command)
		}
		fmt.Println()
	},
}
