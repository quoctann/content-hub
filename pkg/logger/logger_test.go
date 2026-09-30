package logger

import (
	"context"
	"testing"
)

func TestExtractTraceInfoIncludesClientInfo(t *testing.T) {
	ctx := WithClientInfo(context.Background(), ClientInfo{IP: "198.51.100.1", Country: "US", Ray: "abc123-SJC"})
	fields := extractTraceInfo(ctx)
	want := map[string]string{"client_ip": "198.51.100.1", "cf_country": "US", "cf_ray": "abc123-SJC"}
	for _, field := range fields {
		if val, ok := want[field.Key]; ok {
			if field.StringVal != val {
				t.Fatalf("%s = %q, want %q", field.Key, field.StringVal, val)
			}
			delete(want, field.Key)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing fields: %v", want)
	}
}
