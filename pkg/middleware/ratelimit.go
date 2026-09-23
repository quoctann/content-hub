package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

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
		// Not RemoteAddr: it includes the ephemeral port (a new bucket per TCP
		// connection) and behind nginx it is the proxy's IP, not the client's.
		// ClientIP only trusts X-Forwarded-For from SERVER_TRUSTED_PROXIES.
		ip := c.ClientIP()
		if !getLimiter(ip).Allow() {
			return nil, &server.HTTPError{Code: http.StatusTooManyRequests, Message: "too many requests"}
		}
		return c, nil
	}
}
