package commands

import (
	"fmt"
	"os"

	"github.com/quoctann/content-hub/cmd/app/bootstrap"
	"github.com/spf13/cobra"
)

var env string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "api",
	Short: "Content Hub API Server",
	Long: `Content Hub API Server follows clean architecture principles.
It provides RESTful endpoints for managing content and users.`,
	Run: run,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&env, "env", "e", "", "environment (local|dev|prod)")
	if env == "" {
		env = os.Getenv("APP_ENV")
	}
}

func run(cmd *cobra.Command, args []string) {
	// Panic recovery at the CLI level
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Panic recovered in CLI: %v\n", r)
			os.Exit(1)
		}
	}()

	// Initialize and start the application
	app, err := bootstrap.NewApp(env)
	if err != nil {
		fmt.Printf("Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	if err := app.Start(); err != nil {
		fmt.Printf("Application stopped with error: %v\n", err)
		os.Exit(1)
	}
}
