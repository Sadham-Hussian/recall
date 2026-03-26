package setup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Show instructions to enable recall shell integration",
	Run: func(cmd *cobra.Command, args []string) {

		shell := detectShell()

		home, _ := os.UserHomeDir()

		var rcFile string
		var command string

		switch shell {

		case "zsh":
			rcFile = filepath.Join(home, ".zshrc")
			command = `eval "$(recall init zsh)"`

		case "bash":
			rcFile = filepath.Join(home, ".bashrc")
			command = `eval "$(recall init bash)"`

		case "fish":
			rcFile = filepath.Join(home, ".config/fish/config.fish")
			command = `recall init fish | source`

		default:
			fmt.Println("Could not detect your shell.")
			fmt.Println("Supported shells: zsh, bash, fish")
			return
		}

		fmt.Println()
		fmt.Println("Recall Shell Integration")
		fmt.Println("────────────────────────")
		fmt.Println()

		fmt.Println("Detected shell:", shell)
		fmt.Println()

		fmt.Printf("1. Open your shell config file:\n\n")
		fmt.Printf("   %s\n\n", rcFile)

		fmt.Println("2. Add the following line:")
		fmt.Println()
		fmt.Printf("   %s\n\n", command)

		fmt.Println("3. Reload your shell:")
		fmt.Println()

		fmt.Printf("   source %s\n", rcFile)

		fmt.Println()
		fmt.Println("Recall will now automatically record your commands.")
		fmt.Println()
	},
}

func GetInstallCmd() *cobra.Command {
	return installCmd
}

func detectShell() string {

	shell := filepath.Base(os.Getenv("SHELL"))

	switch shell {

	case "zsh":
		return "zsh"

	case "bash":
		return "bash"

	case "fish":
		return "fish"

	default:
		return "unknown"
	}
}
