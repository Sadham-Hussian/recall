package workflow

import (
	"fmt"
	"recall/internal/config"
	workflow_svc "recall/internal/services/workflow"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a workflow",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config.LoadConfig()
		svc, err := workflow_svc.NewWorkflowService()
		if err != nil {
			return err
		}
		if err := svc.Delete(args[0]); err != nil {
			return err
		}
		fmt.Printf("✔ Workflow '%s' deleted\n", args[0])
		return nil
	},
}
