package commands

import (
	"log"

	"github.com/quoctann/content-hub/cmd/db/bootstrap"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new migration file",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			log.Fatal("Migration name is required (--name)")
		}

		creator := bootstrap.InitMigrationCreator()
		if err := creator.CreateMigration(migrationsPath, name); err != nil {
			log.Fatalf("Failed to create migration: %v", err)
		}
	},
}

func init() {
	createCmd.Flags().StringP("name", "n", "", "Migration name")
}
