package cmd

import (
	"os"

	"recall/cmd/ask"
	"recall/cmd/completion"
	"recall/cmd/daemon"
	"recall/cmd/doctor"
	"recall/cmd/embed"
	"recall/cmd/query"
	"recall/cmd/record"
	"recall/cmd/session"
	"recall/cmd/setup"
	"recall/cmd/stats"
	"recall/cmd/upgrade"
	"recall/cmd/workflow"
	"recall/internal/config"
	upgradesvc "recall/internal/services/upgrade"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "recall",
	Short: "Recall is a CLI for memory and notes workflows",
	Long:  "Recall is a Cobra-based CLI that will host commands for memory and notes workflows.",
}

func init() {
	rootCmd.AddCommand(setup.GetInstallCmd())
	rootCmd.AddCommand(setup.GetUninstallCmd())
	rootCmd.AddCommand(setup.GetMigrateCmd())
	rootCmd.AddCommand(setup.GetVersionCmd())
	rootCmd.AddCommand(setup.GetHookCmd())
	rootCmd.AddCommand(setup.GetConfigCmd())
	rootCmd.AddCommand(setup.GetInitCmd())

	rootCmd.AddCommand(record.GetHistoryCmd())
	rootCmd.AddCommand(record.GetRecordCmd())

	rootCmd.AddCommand(query.GetLastCmd())
	rootCmd.AddCommand(query.GetListCmd())
	rootCmd.AddCommand(query.GetSearchCmd())
	rootCmd.AddCommand(query.GetSuggestCmd())

	rootCmd.AddCommand(session.GetSessionCmd())
	rootCmd.AddCommand(session.GetContinueCmd())

	rootCmd.AddCommand(embed.GetEmbedCmd())

	rootCmd.AddCommand(ask.GetAskCmd())

	rootCmd.AddCommand(doctor.GetDoctorCmd())

	rootCmd.AddCommand(daemon.GetDaemonCmd())

	rootCmd.AddCommand(workflow.GetWorkflowCmd())

	rootCmd.AddCommand(upgrade.GetUpgradeCmd())

	// outputSensitiveCommands emit shell script to stdout that the user's shell
	// eval's. They must produce no incidental output and must not trigger
	// side effects like creating the default config file.
	var outputSensitiveCommands = map[string]bool{
		"hook":    true,
		"install": true,
	}

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if outputSensitiveCommands[topLevelName(cmd)] {
			return nil
		}

		cfg := config.LoadConfig()
		upgradesvc.MaybeCheckInBackground(cfg)
		return nil
	}
	rootCmd.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		if outputSensitiveCommands[topLevelName(cmd)] {
			return nil
		}
		upgradesvc.PrintNoticeIfAvailable(config.AppConfig, setup.Version, topLevelName(cmd))
		return nil
	}

	rootCmd.AddCommand(completion.GetCompletionCmd())

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(stats.GetStatsCmd())

}

func Execute() error {
	rootCmd.SilenceUsage = true
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)

	return rootCmd.Execute()
}

// topLevelName returns the direct child of rootCmd in cmd's ancestor chain,
// e.g. "daemon" for `recall daemon start`. If cmd is already a top-level
// command, its own name is returned. Returns "" if cmd is the root.
func topLevelName(cmd *cobra.Command) string {
	for cmd != nil && cmd.Parent() != nil && cmd.Parent().Parent() != nil {
		cmd = cmd.Parent()
	}
	if cmd == nil || cmd.Parent() == nil {
		return ""
	}
	return cmd.Name()
}
