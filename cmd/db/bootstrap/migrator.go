package bootstrap

import (
	"fmt"

	"github.com/quoctann/content-hub/cmd/shared"
	"github.com/quoctann/content-hub/internal/database/migrator"
	"github.com/quoctann/content-hub/internal/domain"
)

// InitMigrator initializes a domain.Migrator using the shared config.
func InitMigrator(env, migrationsPath string) (domain.Migrator, error) {
	cfg, err := shared.LoadConfig(env)
	if err != nil {
		return nil, err
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	if cfg.Database.Schema != "" {
		dbURL = fmt.Sprintf("%s&search_path=%s", dbURL, cfg.Database.Schema)
	}

	return migrator.NewMigrator(dbURL, migrationsPath)
}

// InitMigrationCreator returns an implementation for creating migration files.
func InitMigrationCreator() domain.MigrationCreator {
	return migrator.NewMigrationCreator()
}
