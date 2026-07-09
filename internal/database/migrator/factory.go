package migrator

import (
	"github.com/quoctann/content-hub/internal/domain"
)

func NewMigrator(databaseURL string, migrationsPath string) (domain.Migrator, error) {
	return NewGolangMigrateAdapter(databaseURL, migrationsPath)
}

func NewMigrationCreator() domain.MigrationCreator {
	return &FileBasedCreator{}
}
