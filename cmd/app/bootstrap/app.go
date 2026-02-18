package bootstrap

import (
	"context"
	"time"

	"github.com/quoctann/content-hub/pkg/logger"
	"github.com/quoctann/content-hub/pkg/middleware"
	"github.com/quoctann/content-hub/pkg/server"
)

// App represents the fully initialized application.
type App struct {
	deps   *Dependencies
	server server.Server
}

// NewApp creates and initializes a new application instance.
func NewApp(env string) (*App, error) {
	deps, err := InitDependencies(env)
	if err != nil {
		return nil, err
	}

	srv := server.NewHTTPServer(
		server.WithAddr(":"+deps.Config.Server.Port),
		server.WithShutdownTimeout(30*time.Second),
	)

	app := &App{
		deps:   deps,
		server: srv,
	}

	srv.OnBeforeStart(func() error {
		app.deps.Logger.InfoWithoutCtx("Starting server", logger.String("addr", ":"+app.deps.Config.Server.Port))

		// Apply CORS middleware
		ginRouter := srv.Engine()
		ginRouter.Use(middleware.CORSMiddleware(app.deps.Config))

		// Auto-migration
		if err := RunMigrations(app.deps); err != nil {
			return err
		}

		SetupRouter(srv.Router(), app.deps)
		return nil
	})

	srv.OnAfterStop(func() error {
		app.deps.Logger.InfoWithoutCtx("Shutting down server...")
		if app.deps.DBPool != nil {
			app.deps.DBPool.Close()
		}
		if app.deps.Logger != nil {
			app.deps.Logger.Sync()
		}
		app.deps.Logger.InfoWithoutCtx("Server exited gracefully")
		return nil
	})

	return app, nil
}

// Start runs the application server.
func (a *App) Start() error {
	return a.server.Start()
}

// Stop gracefully shuts down the application.
func (a *App) Stop(ctx context.Context) error {
	return a.server.Stop(ctx)
}
