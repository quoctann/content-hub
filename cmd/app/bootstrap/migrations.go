package bootstrap

import (
	"fmt"

	"github.com/quoctann/content-hub/internal/database/migrator"
)

// RunMigrations checks the configuration and runs database migrations if enabled.
func RunMigrations(deps *Dependencies) error {
	if !deps.Config.Database.AutoMigrate {
		return nil
	}

	deps.Logger.InfoWithoutCtx("Running database migrations...")
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		deps.Config.Database.User,
		deps.Config.Database.Password,
		deps.Config.Database.Host,
		deps.Config.Database.Port,
		deps.Config.Database.Name,
		deps.Config.Database.SSLMode,
	)

	if deps.Config.Database.Schema != "" {
		dbURL = fmt.Sprintf("%s&search_path=%s", dbURL, deps.Config.Database.Schema)
	}

	migrationsPath := deps.Config.Database.MigrationsPath
	if migrationsPath == "" {
		migrationsPath = "migrations"
	}

	// Migrator handles the "file://" prefix logic internally
	m, err := migrator.NewMigrator(dbURL, migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	if err := m.Up(); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	deps.Logger.InfoWithoutCtx("Database migrations applied successfully")
	return nil
}
