package workflow

import (
	workflow_svc "recall/internal/services/workflow"
	"strings"

	"github.com/spf13/cobra"
)

// workflowNameCompletion returns the list of saved workflow names for
// shell completion, filtered by the partial token the user has typed.
// Any error (no DB yet, config missing, etc.) silently yields no
// suggestions — completion must never fail the shell.
func workflowNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Only complete the first positional arg.
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	svc, err := workflow_svc.NewWorkflowService()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	workflows, err := svc.List()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	out := make([]string, 0, len(workflows))
	for _, w := range workflows {
		if toComplete == "" || strings.HasPrefix(w.Name, toComplete) {
			// Cobra's zsh generator uses the first ":" to separate the
			// value from its description. Escape any ":" inside Name
			// or Description so descriptions render correctly.
			desc := strings.ReplaceAll(w.Description, ":", "\\:")
			name := strings.ReplaceAll(w.Name, ":", "\\:")
			if desc != "" {
				out = append(out, name+"\t"+desc)
			} else {
				out = append(out, name)
			}
		}
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}
