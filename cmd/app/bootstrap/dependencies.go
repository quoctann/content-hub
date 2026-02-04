package bootstrap

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quoctann/content-hub/cmd/shared"
	"github.com/quoctann/content-hub/pkg/config"
	"github.com/quoctann/content-hub/pkg/database"
	"github.com/quoctann/content-hub/pkg/logger"
)

// Dependencies holds all initialized application dependencies.
type Dependencies struct {
	Logger logger.ILogger
	Config *config.Config
	DBPool *pgxpool.Pool
}

// InitDependencies initializes all application dependencies.
// It uses shared initialization logic and follows the fail‑fast principle.
func InitDependencies(env string) (*Dependencies, error) {
	l, err := shared.InitLogger(env)
	if err != nil {
		return nil, err
	}

	cfg, err := shared.LoadConfig(env)
	if err != nil {
		return nil, err
	}

	dbPool, err := database.NewPostgresConnection(cfg, l)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Dependencies{
		Logger: l,
		Config: cfg,
		DBPool: dbPool,
	}, nil
}
