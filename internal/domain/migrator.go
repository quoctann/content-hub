package domain

// Clean-Arch: this is Port (Driver Port / Service Provider Interface)

// Migrator defines the contract for database migration operations.
// This is the *port* in the clean‑architecture sense – it has no external
// dependencies and can be implemented by any migration library.
type Migrator interface {
	// Up runs all pending migrations.
	Up() error
	// Down rolls back the most recent migration.
	Down() error
	// Force sets the migration version to a specific value. Useful for
	// recovering from a dirty state.
	Force(version int) error
	// Version returns the current migration version and whether the state
	// is dirty (i.e. the last migration failed).
	Version() (version uint, dirty bool, err error)
}

// MigrationCreator defines the contract for creating new migration files.
// It does not need a database connection – it only works with the file system.
type MigrationCreator interface {
	// CreateMigration creates a pair of up/down SQL files with a timestamp‑
	// based name inside the provided migrationsPath.
	CreateMigration(migrationsPath, name string) error
}
