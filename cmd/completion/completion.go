package completion

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate shell completion script",
	Long: `Generate the autocompletion script for recall for the specified shell.

To load completions:

Bash:
  # Requires bash-completion: apt install bash-completion (if not installed)
  $ mkdir -p ~/.local/share/bash-completion/completions
  $ recall completion bash > ~/.local/share/bash-completion/completions/recall

Zsh:
  # Ensure $fpath includes a user-writable completions dir, then:
  $ recall completion zsh > "${fpath[1]}/_recall"
  # Restart your shell (or run: autoload -U compinit && compinit)

Fish:
  $ recall completion fish > ~/.config/fish/completions/recall.fish
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletionV2(os.Stdout, true)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		}
		return nil
	},
}

func GetCompletionCmd() *cobra.Command {
	return completionCmd
}
