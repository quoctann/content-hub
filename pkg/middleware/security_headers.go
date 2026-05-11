package middleware

import (
	"github.com/quoctann/content-hub/pkg/server"
)

// SecurityHeaders returns a middleware that sets common security-related HTTP response headers.
func SecurityHeaders() server.MiddlewareFunc {
	return func(c server.Context) (server.Context, error) {
		c.SetHeader("X-Content-Type-Options", "nosniff")
		c.SetHeader("X-Frame-Options", "DENY")
		c.SetHeader("Referrer-Policy", "strict-origin-when-cross-origin")
		c.SetHeader("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.SetHeader("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		return c, nil
	}
}
