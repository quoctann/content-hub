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
	ServiceName     string
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
	middleware []gin.HandlerFunc
	engine     *gin.Engine
	httpServer *http.Server
}

// NewHTTPServer creates a new HTTPServer instance
func NewHTTPServer(opts ...HTTPOption) *HTTPServer {
	s := &HTTPServer{
		HookManager: &HookManager{},
		config: HTTPConfig{
			Addr:            ":8080",
			ServiceName:     "content-hub-backend",
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 10 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	s.engine = newGinEngine(s.config.ServiceName, s.middleware...)

	s.httpServer = &http.Server{
		Addr:           s.config.Addr,
		Handler:        s.engine,
		ReadTimeout:    s.config.ReadTimeout,
		WriteTimeout:   s.config.WriteTimeout,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	return s
}

func newGinEngine(serviceName string, outerMiddleware ...gin.HandlerFunc) *gin.Engine {
	engine := gin.New()
	engine.Use(outerMiddleware...)
	engine.Use(ginJSONLogger(serviceName), ginJSONRecovery())
	return engine
}

func ginJSONLogger(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = "unmatched"
		}
		entry := map[string]interface{}{
			"timestamp":  time.Now().Format(time.RFC3339Nano),
			"level":      "info",
			"service":    serviceName,
			"env":        os.Getenv("APP_ENV"),
			"msg":        "http_request",
			"status":     c.Writer.Status(),
			"latency_ms": float64(time.Since(start).Microseconds()) / 1000,
			"method":     c.Request.Method,
			"path":       path,
			"body_size":  c.Writer.Size(),
		}

		addCorrelationFields(entry, c.Request.Context())
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
			"msg":        "panic_recovered",
			"error":      fmt.Sprint(recovered),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"stacktrace": string(debug.Stack()),
		}
		addCorrelationFields(entry, c.Request.Context())
		writeJSONLog(os.Stderr, entry)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

func writeJSONLog(file *os.File, entry map[string]interface{}) {
	encoded, err := json.Marshal(entry)
	if err != nil {
		return
	}
	_, _ = file.Write(append(encoded, '\n'))
}

func addCorrelationFields(entry map[string]interface{}, ctx context.Context) {
	if requestID := logger.RequestIDFromContext(ctx); requestID != "" {
		entry["request_id"] = requestID
	}
	if spanContext := trace.SpanContextFromContext(ctx); spanContext.IsValid() {
		entry["trace_id"] = spanContext.TraceID().String()
		entry["span_id"] = spanContext.SpanID().String()
	}
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

func WithServiceName(serviceName string) HTTPOption {
	return func(s *HTTPServer) { s.config.ServiceName = serviceName }
}

// WithGinMiddleware registers middleware outside access logging and recovery.
func WithGinMiddleware(middleware ...gin.HandlerFunc) HTTPOption {
	return func(s *HTTPServer) { s.middleware = append(s.middleware, middleware...) }
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
