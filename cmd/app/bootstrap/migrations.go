package bootstrap

import (
	"context"
	"fmt"

	"github.com/quoctann/content-hub/internal/database/migrator"
	"github.com/quoctann/content-hub/pkg/database"
)

// RunMigrations checks the configuration and runs database migrations if enabled.
func RunMigrations(deps *Dependencies) error {
	if !deps.Config.Database.AutoMigrate {
		return nil
	}

	deps.Logger.InfoWithoutCtx("Running database migrations...")
	// golang-migrate creates schema_migrations as soon as it connects.
	if err := database.EnsureSchema(context.Background(), deps.Config.Database); err != nil {
		return err
	}

	migrationsPath := deps.Config.Database.MigrationsPath
	if migrationsPath == "" {
		migrationsPath = "migrations"
	}

	// Migrator handles the "file://" prefix logic internally
	m, err := migrator.NewMigrator(database.URL(deps.Config.Database), migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	if err := m.Up(); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	deps.Logger.InfoWithoutCtx("Database migrations applied successfully")
	return nil
}
