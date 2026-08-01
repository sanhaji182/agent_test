package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/config"
	"github.com/go-go-golems/gotest-agent/internal/db"
)

func TestRecordingSessionEndpoints(t *testing.T) {
	s := NewServer(&config.Config{MaxConcurrentRuns: 1}, db.NewMemoryStore(), nil)

	// Create session
	resp := postJSON(s.router, "/api/v1/recording-sessions", `{"name":"Login Flow","project_path":"/tmp/p","base_url":"https://example.com"}`)
	if resp.Code != 201 {
		t.Fatalf("POST /recording-sessions: %d, body: %s", resp.Code, resp.Body.String())
	}
	var sess map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&sess)
	sessionID := sess["id"].(string)
	if sessionID == "" {
		t.Fatal("expected session ID")
	}

	// List sessions
	resp = getJSON(s.router, "/api/v1/recording-sessions")
	if resp.Code != 200 {
		t.Fatalf("GET /recording-sessions: %d", resp.Code)
	}

	// Get session with events
	resp = getJSON(s.router, "/api/v1/recording-sessions/"+sessionID)
	if resp.Code != 200 {
		t.Fatalf("GET /recording-sessions/{id}: %d", resp.Code)
	}

	// Add event
	resp = postJSON(s.router, "/api/v1/recording-sessions/"+sessionID+"/events", `{"event_type":"click","selector":"#btn"}`)
	if resp.Code != 201 {
		t.Fatalf("POST events: %d, body: %s", resp.Code, resp.Body.String())
	}

	// List events
	resp = getJSON(s.router, "/api/v1/recording-sessions/"+sessionID+"/events")
	if resp.Code != 200 {
		t.Fatalf("GET events: %d", resp.Code)
	}

	// Generate test
	resp = postJSON(s.router, "/api/v1/recording-sessions/"+sessionID+"/generate", "")
	if resp.Code != 200 {
		t.Fatalf("POST generate: %d, body: %s", resp.Code, resp.Body.String())
	}

	// Update session
	resp = patchJSON(s.router, "/api/v1/recording-sessions/"+sessionID, `{"name":"Updated Flow","status":"completed"}`)
	if resp.Code != 200 {
		t.Fatalf("PATCH session: %d", resp.Code)
	}

	// Delete session
	resp = delJSON(s.router, "/api/v1/recording-sessions/"+sessionID)
	if resp.Code != 204 {
		t.Fatalf("DELETE session: %d", resp.Code)
	}

	// Verify deleted
	resp = getJSON(s.router, "/api/v1/recording-sessions/"+sessionID)
	if resp.Code != 404 {
		t.Fatalf("expected 404 after delete, got %d", resp.Code)
	}
}

func TestRecordingSessionValidation(t *testing.T) {
	s := NewServer(&config.Config{MaxConcurrentRuns: 1}, db.NewMemoryStore(), nil)

	// Missing name
	resp := postJSON(s.router, "/api/v1/recording-sessions", `{"project_path":"/tmp/p","base_url":"https://example.com"}`)
	if resp.Code != 400 {
		t.Errorf("expected 400 for missing name, got %d", resp.Code)
	}

	// Add event to nonexistent session
	resp = postJSON(s.router, "/api/v1/recording-sessions/nope/events", `{"event_type":"click"}`)
	if resp.Code != 404 {
		t.Errorf("expected 404 for unknown session, got %d", resp.Code)
	}
}

func TestWebhookRegistrationEndpoints(t *testing.T) {
	s := NewServer(&config.Config{MaxConcurrentRuns: 1}, db.NewMemoryStore(), nil)

	// Register webhook
	resp := postJSON(s.router, "/api/v1/webhooks/register", `{"repository_url":"https://github.com/owner/repo","github_token":"ghp_test"}`)
	if resp.Code != 201 {
		t.Fatalf("POST /webhooks/register: %d", resp.Code)
	}
	var reg map[string]string
	json.NewDecoder(resp.Body).Decode(&reg)
	webhookID := reg["webhook_id"]

	// List webhooks
	resp = getJSON(s.router, "/api/v1/webhooks")
	if resp.Code != 200 {
		t.Fatalf("GET /webhooks: %d", resp.Code)
	}

	// Get webhook
	resp = getJSON(s.router, "/api/v1/webhooks/"+webhookID)
	if resp.Code != 200 {
		t.Fatalf("GET /webhooks/{id}: %d", resp.Code)
	}

	// Update status
	resp = patchJSON(s.router, "/api/v1/webhooks/"+webhookID+"/status", `{"status":"inactive"}`)
	if resp.Code != 200 {
		t.Fatalf("PATCH webhook status: %d", resp.Code)
	}

	// Sync
	resp = postJSON(s.router, "/api/v1/webhooks/"+webhookID+"/sync", "")
	if resp.Code != 200 {
		t.Fatalf("POST webhook sync: %d", resp.Code)
	}

	// Delete
	resp = delJSON(s.router, "/api/v1/webhooks/"+webhookID)
	if resp.Code != 204 {
		t.Fatalf("DELETE webhook: %d", resp.Code)
	}
}

func TestWebhookNotFound(t *testing.T) {
	s := NewServer(&config.Config{MaxConcurrentRuns: 1}, db.NewMemoryStore(), nil)

	resp := getJSON(s.router, "/api/v1/webhooks/nope")
	if resp.Code != 404 {
		t.Errorf("expected 404, got %d", resp.Code)
	}
}

// ---- HTTP helpers ----
func getJSON(mux http.Handler, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", "http://example.com"+path, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

func postJSON(mux http.Handler, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", "http://example.com"+path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

func patchJSON(mux http.Handler, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("PATCH", "http://example.com"+path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

func delJSON(mux http.Handler, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("DELETE", "http://example.com"+path, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}
