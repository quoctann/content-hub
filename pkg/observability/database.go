package observability

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

// RegisterDatabasePoolMetrics exposes the application's pgx pool state. The
// database exporter remains responsible for server-side PostgreSQL metrics.
func RegisterDatabasePoolMetrics(reg prometheus.Registerer, pool *pgxpool.Pool) error {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	collectors := []prometheus.Collector{
		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace: "content_hub", Subsystem: "database", Name: "pool_acquired_connections", Help: "Connections currently acquired from the pool.",
		}, func() float64 { return float64(pool.Stat().AcquiredConns()) }),
		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace: "content_hub", Subsystem: "database", Name: "pool_idle_connections", Help: "Idle connections in the pool.",
		}, func() float64 { return float64(pool.Stat().IdleConns()) }),
		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace: "content_hub", Subsystem: "database", Name: "pool_total_connections", Help: "Total connections in the pool.",
		}, func() float64 { return float64(pool.Stat().TotalConns()) }),
		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace: "content_hub", Subsystem: "database", Name: "pool_max_connections", Help: "Maximum configured pool connections.",
		}, func() float64 { return float64(pool.Stat().MaxConns()) }),
	}
	for _, collector := range collectors {
		if err := reg.Register(collector); err != nil {
			return fmt.Errorf("register database pool metric: %w", err)
		}
	}
	return nil
}
