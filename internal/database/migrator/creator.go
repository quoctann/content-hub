package migrator

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileBasedCreator implements domain.MigrationCreator for local file system.
type FileBasedCreator struct{}

// CreateMigration creates a pair of up/down migration files.
func (c *FileBasedCreator) CreateMigration(migrationsPath, name string) error {
	timestamp := time.Now().Unix()
	upFile := filepath.Join(migrationsPath, fmt.Sprintf("%d_%s.up.sql", timestamp, name))
	downFile := filepath.Join(migrationsPath, fmt.Sprintf("%d_%s.down.sql", timestamp, name))

	if err := os.MkdirAll(migrationsPath, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	if err := os.WriteFile(upFile, []byte("-- Up migration\n"), 0644); err != nil {
		return fmt.Errorf("failed to create up migration file: %w", err)
	}

	if err := os.WriteFile(downFile, []byte("-- Down migration\n"), 0644); err != nil {
		return fmt.Errorf("failed to create down migration file: %w", err)
	}

	fmt.Printf("Created migration files:\n %s\n %s\n", upFile, downFile)
	return nil
}
