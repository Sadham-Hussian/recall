package upgrade

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"recall/cmd/setup"
	upgradesvc "recall/internal/services/upgrade"

	"github.com/spf13/cobra"
)

var (
	checkOnly bool
	yes       bool
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade recall to the latest release",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Minute)
		defer cancel()

		release, err := upgradesvc.LatestRelease(ctx)
		if err != nil {
			return fmt.Errorf("fetch latest release: %w", err)
		}

		current := setup.Version
		if !upgradesvc.IsNewer(current, release.TagName) {
			fmt.Printf("recall %s is already the latest.\n", current)
			return nil
		}

		fmt.Printf("Current: %s\nLatest:  %s\n", current, release.TagName)
		if checkOnly {
			fmt.Println("Run `recall upgrade` to install.")
			return nil
		}

		if !yes && !promptYesNo("Upgrade now? [Y/n]: ") {
			fmt.Println("Cancelled.")
			return nil
		}

		fmt.Println("Downloading…")
		tarball, checksums, err := upgradesvc.Download(ctx, release)
		if err != nil {
			return fmt.Errorf("download: %w", err)
		}
		defer os.Remove(tarball)
		defer os.Remove(checksums)

		fmt.Println("Verifying checksum…")
		if err := upgradesvc.VerifyChecksum(tarball, checksums); err != nil {
			return fmt.Errorf("checksum: %w", err)
		}

		tmpDir, err := os.MkdirTemp("", "recall-upgrade-*")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDir)

		binary, err := upgradesvc.Extract(tarball, tmpDir)
		if err != nil {
			return fmt.Errorf("extract: %w", err)
		}

		if err := upgradesvc.SwapBinary(binary); err != nil {
			return fmt.Errorf("install: %w", err)
		}

		fmt.Printf("✔ Upgraded to %s\n", release.TagName)
		return nil
	},
}

func promptYesNo(prompt string) bool {
	fmt.Print(prompt)
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		return false
	}
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "" || line == "y" || line == "yes"
}

func GetUpgradeCmd() *cobra.Command {
	upgradeCmd.Flags().BoolVar(&checkOnly, "check", false, "check without installing")
	upgradeCmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation")
	return upgradeCmd
}
