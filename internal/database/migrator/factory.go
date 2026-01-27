package migrator

import (
	"github.com/quoctann/content-hub/internal/domain"
)

// NewMigrator returns a domain.Migrator implementation.
// Currently it wraps the GolangMigrateAdapter.
func NewMigrator(databaseURL string, migrationsPath string) (domain.Migrator, error) {
	// Use the GolangMigrateAdapter as the concrete implementation.
	// If we later switch to another library, only this function changes.
	return NewGolangMigrateAdapter(databaseURL, migrationsPath)
}

// NewMigrationCreator returns a domain.MigrationCreator implementation.
// It uses the file‑based creator defined in creator.go.
func NewMigrationCreator() domain.MigrationCreator {
	return &FileBasedCreator{}
}
