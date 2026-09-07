package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
)

// TracingConfig contains the application-specific OpenTelemetry settings.
// Endpoint is an OTLP/gRPC host:port, for example lgtm.observability.svc:4317.
type TracingConfig struct {
	Endpoint    string
	ServiceName string
	Version     string
	Environment string
	SampleRatio float64
	TLS         bool
}

// InitTracing configures W3C propagation and an OTLP trace exporter. An empty
// endpoint leaves tracing disabled while preserving incoming trace propagation.
func InitTracing(ctx context.Context, cfg TracingConfig) (func(context.Context) error, error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	if cfg.Endpoint == "" {
		otel.SetTracerProvider(trace.NewNoopTracerProvider())
		return func(context.Context) error { return nil }, nil
	}
	if cfg.SampleRatio < 0 || cfg.SampleRatio > 1 {
		return nil, fmt.Errorf("trace sample ratio must be between 0 and 1")
	}

	exporterOptions := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(cfg.Endpoint)}
	if cfg.TLS {
		exporterOptions = append(exporterOptions, otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
	} else {
		exporterOptions = append(exporterOptions, otlptracegrpc.WithInsecure())
	}
	exporter, err := otlptracegrpc.New(ctx, exporterOptions...)
	if err != nil {
		return nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}

	res, err := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceName(cfg.ServiceName),
		semconv.ServiceVersion(cfg.Version),
		semconv.DeploymentEnvironmentName(cfg.Environment),
	))
	if err != nil {
		return nil, fmt.Errorf("create trace resource: %w", err)
	}
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio))),
	)
	otel.SetTracerProvider(provider)
	return provider.Shutdown, nil
}
