package bootstrap

import (
	"fmt"

	"github.com/quoctann/content-hub/cmd/shared"
	"github.com/quoctann/content-hub/internal/database/migrator"
	"github.com/quoctann/content-hub/internal/domain"
)

func InitMigrator(migrationsPath string) (domain.Migrator, error) {
	cfg, err := shared.LoadConfig()
	if err != nil {
		return nil, err
	}

	if migrationsPath == "" {
		migrationsPath = cfg.Database.MigrationsPath
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

func InitMigrationCreator() domain.MigrationCreator {
	return migrator.NewMigrationCreator()
}
