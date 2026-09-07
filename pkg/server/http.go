package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quoctann/content-hub/pkg/logger"
	"go.opentelemetry.io/otel/trace"
)

// HTTPConfig defines the configuration for the HTTP server
type HTTPConfig struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// HTTPOption is a functional option for configuring the HTTPServer
type HTTPOption func(*HTTPServer)

// HTTPServer implements the Server interface for HTTP using Gin
type HTTPServer struct {
	*HookManager
	config     HTTPConfig
	engine     *gin.Engine
	httpServer *http.Server
}

// NewHTTPServer creates a new HTTPServer instance
func NewHTTPServer(opts ...HTTPOption) *HTTPServer {
	s := &HTTPServer{
		HookManager: &HookManager{},
		config: HTTPConfig{
			Addr:            ":8080",
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 10 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	s.engine = newGinEngine()

	s.httpServer = &http.Server{
		Addr:           s.config.Addr,
		Handler:        s.engine,
		ReadTimeout:    s.config.ReadTimeout,
		WriteTimeout:   s.config.WriteTimeout,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	return s
}

func newGinEngine() *gin.Engine {
	engine := gin.New()
	engine.Use(ginJSONLogger(), ginJSONRecovery())
	return engine
}

func ginJSONLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		entry := map[string]interface{}{
			"timestamp":  time.Now().Format(time.RFC3339Nano),
			"level":      "info",
			"service":    "content-hub",
			"env":        os.Getenv("APP_ENV"),
			"msg":        "http_request",
			"status":     c.Writer.Status(),
			"latency_ms": float64(time.Since(start).Microseconds()) / 1000,
			"client_ip":  c.ClientIP(),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"user_agent": c.Request.UserAgent(),
			"body_size":  c.Writer.Size(),
		}

		if requestID := traceValue(c, "X-Request-ID", logger.RequestIDKey); requestID != "" {
			entry["request_id"] = requestID
		}
		spanContext := trace.SpanContextFromContext(c.Request.Context())
		if spanContext.IsValid() {
			entry["trace_id"] = spanContext.TraceID().String()
			entry["span_id"] = spanContext.SpanID().String()
		}
		if errs := c.Errors.String(); errs != "" {
			entry["error"] = errs
		}
		if c.Writer.Status() >= http.StatusInternalServerError {
			entry["level"] = "error"
			writeJSONLog(os.Stderr, entry)
			return
		}

		writeJSONLog(os.Stdout, entry)
	}
}

func ginJSONRecovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		entry := map[string]interface{}{
			"timestamp":  time.Now().Format(time.RFC3339Nano),
			"level":      "error",
			"service":    "content-hub",
			"env":        os.Getenv("APP_ENV"),
			"msg":        "panic_recovered",
			"error":      fmt.Sprint(recovered),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"stacktrace": string(debug.Stack()),
		}
		addTraceFields(c, entry)
		writeJSONLog(os.Stderr, entry)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

func addTraceFields(c *gin.Context, entry map[string]interface{}) {
	if requestID := traceValue(c, "X-Request-ID", logger.RequestIDKey); requestID != "" {
		entry["request_id"] = requestID
	}
	spanContext := trace.SpanContextFromContext(c.Request.Context())
	if spanContext.IsValid() {
		entry["trace_id"] = spanContext.TraceID().String()
		entry["span_id"] = spanContext.SpanID().String()
	}
}

func writeJSONLog(file *os.File, entry map[string]interface{}) {
	encoded, err := json.Marshal(entry)
	if err != nil {
		return
	}
	_, _ = file.Write(append(encoded, '\n'))
}

func traceValue(c *gin.Context, headerName, contextKey string) string {
	if value := c.GetHeader(headerName); value != "" {
		return value
	}
	if value, ok := c.Get(contextKey); ok {
		if valueString, ok := value.(string); ok {
			return valueString
		}
	}
	return ""
}

// WithGinMode sets the Gin mode (debug or release)
func WithGinMode(mode string) HTTPOption {
	return func(s *HTTPServer) {
		gin.SetMode(mode)
	}
}

// Functional options
func WithAddr(addr string) HTTPOption {
	return func(s *HTTPServer) { s.config.Addr = addr }
}

func WithReadTimeout(d time.Duration) HTTPOption {
	return func(s *HTTPServer) { s.config.ReadTimeout = d }
}

func WithWriteTimeout(d time.Duration) HTTPOption {
	return func(s *HTTPServer) { s.config.WriteTimeout = d }
}

func WithShutdownTimeout(d time.Duration) HTTPOption {
	return func(s *HTTPServer) { s.config.ShutdownTimeout = d }
}

// Router returns a framework‑agnostic router interface.
// Use this method for setting up routes in a framework‑independent way.
func (s *HTTPServer) Router() Router {
	return NewGinRouter(s.engine)
}

// Engine returns the underlying Gin engine.
// Deprecated: Use Router() for framework‑agnostic route registration.
// This method exists for backward compatibility and Gin‑specific features.
func (s *HTTPServer) Engine() *gin.Engine {
	return s.engine
}

// Start starts the HTTP server and waits for interrupt signals to perform graceful shutdown
func (s *HTTPServer) Start() error {
	if err := s.RunBeforeStart(); err != nil {
		return err
	}

	errChan := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	if err := s.RunAfterStart(); err != nil {
		return err
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return err
	case <-quit:
		// Graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
		defer cancel()
		return s.Stop(ctx)
	}
}

// Stop gracefully shuts down the server
func (s *HTTPServer) Stop(ctx context.Context) error {
	if err := s.RunBeforeStop(); err != nil {
		return err
	}

	// Use provided context or default shutdown timeout
	shutdownCtx := ctx
	if shutdownCtx == nil {
		var cancel context.CancelFunc
		shutdownCtx, cancel = context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
		defer cancel()
	}

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return err
	}

	return s.RunAfterStop()
}
