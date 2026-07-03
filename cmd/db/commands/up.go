package commands

import (
	"fmt"
	"log"

	"github.com/quoctann/content-hub/cmd/db/bootstrap"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Run migrations up",
	Run: func(cmd *cobra.Command, args []string) {
		m, err := bootstrap.InitMigrator(migrationsPath)
		if err != nil {
			log.Fatalf("Failed to initialize migrator: %v", err)
		}

		if err := m.Up(); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
		fmt.Println("Migration up completed successfully")
	},
}
