package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/quoctann/content-hub/pkg/logger"
)

const (
	requestIDHeader = "X-Request-ID"
)

// RequestIDMiddleware propagates a caller request ID or creates a new one.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(requestIDHeader)
		if _, err := uuid.Parse(requestID); err != nil {
			requestID = uuid.NewString()
		}

		c.Header(requestIDHeader, requestID)
		c.Set(logger.RequestIDKey, requestID)
		c.Request = c.Request.WithContext(logger.WithRequestID(c.Request.Context(), requestID))
		c.Next()
	}
}
