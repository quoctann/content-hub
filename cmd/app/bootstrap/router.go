package bootstrap

import (
	"time"

	"golang.org/x/time/rate"

	httpDelivery "github.com/quoctann/content-hub/internal/delivery/http"
	"github.com/quoctann/content-hub/internal/repository/postgres"
	"github.com/quoctann/content-hub/internal/usecase"
	"github.com/quoctann/content-hub/pkg/middleware"
	"github.com/quoctann/content-hub/pkg/server"
)

// SetupRouter registers all application routes and handlers.
// It uses the framework‑agnostic server.Router interface.
func SetupRouter(router server.Router, deps *Dependencies) {
	// Setup generic routes (health, swagger, etc.)
	// Note: We'll need to update httpDelivery to use server.Router
	httpDelivery.RegisterRoutes(router)

	// Setup health check endpoints
	httpDelivery.RegisterHealthChecks(router, deps.DBPool)

	// Setup Content module
	contentRepo := postgres.NewContentRepo(deps.DBPool)
	contentUsecase := usecase.NewContentUsecase(contentRepo, 5*time.Second)

	// Setup Account module
	accountRepo := postgres.NewAccountRepo(deps.DBPool)
	jwtExpiry, _ := time.ParseDuration(deps.Config.Security.JWTExpiry)
	if jwtExpiry == 0 {
		jwtExpiry = 24 * time.Hour
	}
	accountUsecase := usecase.NewAccountUsecase(accountRepo, deps.Config.Security.JWTSecret, jwtExpiry)

	// Public routes (Login) — rate limited to 5 attempts per minute per IP
	loginGroup := router.Group("/")
	loginGroup.Use(middleware.RateLimiter(rate.Every(time.Minute/5), 5))
	httpDelivery.NewAccountHandler(loginGroup, accountUsecase, deps.Config, deps.Logger)

	// Protected content routes
	contentGroup := router.Group("/")
	// contentGroup.Use(middleware.APIKeyAuth(deps.Config.Security.APIKey))
	httpDelivery.NewContentHandler(contentGroup, contentUsecase, deps.Logger)

	// Protected routes (JWT auth + CSRF)
	adminGroup := router.Group("/admin/contents")
	adminGroup.Use(middleware.JWTAuth(middleware.JWTAuthConfig{
		Secret: deps.Config.Security.JWTSecret,
		Expiry: jwtExpiry, // Re-use the parsed jwtExpiry variable
		Issuer: "content-hub",
	}))
	adminGroup.Use(middleware.CSRFProtection())
	httpDelivery.NewAdminContentHandler(adminGroup, contentUsecase, deps.Logger)
}
