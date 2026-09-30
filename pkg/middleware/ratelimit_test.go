package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/quoctann/content-hub/pkg/logger"
	"github.com/quoctann/content-hub/pkg/server"
)

func newRateLimitedEngine(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(ClientInfoMiddleware())
	router := server.NewGinRouter(engine)
	router.Use(RateLimiter(rate.Every(1<<62), 1)) // one request per key, effectively no refill
	router.GET("/", func(c server.Context) { c.Status(http.StatusOK) })
	return engine
}

func doRequest(engine *gin.Engine, remoteAddr, cloudflareIP, forwardedFor string) int {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remoteAddr
	if cloudflareIP != "" {
		req.Header.Set("CF-Connecting-IP", cloudflareIP)
	}
	if forwardedFor != "" {
		req.Header.Set("X-Forwarded-For", forwardedFor)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec.Code
}

func TestRateLimiterIgnoresSourcePort(t *testing.T) {
	engine := newRateLimitedEngine(t)

	if got := doRequest(engine, "203.0.113.7:40001", "", ""); got != http.StatusOK {
		t.Fatalf("first request status = %d, want 200", got)
	}
	// Same client on a new TCP connection must share the bucket.
	if got := doRequest(engine, "203.0.113.7:40002", "", ""); got != http.StatusTooManyRequests {
		t.Fatalf("second request status = %d, want 429", got)
	}
}

func TestRateLimiterUsesCloudflareIP(t *testing.T) {
	engine := newRateLimitedEngine(t)

	// Two different clients behind the same proxy get separate buckets.
	if got := doRequest(engine, "10.42.0.5:1111", "198.51.100.1", "1.2.3.4"); got != http.StatusOK {
		t.Fatalf("client A status = %d, want 200", got)
	}
	if got := doRequest(engine, "10.42.0.5:2222", "198.51.100.2", "1.2.3.4"); got != http.StatusOK {
		t.Fatalf("client B status = %d, want 200", got)
	}
	// Rotating X-Forwarded-For or the proxy's source port cannot reset A's bucket.
	if got := doRequest(engine, "10.42.0.5:3333", "198.51.100.1", "5.6.7.8"); got != http.StatusTooManyRequests {
		t.Fatalf("client A status = %d, want 429", got)
	}
}

func TestRateLimiterFallsBackToPeerForMissingOrInvalidCloudflareIP(t *testing.T) {
	engine := newRateLimitedEngine(t)

	if got := doRequest(engine, "203.0.113.7:1", "", "198.51.100.1"); got != http.StatusOK {
		t.Fatalf("first request status = %d, want 200", got)
	}
	// Neither a forged forwarding chain nor a malformed Cloudflare value gets a new bucket.
	if got := doRequest(engine, "203.0.113.7:2", "198.51.100.1, 198.51.100.2", "198.51.100.99"); got != http.StatusTooManyRequests {
		t.Fatalf("second request status = %d, want 429", got)
	}
}

func TestRateLimiterNormalizesIPv6(t *testing.T) {
	engine := newRateLimitedEngine(t)
	if got := doRequest(engine, "[::1]:1234", "2001:db8::1", ""); got != http.StatusOK {
		t.Fatalf("first request status = %d, want 200", got)
	}
	if got := doRequest(engine, "[::1]:5678", "2001:0db8:0:0:0:0:0:1", ""); got != http.StatusTooManyRequests {
		t.Fatalf("second request status = %d, want 429", got)
	}
}

func TestRateLimiterRejectsDuplicateCloudflareHeaders(t *testing.T) {
	engine := newRateLimitedEngine(t)
	if got := doRequest(engine, "203.0.113.7:1", "", ""); got != http.StatusOK {
		t.Fatalf("first request status = %d, want 200", got)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.7:2"
	req.Header.Add("CF-Connecting-IP", "198.51.100.1")
	req.Header.Add("CF-Connecting-IP", "198.51.100.2")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("duplicate header status = %d, want 429", rec.Code)
	}
}

func TestClientInfoMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(ClientInfoMiddleware())
	var got logger.ClientInfo
	engine.GET("/", func(c *gin.Context) {
		got = logger.ClientInfoFromContext(c.Request.Context())
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.42.0.5:1234"
	req.Header.Set("CF-Connecting-IP", "2001:0db8::1")
	req.Header.Set("CF-IPCountry", "US")
	req.Header.Set("CF-Ray", "abc123-SJC")
	engine.ServeHTTP(httptest.NewRecorder(), req)
	if got != (logger.ClientInfo{IP: "2001:db8::1", Country: "US", Ray: "abc123-SJC"}) {
		t.Fatalf("client info = %+v", got)
	}

	// Metadata without a valid Cloudflare IP is not attributed to the peer.
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.42.0.5:5678"
	req.Header.Set("CF-Connecting-IP", "invalid")
	req.Header.Set("CF-IPCountry", "US")
	req.Header.Set("CF-Ray", "abc123-SJC")
	engine.ServeHTTP(httptest.NewRecorder(), req)
	if got != (logger.ClientInfo{IP: "10.42.0.5"}) {
		t.Fatalf("fallback client info = %+v", got)
	}
}
