package commands

import (
	"fmt"
	"log"
	"strconv"

	"github.com/quoctann/content-hub/cmd/db/bootstrap"
	"github.com/spf13/cobra"
)

var forceCmd = &cobra.Command{
	Use:   "force",
	Short: "Force migration to a specific version",
	Run: func(cmd *cobra.Command, args []string) {
		versionStr, _ := cmd.Flags().GetString("version")
		if versionStr == "" {
			log.Fatal("Version is required (--version)")
		}
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			log.Fatalf("Invalid version: %v", err)
		}

		m, err := bootstrap.InitMigrator(migrationsPath)
		if err != nil {
			log.Fatalf("Failed to initialize migrator: %v", err)
		}

		if err := m.Force(version); err != nil {
			log.Fatalf("Migration force failed: %v", err)
		}
		fmt.Printf("Migration forced to version %d successfully\n", version)
	},
}

func init() {
	forceCmd.Flags().StringP("version", "v", "", "Migration version")
}
