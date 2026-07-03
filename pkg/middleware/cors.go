package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/quoctann/content-hub/pkg/config"
)

func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	allowed := buildAllowedOrigins(cfg.Security.AllowedOrigins)

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if allowed[origin] {
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

func buildAllowedOrigins(origins []string) map[string]bool {
	if len(origins) > 0 {
		m := make(map[string]bool, len(origins))
		for _, o := range origins {
			m[o] = true
		}
		return m
	}

	return map[string]bool{
		"http://localhost:5173": true,
		"http://localhost:3000": true,
		"http://127.0.0.1:5173": true,
		"http://127.0.0.1:3000": true,
	}
}
