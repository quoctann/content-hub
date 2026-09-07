package bootstrap

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	traceMiddleware "github.com/quoctann/content-hub/internal/delivery/http/middleware"
	"github.com/quoctann/content-hub/pkg/logger"
	"github.com/quoctann/content-hub/pkg/middleware"
	"github.com/quoctann/content-hub/pkg/observability"
	"github.com/quoctann/content-hub/pkg/server"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type App struct {
	deps          *Dependencies
	server        server.Server
	metricsServer *http.Server
}

func NewApp() (*App, error) {
	deps, err := InitDependencies()
	if err != nil {
		return nil, err
	}

	// Determine Gin mode based on environment
	ginMode := gin.DebugMode
	if deps.Config.Server.AppEnv == "prod" {
		ginMode = gin.ReleaseMode
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
		ginRouter.Use(
			traceMiddleware.TraceMiddleware(),
			otelgin.Middleware(app.deps.Config.Tracing.ServiceName),
			middleware.CORSMiddleware(app.deps.Config),
		)

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

		// Metrics
		metrics, err := observability.NewMetrics(nil)
		if err != nil {
			return err
		}
		if err := observability.RegisterDatabasePoolMetrics(nil, app.deps.DBPool); err != nil {
			return err
		}

		ginRouter.Use(metrics.Middleware())

		metricsListener, err := net.Listen("tcp", ":"+app.deps.Config.Metrics.Port)
		if err != nil {
			return fmt.Errorf("listen for metrics: %w", err)
		}
		app.metricsServer = &http.Server{Handler: metrics.Handler()}
		go func() {
			_ = app.metricsServer.Serve(metricsListener)
		}()

		SetupRouter(srv.Router(), app.deps)
		return nil
	})

	srv.OnAfterStop(func() error {
		app.deps.Logger.InfoWithoutCtx("Shutting down server...")
		if app.metricsServer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := app.metricsServer.Shutdown(ctx)
			cancel()
			if err != nil {
				return err
			}
		}

		if app.deps.DBPool != nil {
			app.deps.DBPool.Close()
		}

		if app.deps.TracingShutdown != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := app.deps.TracingShutdown(ctx)
			cancel()
			if err != nil {
				return err
			}
		}

		app.deps.Logger.InfoWithoutCtx("Server exited gracefully")
		if app.deps.Logger != nil {
			_ = app.deps.Logger.Sync()
		}
		return nil
	})

	return app, nil
}

func (a *App) Start() error {
	defer func() {
		if a.metricsServer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = a.metricsServer.Shutdown(ctx)
		}
	}()

	if a.deps.TracingShutdown != nil {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = a.deps.TracingShutdown(ctx)
		}()
	}

	return a.server.Start()
}

func (a *App) Stop(ctx context.Context) error {
	return a.server.Stop(ctx)
}
