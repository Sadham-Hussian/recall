package setup

import (
	"fmt"
	"log"

	"recall/internal/storage"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := storage.NewDB()
		if err != nil {
			log.Fatal(err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			log.Fatal(err)
		}
		defer sqlDB.Close()

		fmt.Println("Migrations applied successfully.")
	},
}

func GetMigrateCmd() *cobra.Command {
	return migrateCmd
}
