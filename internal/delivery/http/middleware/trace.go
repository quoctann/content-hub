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

// TraceMiddleware adds tracing information to the request context and headers
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Request ID
		reqID := c.GetHeader(HeaderRequestID)
		if reqID == "" {
			reqID = uuid.New().String()
		}
		c.Writer.Header().Set(HeaderRequestID, reqID)
		c.Set(logger.RequestIDKey, reqID)

		// 2. Trace ID (Propagate if exists, else generate)
		traceID := c.GetHeader(HeaderTraceID)
		if traceID == "" {
			traceID = uuid.New().String()
		}
		c.Writer.Header().Set(HeaderTraceID, traceID)
		c.Set(logger.TraceIDKey, traceID)

		// 3. Span ID (New span for this service)
		spanID := uuid.New().String()
		c.Writer.Header().Set(HeaderSpanID, spanID)
		c.Set(logger.SpanIDKey, spanID)

		// 4. Update Context for Logger (This is crucial)
		// We need to wrap the context so that logger.extractTraceInfo can find these values
		ctx := c.Request.Context()
		ctx = logger.WithTraceContext(ctx, reqID, traceID, spanID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
