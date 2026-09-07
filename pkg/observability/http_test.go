package observability

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMiddlewareRecordsTemplatedRouteStatusAndDuration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	registry := prometheus.NewRegistry()
	metrics, err := NewMetrics(registry)
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.Use(metrics.Middleware())
	router.GET("/contents/:id", func(c *gin.Context) { c.Status(http.StatusCreated) })

	request := httptest.NewRequest(http.MethodGet, "/contents/123?ignored=1", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	if got := testutil.ToFloat64(metrics.InFlight); got != 0 {
		t.Fatalf("in-flight = %v after request, want 0", got)
	}
	if err := testutil.CollectAndCompare(registry, strings.NewReader(`
# HELP content_hub_http_requests_total Total number of HTTP requests.
# TYPE content_hub_http_requests_total counter
content_hub_http_requests_total{method="GET",route="/contents/:id",status="201"} 1
# HELP content_hub_http_requests_in_flight Current number of HTTP requests in flight.
# TYPE content_hub_http_requests_in_flight gauge
content_hub_http_requests_in_flight{method="GET",route="/contents/:id"} 0
`), "content_hub_http_requests_total", "content_hub_http_requests_in_flight"); err != nil {
		t.Fatal(err)
	}
	if got := testutil.CollectAndCount(metrics.RequestDuration); got != 1 {
		t.Fatalf("duration metric families = %d, want 1", got)
	}
}

func TestMiddlewareUsesUnmatchedRouteFor404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	registry := prometheus.NewRegistry()
	metrics, err := NewMetrics(registry)
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.Use(metrics.Middleware())

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/not-found", nil))

	if err := testutil.CollectAndCompare(registry, strings.NewReader(`
# HELP content_hub_http_requests_total Total number of HTTP requests.
# TYPE content_hub_http_requests_total counter
content_hub_http_requests_total{method="POST",route="unmatched",status="404"} 1
`), "content_hub_http_requests_total"); err != nil {
		t.Fatal(err)
	}
}

func TestMiddlewareTracksInFlightRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	registry := prometheus.NewRegistry()
	metrics, err := NewMetrics(registry)
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	release := make(chan struct{})
	router := gin.New()
	router.Use(metrics.Middleware())
	router.GET("/slow", func(c *gin.Context) {
		close(started)
		<-release
	})

	done := make(chan struct{})
	go func() {
		router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/slow", nil))
		close(done)
	}()
	<-started
	if got := testutil.ToFloat64(metrics.InFlight.WithLabelValues(http.MethodGet, "/slow")); got != 1 {
		t.Fatalf("in-flight = %v while handler is running, want 1", got)
	}
	close(release)
	<-done
}

func TestRegisterMetricsExposesPrometheusHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	registry := prometheus.NewRegistry()
	metrics, err := NewMetrics(registry)
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.Use(metrics.Middleware())
	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	RegisterMetrics(router, metrics)
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health", nil))

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if !strings.Contains(response.Body.String(), "content_hub_http_requests_total") {
		t.Fatal("metrics response does not contain request counter")
	}
}
