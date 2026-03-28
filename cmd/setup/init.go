package setup

import (
	"fmt"
	"recall/internal/config"
	"recall/internal/storage"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Recall (config, database, migrations)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Initializing Recall...")

		// 1. Load or create config
		config.LoadConfig()
		fmt.Println("✔ Config loaded")

		// 2. Resolve DB path
		db, err := storage.NewDB()
		if err != nil {
			return fmt.Errorf("failed to create db: %w", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("failed to get db: %w", err)
		}
		sqlDB.Close()

		fmt.Println("✔ Database connected")

		// 7. Done + next steps
		fmt.Println("\nRecall initialized successfully 🚀")

		fmt.Println("Next steps:")

		fmt.Println("1. Enable shell integration:")
		fmt.Println(`   eval "$(recall hook zsh)"`)
		fmt.Println()

		fmt.Println("2. (Optional) Import history:")
		fmt.Println("   recall history")
		fmt.Println()

		fmt.Println("3. (Optional) Generate embeddings:")
		fmt.Println("   recall embed")
		fmt.Println()

		fmt.Println("4. Try it:")
		fmt.Println(`   recall ask "find docker command"`)

		return nil
	},
}

func GetInitCmd() *cobra.Command {
	return initCmd
}
