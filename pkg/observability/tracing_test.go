package observability

import (
	"context"
	"testing"
)

func TestInitTracingRejectsInvalidSampleRatio(t *testing.T) {
	_, err := InitTracing(context.Background(), TracingConfig{
		Endpoint:    "tempo:4317",
		ServiceName: "content-hub",
		SampleRatio: 1.1,
	})
	if err == nil {
		t.Fatal("InitTracing() error = nil, want invalid sample ratio error")
	}
}

func TestInitTracingWithoutEndpointIsNoop(t *testing.T) {
	shutdown, err := InitTracing(context.Background(), TracingConfig{})
	if err != nil {
		t.Fatalf("InitTracing() error = %v", err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown() error = %v", err)
	}
}
