package cmd

import (
	"os"

	"recall/cmd/ask"
	"recall/cmd/doctor"
	"recall/cmd/embed"
	"recall/cmd/query"
	"recall/cmd/record"
	"recall/cmd/session"
	"recall/cmd/setup"

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
	rootCmd.AddCommand(setup.GetInitCmd())
	rootCmd.AddCommand(setup.GetConfigCmd())

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
}

func Execute() error {
	rootCmd.SilenceUsage = true
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)

	return rootCmd.Execute()
}
