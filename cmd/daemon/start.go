package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon service",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, _ := os.UserHomeDir()
		switch runtime.GOOS {
		case "darwin":
			if err := exec.Command("launchctl", "load", plistPath(home)).Run(); err != nil {
				return fmt.Errorf("launchctl load failed: %w", err)
			}
		case "linux":
			if err := exec.Command("systemctl", "--user", "start", "recall-daemon").Run(); err != nil {
				return fmt.Errorf("systemctl start failed: %w", err)
			}
		default:
			return fmt.Errorf("not supported on %s", runtime.GOOS)
		}
		fmt.Println("✔ Daemon started")
		return nil
	},
}
