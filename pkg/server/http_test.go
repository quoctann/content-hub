package server

import (
	"context"
	"testing"

	"github.com/quoctann/content-hub/pkg/logger"
)

func TestAddCorrelationFieldsIncludesClientInfo(t *testing.T) {
	ctx := logger.WithClientInfo(context.Background(), logger.ClientInfo{IP: "198.51.100.1", Country: "US", Ray: "abc123-SJC"})
	entry := map[string]interface{}{}
	addCorrelationFields(entry, ctx)
	for key, want := range map[string]string{"client_ip": "198.51.100.1", "cf_country": "US", "cf_ray": "abc123-SJC"} {
		if entry[key] != want {
			t.Errorf("%s = %v, want %s", key, entry[key], want)
		}
	}
}
