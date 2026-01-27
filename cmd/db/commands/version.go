package commands

import (
	"fmt"
	"log"

	"github.com/quoctann/content-hub/cmd/db/bootstrap"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Check current migration version",
	Run: func(cmd *cobra.Command, args []string) {
		m, err := bootstrap.InitMigrator(env, migrationsPath)
		if err != nil {
			log.Fatalf("Failed to initialize migrator: %v", err)
		}

		v, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Failed to get version: %v", err)
		}
		fmt.Printf("Current version: %d (dirty: %v)\n", v, dirty)
	},
}
