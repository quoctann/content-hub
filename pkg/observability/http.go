// Package observability contains framework adapters for application telemetry.
package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const unmatchedRoute = "unmatched"

// Metrics contains the Prometheus instruments used by the HTTP middleware.
// The instruments are public to allow callers to inspect or compose them,
// while registration remains controlled by NewMetrics.
type Metrics struct {
	Requests        *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	InFlight        *prometheus.GaugeVec
	handler         http.Handler
}

// NewMetrics creates and registers HTTP metrics with reg. A nil registerer
// uses the default Prometheus registerer.
func NewMetrics(reg prometheus.Registerer) (*Metrics, error) {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	m := &Metrics{
		Requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "content_hub",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests.",
		}, []string{"method", "route", "status"}),
		RequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "content_hub",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
		}, []string{"method", "route", "status"}),
		InFlight: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "content_hub",
			Subsystem: "http",
			Name:      "requests_in_flight",
			Help:      "Current number of HTTP requests in flight.",
		}, []string{"method", "route"}),
	}

	for _, collector := range []prometheus.Collector{m.Requests, m.RequestDuration, m.InFlight} {
		if err := reg.Register(collector); err != nil {
			return nil, err
		}
	}
	gatherer := prometheus.DefaultGatherer
	if customGatherer, ok := reg.(prometheus.Gatherer); ok {
		gatherer = customGatherer
	}
	m.handler = promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{})
	return m, nil
}

// Middleware returns Gin middleware that records completed requests and
// tracks requests currently being handled. Route labels use Gin's templated
// FullPath rather than the raw URL to keep cardinality bounded.
func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		// FullPath is available before Next for matched routes.
		route := c.FullPath()
		if route == "" {
			route = unmatchedRoute
		}
		m.InFlight.WithLabelValues(method, route).Inc()
		started := time.Now()
		c.Next()
		m.InFlight.WithLabelValues(method, route).Dec()

		status := strconv.Itoa(c.Writer.Status())
		m.Requests.WithLabelValues(method, route, status).Inc()
		m.RequestDuration.WithLabelValues(method, route, status).Observe(time.Since(started).Seconds())
	}
}

// Handler returns the Prometheus exposition handler for these metrics.
func (m *Metrics) Handler() http.Handler { return m.handler }

// RegisterMetrics registers the standard /metrics endpoint on a Gin router.
func RegisterMetrics(router gin.IRouter, metrics *Metrics) {
	router.GET("/metrics", gin.WrapH(metrics.Handler()))
}
