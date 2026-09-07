package logger

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestExtractTraceInfoUsesOpenTelemetrySpanContext(t *testing.T) {
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{1},
		SpanID:  trace.SpanID{2},
	})
	ctx := trace.ContextWithSpanContext(context.Background(), spanContext)
	ctx = WithTraceContext(ctx, "request-123", "legacy-trace", "legacy-span")

	fields := extractTraceInfo(ctx)
	values := make(map[string]string, len(fields))
	for _, field := range fields {
		values[field.Key] = field.StringVal
	}
	if values["request_id"] != "request-123" {
		t.Fatalf("request_id = %q, want request-123", values["request_id"])
	}
	if values["trace_id"] != spanContext.TraceID().String() {
		t.Fatalf("trace_id = %q, want %q", values["trace_id"], spanContext.TraceID().String())
	}
	if values["span_id"] != spanContext.SpanID().String() {
		t.Fatalf("span_id = %q, want %q", values["span_id"], spanContext.SpanID().String())
	}
}
