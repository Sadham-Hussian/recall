package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the daemon service",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, _ := os.UserHomeDir()
		switch runtime.GOOS {
		case "darwin":
			if err := exec.Command("launchctl", "unload", plistPath(home)).Run(); err != nil {
				return fmt.Errorf("launchctl unload failed: %w", err)
			}
		case "linux":
			if err := exec.Command("systemctl", "--user", "stop", "recall-daemon").Run(); err != nil {
				return fmt.Errorf("systemctl stop failed: %w", err)
			}
		default:
			return fmt.Errorf("not supported on %s", runtime.GOOS)
		}
		fmt.Println("✔ Daemon stopped")
		return nil
	},
}
