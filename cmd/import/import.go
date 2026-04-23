package importcmd

import (
	"fmt"
	"os"

	"recall/internal/config"
	exportsvc "recall/internal/services/export"

	"github.com/spf13/cobra"
)

var replace bool

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import recall data from a JSON export",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config.LoadConfig()

		filePath := args[0]
		f, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("open file: %w", err)
		}
		defer f.Close()

		svc, err := exportsvc.NewExportService()
		if err != nil {
			return err
		}

		if replace {
			fmt.Println("⚠ Replace mode: existing data will be wiped before import.")
			fmt.Print("Continue? [y/N]: ")
			var input string
			fmt.Scanln(&input)
			if input != "y" && input != "Y" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		result, err := svc.Import(f, replace)
		if err != nil {
			return err
		}

		fmt.Printf("✔ %s\n", result)
		return nil
	},
}

func GetImportCmd() *cobra.Command {
	importCmd.Flags().BoolVar(&replace, "replace", false, "wipe existing data before importing")
	return importCmd
}
