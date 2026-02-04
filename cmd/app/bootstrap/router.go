package bootstrap

import (
	"time"

	httpDelivery "github.com/quoctann/content-hub/internal/delivery/http"
	"github.com/quoctann/content-hub/internal/repository/postgres"
	"github.com/quoctann/content-hub/internal/usecase"
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

	// Setup User module
	userRepo := postgres.NewUserRepo(deps.DBPool)
	userUsecase := usecase.NewUserUsecase(userRepo)
	httpDelivery.NewUserHandler(router, userUsecase)

	// Setup Content module
	contentRepo := postgres.NewContentRepo(deps.DBPool)
	contentUsecase := usecase.NewContentUsecase(contentRepo, 5*time.Second)
	httpDelivery.NewContentHandler(router, contentUsecase)
}
