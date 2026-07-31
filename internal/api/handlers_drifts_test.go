package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/config"
	"github.com/go-go-golems/gotest-agent/internal/db"
	"github.com/go-go-golems/gotest-agent/internal/drift"
	"github.com/go-go-golems/gotest-agent/internal/webhook"
)

func TestDriftEndpoints(t *testing.T) {
	s := NewServer(&config.Config{MaxConcurrentRuns: 1, AppEnv: "development"}, db.NewMemoryStore(), nil)

	created := s.drifts.Add(drift.Drift{
		Repository: "acme/app",
		Type:       drift.TypeMissingTest,
		FilePath:   "internal/api/server.go",
		Severity:   drift.SeverityHigh,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/drifts?repository=acme/app", nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /drifts status = %d", rec.Code)
	}
	var list []drift.Drift
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("expected 1 drift with ID %s, got %+v", created.ID, list)
	}

	body := strings.NewReader(`{"status":"fixed"}`)
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/drifts/"+created.ID, body)
	rec = httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("PATCH /drifts/{id} status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var updated drift.Drift
	if err := json.NewDecoder(rec.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated: %v", err)
	}
	if updated.Status != drift.StatusFixed {
		t.Fatalf("expected status fixed, got %s", updated.Status)
	}

	req = httptest.NewRequest(http.MethodPatch, "/api/v1/drifts/nope", strings.NewReader(`{"status":"fixed"}`))
	rec = httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown drift, got %d", rec.Code)
	}
}

func TestDetectDriftFromPush(t *testing.T) {
	s := NewServer(&config.Config{MaxConcurrentRuns: 1}, db.NewMemoryStore(), nil)

	s.detectDriftFromPush(webhook.PushEvent{
		Ref:        "refs/heads/main",
		Repository: webhook.Repository{FullName: "acme/app"},
		Commits: []webhook.Commit{
			{ID: "abc", Modified: []string{"internal/service/user.go"}},
		},
	})

	list := s.drifts.List("acme/app", drift.TypeMissingTest, "")
	if len(list) != 1 {
		t.Fatalf("expected 1 missing_test drift, got %d", len(list))
	}
	if list[0].FilePath != "internal/service/user.go" {
		t.Fatalf("unexpected file path: %s", list[0].FilePath)
	}
}
