package commands

import (
	"fmt"
	"os"

	"github.com/quoctann/content-hub/cmd/app/bootstrap"
	"github.com/spf13/cobra"
)

var env string

var rootCmd = &cobra.Command{
	Use:   "api",
	Short: "Content Hub API Server",
	Long: `Content Hub API Server`,
	Run: run,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&env, "env", "e", "", "override APP_ENV (local|dev|prod)")
}

func run(cmd *cobra.Command, args []string) {
	// Panic recovery at the CLI level
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Panic recovered in CLI: %v\n", r)
			os.Exit(1)
		}
	}()

	// --env flag overrides APP_ENV env var so config.Load picks it up.
	if env != "" {
		os.Setenv("APP_ENV", env)
	}

	app, err := bootstrap.NewApp()
	if err != nil {
		fmt.Printf("Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	if err := app.Start(); err != nil {
		fmt.Printf("Application stopped with error: %v\n", err)
		os.Exit(1)
	}
}
