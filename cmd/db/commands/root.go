package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	env            string
	migrationsPath string
)

// rootCmd represents the base command for database operations
var rootCmd = &cobra.Command{
	Use:   "database",
	Short: "Database migration CLI",
	Long:  `CLI tool for managing database migrations using the domain interfaces and clean architecture.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&env, "env", "e", "local", "environment (local|dev|prod)")
	rootCmd.PersistentFlags().StringVarP(&migrationsPath, "path", "p", "migrations", "path to migrations folder")

	// Register subcommands
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(forceCmd)
	rootCmd.AddCommand(versionCmd)
}
