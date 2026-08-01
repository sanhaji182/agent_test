package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func runCORS(t *testing.T, allowedOrigins, requestOrigin, method string) *httptest.ResponseRecorder {
	t.Helper()
	handler := newCORSMiddleware(allowedOrigins)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(method, "/api/v1/health", nil)
	if requestOrigin != "" {
		req.Header.Set("Origin", requestOrigin)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestCORS_WildcardDefault(t *testing.T) {
	// Wildcard mode: echo request origin instead of "*" to support credentials
	// (wildcard + credentials is blocked by browsers per the Fetch spec).
	for _, cfg := range []string{"", "*"} {
		erroneousOrigin := "https://evil.example.com"
		rec := runCORS(t, cfg, erroneousOrigin, http.MethodGet)
		// Echo origin instead of wildcard for credential support
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != erroneousOrigin {
			t.Fatalf("config %q: expected echoed origin, got %q", cfg, got)
		}
		// With wildcards we still echo origins, so credentials are allowed
		if rec.Header().Get("Access-Control-Allow-Credentials") != "true" {
			t.Fatalf("config %q: wildcard+echo mode must allow credentials", cfg)
		}
	}
}

func TestCORS_AllowlistedOriginEchoed(t *testing.T) {
	rec := runCORS(t, "https://app.example.com, https://staging.example.com", "https://app.example.com", http.MethodGet)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Fatalf("expected origin echoed, got %q", got)
	}
	if rec.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatal("allowlisted origin should allow credentials")
	}
	if rec.Header().Get("Vary") != "Origin" {
		t.Fatalf("expected Vary: Origin, got %q", rec.Header().Get("Vary"))
	}
}

func TestCORS_UnlistedOriginNotEchoed(t *testing.T) {
	rec := runCORS(t, "https://app.example.com", "https://evil.example.com", http.MethodGet)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unlisted origin must not get ACAO header, got %q", got)
	}
	// Request itself still succeeds (CORS is browser-enforced).
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestCORS_CaseAndTrailingSlashInsensitive(t *testing.T) {
	rec := runCORS(t, "https://App.Example.com/", "https://app.example.com", http.MethodGet)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Fatalf("expected normalized match, got %q", got)
	}
}

func TestCORS_PreflightShortCircuits(t *testing.T) {
	rec := runCORS(t, "https://app.example.com", "https://app.example.com", http.MethodOptions)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for OPTIONS, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Fatal("preflight should include allowed methods")
	}
}
