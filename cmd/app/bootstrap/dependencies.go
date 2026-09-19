package bootstrap

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/quoctann/content-hub/cmd/shared"
	"github.com/quoctann/content-hub/pkg/config"
	"github.com/quoctann/content-hub/pkg/database"
	"github.com/quoctann/content-hub/pkg/logger"
	"github.com/quoctann/content-hub/pkg/observability"
)

type Dependencies struct {
	Logger logger.ILogger
	Config *config.Config
	DBPool *pgxpool.Pool

	shutdownTracing observability.Shutdown
	closeOnce       sync.Once
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

	shutdownTracing, err := observability.InitTracing(context.Background(), observability.TracingConfig{
		Disabled:    cfg.Observability.OTelSDKDisabled,
		Endpoint:    cfg.Observability.OTelExporterOTLPEndpoint,
		Insecure:    cfg.Observability.OTelExporterOTLPInsecure,
		ServiceName: cfg.Observability.OTelServiceName,
	})
	if err != nil {
		_ = l.Sync()
		return nil, fmt.Errorf("failed to initialize tracing: %w", err)
	}

	dbPool, err := database.NewPostgresConnection(cfg, l)
	if err != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdownTracing(shutdownCtx)
		_ = l.Sync()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Dependencies{
		Logger:          l,
		Config:          cfg,
		DBPool:          dbPool,
		shutdownTracing: shutdownTracing,
	}, nil
}

// Close releases application dependencies. It is safe to call more than once.
func (d *Dependencies) Close() {
	d.closeOnce.Do(func() {
		if d.DBPool != nil {
			d.DBPool.Close()
		}

		if d.shutdownTracing != nil {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := d.shutdownTracing(shutdownCtx); err != nil && d.Logger != nil {
				d.Logger.ErrorWithoutCtx("Failed to shutdown tracing", logger.Error(err))
			}
			cancel()
		}

		if d.Logger != nil {
			d.Logger.InfoWithoutCtx("Server exited gracefully")
			_ = d.Logger.Sync()
		}
	})
}
