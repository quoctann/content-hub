package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/quoctann/content-hub/pkg/server"
)

func newRateLimitedEngine(t *testing.T, trustedProxies []string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.RemoteIPHeaders = []string{"X-Forwarded-For"}
	if err := engine.SetTrustedProxies(trustedProxies); err != nil {
		t.Fatal(err)
	}
	router := server.NewGinRouter(engine)
	router.Use(RateLimiter(rate.Every(1<<62), 1)) // one request per key, effectively no refill
	router.GET("/", func(c server.Context) { c.Status(http.StatusOK) })
	return engine
}

func doRequest(engine *gin.Engine, remoteAddr, forwardedFor string) int {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remoteAddr
	if forwardedFor != "" {
		req.Header.Set("X-Forwarded-For", forwardedFor)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec.Code
}

func TestRateLimiterIgnoresSourcePort(t *testing.T) {
	engine := newRateLimitedEngine(t, nil)

	if got := doRequest(engine, "203.0.113.7:40001", ""); got != http.StatusOK {
		t.Fatalf("first request status = %d, want 200", got)
	}
	// Same client on a new TCP connection must share the bucket.
	if got := doRequest(engine, "203.0.113.7:40002", ""); got != http.StatusTooManyRequests {
		t.Fatalf("second request status = %d, want 429", got)
	}
}

func TestRateLimiterUsesForwardedForFromTrustedProxy(t *testing.T) {
	engine := newRateLimitedEngine(t, []string{"10.42.0.0/16"})

	// Two different clients behind the same proxy get separate buckets.
	if got := doRequest(engine, "10.42.0.5:1111", "198.51.100.1, 10.42.0.9"); got != http.StatusOK {
		t.Fatalf("client A status = %d, want 200", got)
	}
	if got := doRequest(engine, "10.42.0.5:2222", "198.51.100.2, 10.42.0.9"); got != http.StatusOK {
		t.Fatalf("client B status = %d, want 200", got)
	}
	// Client A spoofing an extra left-most entry is still resolved to its real IP.
	if got := doRequest(engine, "10.42.0.5:3333", "1.2.3.4, 198.51.100.1, 10.42.0.9"); got != http.StatusTooManyRequests {
		t.Fatalf("spoofed client A status = %d, want 429", got)
	}
}

func TestRateLimiterIgnoresForwardedForFromUntrustedPeer(t *testing.T) {
	engine := newRateLimitedEngine(t, nil)

	if got := doRequest(engine, "203.0.113.7:1", "198.51.100.1"); got != http.StatusOK {
		t.Fatalf("first request status = %d, want 200", got)
	}
	// Rotating X-Forwarded-For must not grant a fresh bucket.
	if got := doRequest(engine, "203.0.113.7:2", "198.51.100.99"); got != http.StatusTooManyRequests {
		t.Fatalf("second request status = %d, want 429", got)
	}
}
