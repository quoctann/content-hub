package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/quoctann/content-hub/pkg/config"
)

// CORSMiddleware returns a Gin middleware that handles CORS headers.
// It allows requests from origins specified in the configuration.
// Default fallback origins for development if not configured:
// - localhost:5173 (Vite dev server)
// - localhost:3000 (alternative dev port)
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Get allowed origins from config, with defaults for development
		allowedOrigins := getDefaultOrigins()
		origins := cfg.Security.GetAllowedOrigins()
		if len(origins) > 0 {
			allowedOrigins = make(map[string]bool)
			for _, o := range origins {
				allowedOrigins[o] = true
			}
		}

		if allowedOrigins[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, Authorization, Accept, X-CSRF-Token")
			c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, X-Total-Count")
			c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// getDefaultOrigins returns the default allowed origins for development
func getDefaultOrigins() map[string]bool {
	return map[string]bool{
		"http://localhost:5173": true,
		"http://localhost:3000": true,
		"http://127.0.0.1:5173": true,
		"http://127.0.0.1:3000": true,
	}
}
