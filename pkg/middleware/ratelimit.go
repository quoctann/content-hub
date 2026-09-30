package middleware

import (
	"net"
	"net/http"
	"net/netip"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/quoctann/content-hub/pkg/logger"
	"github.com/quoctann/content-hub/pkg/server"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter returns a middleware that limits requests per IP address.
// r is the number of events per second, b is the burst size.
func RateLimiter(r rate.Limit, b int) server.MiddlewareFunc {
	var mu sync.Mutex
	limiters := make(map[string]*ipLimiter)

	// Background cleanup of stale entries every minute.
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, il := range limiters {
				if time.Since(il.lastSeen) > 5*time.Minute {
					delete(limiters, ip)
				}
			}
			mu.Unlock()
		}
	}()

	getLimiter := func(ip string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()
		il, ok := limiters[ip]
		if !ok {
			il = &ipLimiter{limiter: rate.NewLimiter(r, b)}
			limiters[ip] = il
		}
		il.lastSeen = time.Now()
		return il.limiter
	}

	return func(c server.Context) (server.Context, error) {
		ip := logger.ClientInfoFromContext(c.Request().Context()).IP
		if ip == "" {
			ip, _ = clientIP(c.Request())
		}
		if !getLimiter(ip).Allow() {
			return nil, &server.HTTPError{Code: http.StatusTooManyRequests, Message: "too many requests"}
		}
		return c, nil
	}
}

// clientIP uses the client IP supplied by Cloudflare through the tunnel.
// This assumes the origin is reachable only through Cloudflare and intermediate
// proxies preserve its header. For local or malformed requests, use the TCP peer.
func clientIP(req *http.Request) (string, bool) {
	if values := req.Header.Values("CF-Connecting-IP"); len(values) == 1 {
		if ip, err := netip.ParseAddr(values[0]); err == nil {
			return ip.Unmap().String(), true
		}
	}
	peer, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		peer = req.RemoteAddr
	}
	if ip, err := netip.ParseAddr(peer); err == nil {
		return ip.Unmap().String(), false
	}
	return peer, false
}
