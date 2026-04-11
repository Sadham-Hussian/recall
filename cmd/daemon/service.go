package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/spf13/cobra"
)

// ── Templates ────────────────────────────────────────────────────────────────

const launchdPlistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.recall.daemon</string>
	<key>ProgramArguments</key>
	<array>
		<string>{{.BinaryPath}}</string>
		<string>daemon</string>
		<string>run</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
</dict>
</plist>
`

const systemdServiceTemplate = `[Unit]
Description=Recall Embedding Daemon
After=default.target

[Service]
ExecStart={{.BinaryPath}} daemon run
Restart=on-failure
RestartSec=10

[Install]
WantedBy=default.target
`

type serviceTemplateData struct {
	BinaryPath string
}

// ── Shared helpers ────────────────────────────────────────────────────────────

func resolveBinaryPath() (string, error) {
	p, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not determine binary path: %w", err)
	}
	p, err = filepath.EvalSymlinks(p)
	if err != nil {
		return "", fmt.Errorf("could not resolve binary symlink: %w", err)
	}
	return p, nil
}

func plistPath(home string) string {
	return filepath.Join(home, "Library", "LaunchAgents", "com.recall.daemon.plist")
}

func systemdServicePath(home string) string {
	return filepath.Join(home, ".config", "systemd", "user", "recall-daemon.service")
}

// ── install ───────────────────────────────────────────────────────────────────

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install and start the daemon as a system service",
	RunE: func(cmd *cobra.Command, args []string) error {
		bin, err := resolveBinaryPath()
		if err != nil {
			return err
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		data := serviceTemplateData{BinaryPath: bin}

		switch runtime.GOOS {
		case "darwin":
			return launchdInstall(home, data)
		case "linux":
			return systemdInstall(home, data)
		default:
			return fmt.Errorf("not supported on %s — run 'recall daemon run' manually", runtime.GOOS)
		}
	},
}

func launchdInstall(home string, data serviceTemplateData) error {
	dir := filepath.Join(home, "Library", "LaunchAgents")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := plistPath(home)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := template.Must(template.New("p").Parse(launchdPlistTemplate)).Execute(f, data); err != nil {
		return err
	}
	exec.Command("launchctl", "unload", path).Run() // ignore error — may not exist yet
	if err := exec.Command("launchctl", "load", path).Run(); err != nil {
		return fmt.Errorf("launchctl load failed: %w", err)
	}
	fmt.Printf("✔ Installed: %s\n", path)
	fmt.Printf("  Log:       %s\n", filepath.Join(home, ".recall", "daemon.log"))
	return nil
}

func systemdInstall(home string, data serviceTemplateData) error {
	dir := filepath.Join(home, ".config", "systemd", "user")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := systemdServicePath(home)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := template.Must(template.New("s").Parse(systemdServiceTemplate)).Execute(f, data); err != nil {
		return err
	}
	exec.Command("systemctl", "--user", "daemon-reload").Run()
	if err := exec.Command("systemctl", "--user", "enable", "--now", "recall-daemon").Run(); err != nil {
		return fmt.Errorf("systemctl enable failed: %w", err)
	}
	fmt.Printf("✔ Installed: %s\n", path)
	fmt.Printf("  Log:       %s\n", filepath.Join(home, ".recall", "daemon.log"))
	return nil
}
