package middleware

import (
	"fmt"

	"github.com/quoctann/content-hub/pkg/server"
)

func SecurityHeaders(cspConnectSrc string) server.MiddlewareFunc {
	csp := fmt.Sprintf(
		"default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'%s; frame-ancestors 'none'",
		func() string {
			if cspConnectSrc != "" {
				return " " + cspConnectSrc
			}
			return ""
		}(),
	)

	return func(c server.Context) (server.Context, error) {
		c.SetHeader("X-Content-Type-Options", "nosniff")
		c.SetHeader("X-Frame-Options", "DENY")
		c.SetHeader("Referrer-Policy", "strict-origin-when-cross-origin")
		c.SetHeader("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.SetHeader("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.SetHeader("Content-Security-Policy", csp)
		return c, nil
	}
}
