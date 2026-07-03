package bootstrap

import (
	"context"
	"time"

	"github.com/quoctann/content-hub/pkg/logger"
	"github.com/quoctann/content-hub/pkg/middleware"
	"github.com/quoctann/content-hub/pkg/server"
)

type App struct {
	deps   *Dependencies
	server server.Server
}

func NewApp() (*App, error) {
	deps, err := InitDependencies()
	if err != nil {
		return nil, err
	}

	// Determine Gin mode based on environment
	ginMode := "debug"
	if deps.Config.Server.AppEnv == "prod" {
		ginMode = "release"
	}

	srv := server.NewHTTPServer(
		server.WithAddr(":"+deps.Config.Server.Port),
		server.WithShutdownTimeout(30*time.Second),
		server.WithGinMode(ginMode),
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

		// Apply security headers middleware
		srv.Router().Use(middleware.SecurityHeaders(middleware.CSPOptions{
			DefaultSrc: app.deps.Config.Security.CSPDefaultSrc,
			ScriptSrc:  app.deps.Config.Security.CSPScriptSrc,
			StyleSrc:   app.deps.Config.Security.CSPStyleSrc,
			ImgSrc:     app.deps.Config.Security.CSPImgSrc,
			FontSrc:    app.deps.Config.Security.CSPFontSrc,
			ConnectSrc: app.deps.Config.Security.CSPConnectSrc,
			FrameSrc:   app.deps.Config.Security.CSPFrameSrc,
			MediaSrc:   app.deps.Config.Security.CSPMediaSrc,
			ObjectSrc:  app.deps.Config.Security.CSPObjectSrc,
		}))

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
			_ = app.deps.Logger.Sync()
		}
		app.deps.Logger.InfoWithoutCtx("Server exited gracefully")
		return nil
	})

	return app, nil
}

func (a *App) Start() error {
	return a.server.Start()
}

func (a *App) Stop(ctx context.Context) error {
	return a.server.Stop(ctx)
}
