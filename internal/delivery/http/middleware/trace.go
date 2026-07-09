package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/quoctann/content-hub/pkg/logger"
)

const (
	HeaderRequestID = "X-Request-ID"
	HeaderTraceID   = "X-Trace-ID"
	HeaderSpanID    = "X-Span-ID"
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

		// Trace ID (Propagate if exists, else generate)
		traceID := c.GetHeader(HeaderTraceID)
		if traceID == "" {
			traceID = uuid.New().String()
		}
		c.Writer.Header().Set(HeaderTraceID, traceID)
		c.Set(logger.TraceIDKey, traceID)

		// Span ID (New span for this service)
		spanID := uuid.New().String()
		c.Writer.Header().Set(HeaderSpanID, spanID)
		c.Set(logger.SpanIDKey, spanID)

		// Update and wrap the context so that logger.extractTraceInfo can consume it
		ctx := c.Request.Context()
		ctx = logger.WithTraceContext(ctx, reqID, traceID, spanID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
