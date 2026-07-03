package bootstrap

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quoctann/content-hub/cmd/shared"
	"github.com/quoctann/content-hub/pkg/config"
	"github.com/quoctann/content-hub/pkg/database"
	"github.com/quoctann/content-hub/pkg/logger"
)

type Dependencies struct {
	Logger logger.ILogger
	Config *config.Config
	DBPool *pgxpool.Pool
}

func InitDependencies() (*Dependencies, error) {
	cfg, err := shared.LoadConfig()
	if err != nil {
		return nil, err
	}

	l, err := shared.InitLogger(cfg.Server.AppEnv, cfg.Logger.Level)
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
