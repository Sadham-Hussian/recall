package export

import (
	"fmt"
	"os"
	"time"

	"recall/internal/config"
	exportsvc "recall/internal/services/export"

	"github.com/spf13/cobra"
)

var (
	outputFile string
	days       int
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export recall data as JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		config.LoadConfig()

		svc, err := exportsvc.NewExportService()
		if err != nil {
			return err
		}

		var sinceTs int64
		if days > 0 {
			sinceTs = time.Now().AddDate(0, 0, -days).Unix()
		}

		var w *os.File
		if outputFile != "" {
			w, err = os.Create(outputFile)
			if err != nil {
				return fmt.Errorf("create output file: %w", err)
			}
			defer w.Close()
		} else {
			w = os.Stdout
		}

		if err := svc.Export(w, sinceTs); err != nil {
			return err
		}

		if outputFile != "" {
			fmt.Fprintf(os.Stderr, "✔ Exported to %s\n", outputFile)
		}
		return nil
	},
}

func GetExportCmd() *cobra.Command {
	exportCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file path (default: stdout)")
	exportCmd.Flags().IntVar(&days, "days", 0, "export only last N days of commands")
	return exportCmd
}
