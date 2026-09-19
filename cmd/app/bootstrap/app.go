package bootstrap

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quoctann/content-hub/pkg/logger"
	"github.com/quoctann/content-hub/pkg/middleware"
	"github.com/quoctann/content-hub/pkg/server"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
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
	ginMode := gin.DebugMode
	if deps.Config.Server.AppEnv == "prod" {
		ginMode = gin.ReleaseMode
	}

	srv := server.NewHTTPServer(
		server.WithAddr(":"+deps.Config.Server.Port),
		server.WithServiceName(deps.Config.Observability.OTelServiceName),
		server.WithGinMiddleware(
			middleware.RequestIDMiddleware(),
			otelgin.Middleware(
				deps.Config.Observability.OTelServiceName,
				otelgin.WithFilter(func(request *http.Request) bool {
					return request.URL.Path != "/health" && !strings.HasPrefix(request.URL.Path, "/health/")
				}),
			),
		),
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
		app.deps.Close()
		return nil
	})

	return app, nil
}

func (a *App) Start() error {
	defer a.deps.Close()
	return a.server.Start()
}

func (a *App) Stop(ctx context.Context) error {
	return a.server.Stop(ctx)
}
