package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/quoctann/content-hub/pkg/config"
	"github.com/quoctann/content-hub/pkg/logger"
)

func NewPostgresConnection(cfg *config.Config, l logger.ILogger) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.Database.Schema != "" {
		poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
			_, err := conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s", cfg.Database.Schema))
			return err
		}
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
		Logger:   NewLoggerAdapter(l),
		LogLevel: dbLogLevel,
	}
	poolConfig.ConnConfig.Tracer = dbTracer

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Verify schema
	var dbName, user, searchPath string
	var schemas []string

	err = pool.
		QueryRow(context.Background(), `SELECT current_database(), current_user, current_setting('search_path'), current_schemas(false)`).
		Scan(&dbName, &user, &searchPath, &schemas)
	if err != nil {
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
