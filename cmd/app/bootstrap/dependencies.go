package bootstrap

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quoctann/content-hub/cmd/shared"
	"github.com/quoctann/content-hub/pkg/config"
	"github.com/quoctann/content-hub/pkg/database"
	"github.com/quoctann/content-hub/pkg/logger"
	"github.com/quoctann/content-hub/pkg/observability"
)

type Dependencies struct {
	Logger          logger.ILogger
	Config          *config.Config
	DBPool          *pgxpool.Pool
	TracingShutdown func(context.Context) error
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
	l = l.With(
		logger.String("service", cfg.Tracing.ServiceName),
		logger.String("environment", cfg.Server.AppEnv),
		logger.String("version", cfg.Tracing.ServiceVer),
	)
	tracingShutdown, err := observability.InitTracing(context.Background(), observability.TracingConfig{
		Endpoint:    cfg.Tracing.OTLPEndpoint,
		ServiceName: cfg.Tracing.ServiceName,
		Version:     cfg.Tracing.ServiceVer,
		Environment: cfg.Server.AppEnv,
		SampleRatio: cfg.Tracing.SampleRatio,
		TLS:         cfg.Tracing.TLS,
	})
	if err != nil {
		return nil, fmt.Errorf("initialize tracing: %w", err)
	}

	dbPool, err := database.NewPostgresConnection(cfg, l)
	if err != nil {
		_ = tracingShutdown(context.Background())
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	var (
		tracingShutdownOnce sync.Once
		tracingShutdownErr  error
	)
	shutdownTracing := func(ctx context.Context) error {
		tracingShutdownOnce.Do(func() {
			tracingShutdownErr = tracingShutdown(ctx)
		})
		return tracingShutdownErr
	}

	return &Dependencies{
		Logger:          l,
		Config:          cfg,
		DBPool:          dbPool,
		TracingShutdown: shutdownTracing,
	}, nil
}
