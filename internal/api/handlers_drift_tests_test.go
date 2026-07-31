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
)

func TestGenerateDriftTestDisabledWithoutAI(t *testing.T) {
	t.Setenv("GOTEST_AI_PLANNING", "")
	s := NewServer(&config.Config{MaxConcurrentRuns: 1}, db.NewMemoryStore(), nil)
	d := s.drifts.Add(drift.Drift{Repository: "acme/app", Type: drift.TypeMissingTest, FilePath: "a.go", Severity: drift.SeverityHigh})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/drifts/"+d.ID+"/generate-test", nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 without AI, got %d", rec.Code)
	}
}

func TestGenerateDriftTestUnknownDrift(t *testing.T) {
	s := NewServer(&config.Config{MaxConcurrentRuns: 1}, db.NewMemoryStore(), nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drifts/nope/generate-test", nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown drift, got %d", rec.Code)
	}
}

func TestListAndUpdateGeneratedTests(t *testing.T) {
	s := NewServer(&config.Config{MaxConcurrentRuns: 1}, db.NewMemoryStore(), nil)
	d := s.drifts.Add(drift.Drift{Repository: "acme/app", Type: drift.TypeMissingTest, FilePath: "a.go", Severity: drift.SeverityHigh})
	gt := s.driftTests.Add(drift.GeneratedTest{DriftID: d.ID, TestName: "test for a.go", TestCode: "package a"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/drifts/"+d.ID+"/generated-tests", nil)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET generated-tests status = %d", rec.Code)
	}
	var list []drift.GeneratedTest
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(list) != 1 || list[0].ID != gt.ID {
		t.Fatalf("expected 1 generated test %s, got %+v", gt.ID, list)
	}

	req = httptest.NewRequest(http.MethodPatch, "/api/v1/generated-tests/"+gt.ID, strings.NewReader(`{"status":"reviewed"}`))
	rec = httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("PATCH generated-tests status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var updated drift.GeneratedTest
	json.NewDecoder(rec.Body).Decode(&updated)
	if updated.Status != drift.GenStatusReviewed {
		t.Fatalf("expected reviewed, got %s", updated.Status)
	}

	req = httptest.NewRequest(http.MethodPatch, "/api/v1/generated-tests/nope", strings.NewReader(`{"status":"reviewed"}`))
	rec = httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown generated test, got %d", rec.Code)
	}
}
