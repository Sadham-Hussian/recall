package setup

import (
	"fmt"

	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Show instructions to remove recall shell integration",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println()
		fmt.Println("Recall Uninstall Instructions")
		fmt.Println("─────────────────────────────")
		fmt.Println()

		fmt.Println("1. Open your shell config file (e.g. ~/.zshrc or ~/.bashrc)")
		fmt.Println()

		fmt.Println("2. Remove the following line:")
		fmt.Println()

		fmt.Println(`   eval "$(recall init zsh)"`)
		fmt.Println(`   eval "$(recall init bash)"`)
		fmt.Println(`   recall init fish | source`)
		fmt.Println()

		fmt.Println("3. Restart your terminal or reload your shell.")
		fmt.Println()
	},
}

func GetUninstallCmd() *cobra.Command {
	return uninstallCmd
}
