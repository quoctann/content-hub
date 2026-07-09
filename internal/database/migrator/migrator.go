package migrator

import (
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// GolangMigrateAdapter implements the domain.Migrator interface
// using the github.com/golang-migrate/migrate library.
type GolangMigrateAdapter struct {
	m *migrate.Migrate
}

func NewGolangMigrateAdapter(databaseURL string, migrationsPath string) (*GolangMigrateAdapter, error) {
	sourceURL := migrationsPath
	if !strings.HasPrefix(migrationsPath, "file://") {
		sourceURL = "file://" + migrationsPath
	}

	m, err := migrate.New(
		sourceURL,
		databaseURL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migrator: %w", err)
	}
	return &GolangMigrateAdapter{m: m}, nil
}

func (mig *GolangMigrateAdapter) Up() error {
	if err := mig.m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func (mig *GolangMigrateAdapter) Down() error {
	if err := mig.m.Steps(-1); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func (mig *GolangMigrateAdapter) Force(version int) error {
	return mig.m.Force(version)
}

func (mig *GolangMigrateAdapter) Version() (uint, bool, error) {
	return mig.m.Version()
}
