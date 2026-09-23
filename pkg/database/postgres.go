package database

import (
	"context"
	"fmt"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/multitracer"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/quoctann/content-hub/pkg/config"
	"github.com/quoctann/content-hub/pkg/logger"
)

func NewPostgresConnection(cfg *config.Config, l logger.ILogger) (*pgxpool.Pool, error) {
	// URL carries search_path when a schema is configured, same as the migrator.
	poolConfig, err := pgxpool.ParseConfig(URL(cfg.Database))
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set some reasonable defaults
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	// Configure tracer — use DEBUG only in non-production environments to avoid
	// exposing query parameters containing sensitive data in logs.
	dbLogLevel := tracelog.LogLevelWarn
	if cfg.Server.AppEnv == "local" || cfg.Server.AppEnv == "dev" {
		dbLogLevel = tracelog.LogLevelDebug
	}
	dbTracer := &tracelog.TraceLog{
		Logger:   NewLoggerAdapter(l, dbLogLevel != tracelog.LogLevelDebug),
		LogLevel: dbLogLevel,
	}
	poolConfig.ConnConfig.Tracer = multitracer.New(
		dbTracer,
		otelpgx.NewTracer(
			otelpgx.WithDisableSQLStatementInAttributes(),
			otelpgx.WithDisableConnectionDetailsInAttributes(),
		),
	)

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Verify schema
	var dbName, user, searchPath string
	var schemas []string

	err = pool.
		QueryRow(context.Background(), `SELECT current_database(), current_user, current_setting('search_path'), current_schemas(false)`).
		Scan(&dbName, &user, &searchPath, &schemas)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to verify db config: %w", err)
	}

	l.InfoWithoutCtx(
		"Connected to Postgres with runtime config",
		logger.String("db", dbName),
		logger.String("user", user),
		logger.String("search_path", searchPath),
		logger.Any("current_schemas", schemas),
	)

	return pool, nil
}

// EnsureSchema creates the configured schema if it does not exist yet. The
// migrations do not qualify table names (they rely on search_path), and
// golang-migrate keeps its schema_migrations table there too, so the schema
// must exist before the first migration runs on a fresh database.
func EnsureSchema(ctx context.Context, cfg config.Database) error {
	if cfg.Schema == "" {
		return nil
	}
	conn, err := pgx.Connect(ctx, URL(cfg))
	if err != nil {
		return fmt.Errorf("connect to create schema: %w", err)
	}
	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS "+pgx.Identifier{cfg.Schema}.Sanitize()); err != nil {
		return fmt.Errorf("create schema %q: %w", cfg.Schema, err)
	}
	return nil
}
