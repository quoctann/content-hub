package commands

import (
	"fmt"
	"log"

	"github.com/quoctann/content-hub/cmd/db/bootstrap"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Run migrations down",
	Run: func(cmd *cobra.Command, args []string) {
		m, err := bootstrap.InitMigrator(env, migrationsPath)
		if err != nil {
			log.Fatalf("Failed to initialize migrator: %v", err)
		}

		if err := m.Down(); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		fmt.Println("Migration down completed successfully")
	},
}
