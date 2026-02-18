package http

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/quoctann/content-hub/docs"
	"github.com/quoctann/content-hub/pkg/server"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RegisterRoutes sets up all the routes for the application
func RegisterRoutes(r server.Router) {
	// Basic Health Check (deprecated, use /health/live instead)
	r.GET("/health", func(c server.Context) {
		c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	// Leapcell health check
	r.GET("/kaithheathcheck", func(c server.Context) {
		c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	// Swagger
	// Swagger (this still needs gin.HandlerFunc for now as it's a 3rd party lib)
	// We handle this by casting to GinRouter if we know it's Gin, or providing an adapter.
	// For now, we'll keep it simple if possible or just use the raw engine if needed in bootstrap.
	if gr, ok := r.(*server.GinRouter); ok {
		gr.Engine().GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}

// RegisterHealthChecks sets up Kubernetes-compatible health check endpoints
// @Summary Liveness probe
// @Description Returns 200 if the application is running
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health/live [get]
func RegisterHealthChecks(r server.Router, dbPool *pgxpool.Pool) {
	health := r.Group("/health")
	{
		// Liveness probe - checks if the application is running
		health.GET("/live", func(c server.Context) {
			c.JSON(http.StatusOK, map[string]string{
				"status": "alive",
			})
		})

		// Readiness probe - checks if the application is ready to serve traffic
		// @Summary Readiness probe
		// @Description Returns 200 if the application is ready to serve traffic (database is accessible)
		// @Tags health
		// @Produce json
		// @Success 200 {object} map[string]string
		// @Failure 503 {object} map[string]string
		// @Router /health/ready [get]
		health.GET("/ready", func(c server.Context) {
			ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
			defer cancel()

			// Check database connectivity
			if err := dbPool.Ping(ctx); err != nil {
				c.JSON(http.StatusServiceUnavailable, map[string]string{
					"status": "not ready",
					"error":  "database connection failed",
				})
				return
			}

			c.JSON(http.StatusOK, map[string]string{
				"status": "ready",
			})
		})
	}
}
