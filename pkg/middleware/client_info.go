package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/quoctann/content-hub/pkg/logger"
)

// ClientInfoMiddleware attaches Cloudflare client metadata to the request context
// before access logging, recovery and handlers run.
func ClientInfoMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip, fromCloudflare := clientIP(c.Request)
		info := logger.ClientInfo{IP: ip}
		if fromCloudflare {
			if country := singleHeader(c.Request, "CF-IPCountry"); len(country) == 2 &&
				country[0] >= 'A' && country[0] <= 'Z' && country[1] >= 'A' && country[1] <= 'Z' {
				info.Country = country
			}
			if ray := singleHeader(c.Request, "CF-Ray"); len(ray) <= 64 && ray != "" &&
				strings.IndexFunc(ray, func(r rune) bool {
					return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
						(r >= '0' && r <= '9') || r == '-')
				}) == -1 {
				info.Ray = ray
			}
		}
		c.Request = c.Request.WithContext(logger.WithClientInfo(c.Request.Context(), info))
		c.Next()
	}
}

func singleHeader(req *http.Request, key string) string {
	if values := req.Header.Values(key); len(values) == 1 {
		return values[0]
	}
	return ""
}
