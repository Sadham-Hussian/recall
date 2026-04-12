package workflow

import (
	"fmt"
	"log"
	"recall/internal/config"
	workflow_svc "recall/internal/services/workflow"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all workflows",
	Run: func(cmd *cobra.Command, args []string) {
		config.LoadConfig()
		svc, err := workflow_svc.NewWorkflowService()
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		workflows, err := svc.List()
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		if len(workflows) == 0 {
			fmt.Println("No workflows saved.")
			return
		}
		for _, w := range workflows {
			fmt.Printf("  %s", w.Name)
			if w.Description != "" {
				fmt.Printf("  — %s", w.Description)
			}
			fmt.Println()
		}
	},
}
