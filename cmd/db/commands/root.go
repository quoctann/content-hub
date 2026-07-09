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

var rootCmd = &cobra.Command{
	Use:   "database",
	Short: "Database migration CLI",
	Long:  `CLI tool for managing database migrations using the domain interfaces and clean architecture.`,
	// PersistentPreRun applies the optional --env flag before any subcommand runs,
	// so config.Load picks up the overridden APP_ENV value.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if env != "" {
			_ = os.Setenv("APP_ENV", env)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&env, "env", "e", "", "override APP_ENV (local|dev|prod)")
	rootCmd.PersistentFlags().StringVarP(&migrationsPath, "path", "p", "", "path to migrations folder (default: DATABASE_MIGRATIONS_PATH)")

	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(forceCmd)
	rootCmd.AddCommand(versionCmd)
}
