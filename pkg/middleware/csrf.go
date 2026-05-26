package middleware

import (
	"github.com/quoctann/content-hub/pkg/server"
)

func CSRFProtection() server.MiddlewareFunc {
	return func(c server.Context) (server.Context, error) {
		if c.Request().Method == "GET" || c.Request().Method == "HEAD" || c.Request().Method == "OPTIONS" {
			return c, nil
		}

		cookieCSRF, err := c.Cookie("csrf_token")
		if err != nil || cookieCSRF == "" {
			return nil, server.Error403("missing csrf token")
		}

		headerCSRF := c.Request().Header.Get("X-CSRF-Token")
		if headerCSRF == "" {
			return nil, server.Error403("missing csrf token header")
		}

		if cookieCSRF != headerCSRF {
			return nil, server.Error403("csrf token mismatch")
		}

		return c, nil
	}
}
