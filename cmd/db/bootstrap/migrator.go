package bootstrap

import (
	"context"

	"github.com/quoctann/content-hub/cmd/shared"
	"github.com/quoctann/content-hub/internal/database/migrator"
	"github.com/quoctann/content-hub/internal/domain"
	"github.com/quoctann/content-hub/pkg/database"
)

func InitMigrator(migrationsPath string) (domain.Migrator, error) {
	cfg, err := shared.LoadConfig()
	if err != nil {
		return nil, err
	}

	if migrationsPath == "" {
		migrationsPath = cfg.Database.MigrationsPath
	}

	// golang-migrate creates schema_migrations as soon as it connects.
	if err := database.EnsureSchema(context.Background(), cfg.Database); err != nil {
		return nil, err
	}

	return migrator.NewMigrator(database.URL(cfg.Database), migrationsPath)
}

func InitMigrationCreator() domain.MigrationCreator {
	return migrator.NewMigrationCreator()
}
