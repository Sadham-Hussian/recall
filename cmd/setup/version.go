package setup

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print recall version",
	Run: func(cmd *cobra.Command, args []string) {
		if Version == "dev" {
			fmt.Fprintln(cmd.OutOrStdout(), "recall dev")
			return
		}
		fmt.Fprintf(cmd.OutOrStdout(), "recall %s (commit %s, built %s)\n", Version, Commit, Date)
	},
}

func GetVersionCmd() *cobra.Command {
	return versionCmd
}
