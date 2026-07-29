package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := newRateLimiter(5, time.Minute)
	for i := 0; i < 5; i++ {
		if !rl.allow("client-1") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := newRateLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		rl.allow("client-1")
	}
	if rl.allow("client-1") {
		t.Fatal("4th request should be blocked")
	}
	// Different client should still be allowed
	if !rl.allow("client-2") {
		t.Fatal("different client should not be rate limited")
	}
}

func TestRateLimiter_ResetsAfterWindow(t *testing.T) {
	rl := newRateLimiter(2, 50*time.Millisecond)
	rl.allow("c")
	rl.allow("c")
	if rl.allow("c") {
		t.Fatal("should be blocked before window reset")
	}
	time.Sleep(60 * time.Millisecond)
	if !rl.allow("c") {
		t.Fatal("should be allowed after window reset")
	}
}

func TestRateLimitMiddleware_Returns429(t *testing.T) {
	handler := rateLimitMiddleware(2, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First 2 requests pass
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: got %d, want 200", i+1, rr.Code)
		}
	}

	// 3rd request gets 429
	req := httptest.NewRequest("GET", "/health", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("got %d, want 429", rr.Code)
	}
	if rr.Header().Get("Retry-After") != "60" {
		t.Fatal("missing Retry-After header")
	}
}

func TestRateLimitMiddleware_XForwardedFor(t *testing.T) {
	handler := rateLimitMiddleware(1, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request via proxy with X-Forwarded-For
	req := httptest.NewRequest("GET", "/health", nil)
	req.RemoteAddr = "10.0.0.1:9999"
	req.Header.Set("X-Forwarded-For", "203.0.113.5")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("first request: got %d", rr.Code)
	}

	// Same forwarded IP should be blocked
	req2 := httptest.NewRequest("GET", "/health", nil)
	req2.RemoteAddr = "10.0.0.2:9999" // different proxy hop
	req2.Header.Set("X-Forwarded-For", "203.0.113.5")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("same forwarded IP should be blocked, got %d", rr2.Code)
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := newRateLimiter(10, 10*time.Millisecond)
	rl.allow("old-client")
	time.Sleep(25 * time.Millisecond)
	rl.cleanup()
	rl.mu.Lock()
	_, exists := rl.buckets["old-client"]
	rl.mu.Unlock()
	if exists {
		t.Fatal("stale bucket should be cleaned up")
	}
}
