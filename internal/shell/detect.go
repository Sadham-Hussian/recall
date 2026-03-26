package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Detect() (Shell, error) {
	name := detectParentProcess()

	switch name {
	case "zsh":
		return NewZshShell(), nil
	case "bash":
		return NewBashShell(), nil
	case "fish":
		return NewFishShell(), nil
	default:
		// fallback to $SHELL
		env := os.Getenv("SHELL")
		base := filepath.Base(env)

		switch base {
		case "zsh":
			return NewZshShell(), nil
		case "bash":
			return NewBashShell(), nil
		case "fish":
			return NewFishShell(), nil
		}
	}

	return nil, fmt.Errorf("unsupported shell")
}

func detectParentProcess() string {
	ppid := os.Getppid()

	out, err := exec.Command("ps", "-p", fmt.Sprintf("%d", ppid), "-o", "comm=").Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}
