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

func SetupRouter(router server.Router, deps *Dependencies) {
	httpDelivery.RegisterHealthChecks(router)
	httpDelivery.RegisterK8SHealthChecks(router, deps.DBPool)

	// Setup Content module
	contentRepo := postgres.NewContentRepo(deps.DBPool)
	contentUsecase := usecase.NewContentUsecase(contentRepo, 5*time.Second)

	// Setup Account module
	accountRepo := postgres.NewAccountRepo(deps.DBPool)
	accountUsecase := usecase.NewAccountUsecase(accountRepo, deps.Config.Security.JWTSecret, deps.Config.Security.JWTExpiry)

	// Public routes (Login) — rate limited to 5 attempts per minute per IP
	loginGroup := router.Group("/")
	loginGroup.Use(middleware.RateLimiter(rate.Every(time.Minute/5), 5))
	httpDelivery.NewAccountHandler(loginGroup, accountUsecase, deps.Config, deps.Logger)

	// Protected content routes
	contentGroup := router.Group("/")
	// contentGroup.Use(middleware.APIKeyAuth(deps.Config.Security.APIKey))
	httpDelivery.NewContentHandler(contentGroup, contentUsecase, deps.Logger)

	// Protected routes (JWT auth + CSRF)
	adminGroup := router.Group("/admin")
	adminGroup.Use(middleware.JWTAuth(middleware.JWTAuthConfig{
		Secret: deps.Config.Security.JWTSecret,
		Expiry: deps.Config.Security.JWTExpiry,
		Issuer: "content-hub",
	}))
	adminGroup.Use(middleware.CSRFProtection())
	httpDelivery.NewAdminContentHandler(adminGroup.Group("/contents"), contentUsecase, deps.Logger)
	
	mediaGroup := adminGroup.Group("/media")
	mediaGroup.Use(middleware.RateLimiter(rate.Every(time.Minute/12), 5))
	httpDelivery.NewUploadHandler(mediaGroup, deps.MediaStore, deps.Config.Upload, deps.Logger)
}
