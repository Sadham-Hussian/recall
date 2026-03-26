package setup

import (
	"fmt"
	"log"

	"recall/internal/config"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Initialize recall config",
	Run: func(cmd *cobra.Command, args []string) {

		path, err := config.EnsureConfigFile()
		if err != nil {
			log.Fatalf("failed to initialize config: %v", err)
		}

		fmt.Println("Config ready at:", path)
	},
}

func GetConfigCmd() *cobra.Command {
	return configCmd
}
