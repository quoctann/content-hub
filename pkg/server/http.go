package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
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
		engine: gin.Default(),
	}

	for _, opt := range opts {
		opt(s)
	}

	s.httpServer = &http.Server{
		Addr:           s.config.Addr,
		Handler:        s.engine,
		ReadTimeout:    s.config.ReadTimeout,
		WriteTimeout:   s.config.WriteTimeout,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	return s
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
