package middleware

import (
	"github.com/quoctann/content-hub/pkg/server"
)

// APIKeyAuth is a middleware that checks for a valid API key in the request header.
// Key: "X-API-Key"
func APIKeyAuth(validKey string) server.MiddlewareFunc {
	return func(c server.Context) (server.Context, error) {
		key := c.Request().Header.Get("X-API-Key")
		if key == "" {
			// Also check query param as a fallback/convenience?
			// The user asked for "simple", usually header is best practice.
			// Let's stick to header for now.
			return nil, server.Error401("missing API key")
		}

		if key != validKey {
			return nil, server.Error401("invalid API key")
		}

		return c, nil
	}
}
