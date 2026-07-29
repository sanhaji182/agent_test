package api

import (
	"net/http"
	"sync"
	"time"
)

// rateLimiter implements a simple token-bucket rate limiter per client IP.
type rateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     int           // requests allowed per window
	window   time.Duration // window duration
}

type bucket struct {
	tokens    int
	lastReset time.Time
}

func newRateLimiter(rate int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		buckets: make(map[string]*bucket),
		rate:    rate,
		window:  window,
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.buckets[key]
	if !exists || now.Sub(b.lastReset) >= rl.window {
		rl.buckets[key] = &bucket{tokens: rl.rate - 1, lastReset: now}
		return true
	}
	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}

// cleanup removes stale buckets to prevent memory growth.
func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	for k, b := range rl.buckets {
		if now.Sub(b.lastReset) > rl.window*2 {
			delete(rl.buckets, k)
		}
	}
}

// rateLimitMiddleware returns 429 Too Many Requests when a client exceeds the limit.
func rateLimitMiddleware(rate int, window time.Duration) func(http.Handler) http.Handler {
	rl := newRateLimiter(rate, window)

	// Periodic cleanup goroutine
	go func() {
		ticker := time.NewTicker(window * 2)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use X-Forwarded-For if behind proxy, else RemoteAddr
			key := r.RemoteAddr
			if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
				key = fwd
			}

			if !rl.allow(key) {
				w.Header().Set("Retry-After", "60")
				writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
