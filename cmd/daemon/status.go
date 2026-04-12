package daemon

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show daemon service status",
	RunE: func(cmd *cobra.Command, args []string) error {
		switch runtime.GOOS {
		case "darwin":
			out, err := exec.Command("launchctl", "list", "com.recall.daemon").CombinedOutput()
			if err != nil {
				fmt.Println("Daemon is not running (not found in launchctl)")
				return nil
			}
			fmt.Print(string(out))
		case "linux":
			out, _ := exec.Command("systemctl", "--user", "status", "recall-daemon").CombinedOutput()
			fmt.Print(string(out))
		default:
			return fmt.Errorf("not supported on %s", runtime.GOOS)
		}
		return nil
	},
}
