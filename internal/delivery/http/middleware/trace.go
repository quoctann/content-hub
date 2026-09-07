package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/quoctann/content-hub/pkg/logger"
)

const (
	HeaderRequestID = "X-Request-ID"
)

func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Request ID
		reqID := c.GetHeader(HeaderRequestID)
		if reqID == "" {
			reqID = uuid.New().String()
		}
		c.Writer.Header().Set(HeaderRequestID, reqID)
		c.Set(logger.RequestIDKey, reqID)

		// OpenTelemetry owns W3C trace propagation; this middleware only creates
		// the application-level request correlation identifier.
		ctx := c.Request.Context()
		ctx = logger.WithTraceContext(ctx, reqID, "", "")
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
