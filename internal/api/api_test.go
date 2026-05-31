package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/api"
	"github.com/go-go-golems/gotest-agent/internal/config"
	"github.com/go-go-golems/gotest-agent/internal/db"
)

func TestHealthEndpoint(t *testing.T) {
	cfg := &config.Config{AppPort: "8080"}
	srv := api.NewServer(cfg, db.NewMemoryStore())

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Fatalf("expected status ok, got %s", resp["status"])
	}
}

func TestAPIKeyAuth_NoKey(t *testing.T) {
	cfg := &config.Config{AppPort: "8080", APIKey: "secret"}
	srv := api.NewServer(cfg, db.NewMemoryStore())

	req := httptest.NewRequest("GET", "/api/v1/runs", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAPIKeyAuth_ValidKey(t *testing.T) {
	cfg := &config.Config{AppPort: "8080", APIKey: "secret"}
	srv := api.NewServer(cfg, db.NewMemoryStore())

	req := httptest.NewRequest("GET", "/api/v1/runs", nil)
	req.Header.Set("X-Api-Key", "secret")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCreateAndGetRun(t *testing.T) {
	cfg := &config.Config{AppPort: "8080"}
	srv := api.NewServer(cfg, db.NewMemoryStore())

	// Create a run
	body := `{"project_path":"/tmp/test","requirements":"test login"}`
	req := httptest.NewRequest("POST", "/api/v1/runs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", w.Code, w.Body.String())
	}

	var createResp map[string]string
	json.NewDecoder(w.Body).Decode(&createResp)
	runID := createResp["run_id"]
	if runID == "" {
		t.Fatal("expected run_id in response")
	}

	// Get the run
	req = httptest.NewRequest("GET", "/api/v1/runs/"+runID, nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var run map[string]interface{}
	json.NewDecoder(w.Body).Decode(&run)
	if run["id"] != runID {
		t.Fatalf("expected id %s, got %v", runID, run["id"])
	}
	if run["state"] != "idle" {
		t.Fatalf("expected state idle, got %v", run["state"])
	}
}

func TestListRuns(t *testing.T) {
	cfg := &config.Config{AppPort: "8080"}
	srv := api.NewServer(cfg, db.NewMemoryStore())

	// Create two runs
	for i := 0; i < 2; i++ {
		body := `{"project_path":"/tmp/test"}`
		req := httptest.NewRequest("POST", "/api/v1/runs", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
	}

	// List runs
	req := httptest.NewRequest("GET", "/api/v1/runs", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var runs []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&runs)
	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs))
	}
}

func TestGetRun_NotFound(t *testing.T) {
	cfg := &config.Config{AppPort: "8080"}
	srv := api.NewServer(cfg, db.NewMemoryStore())

	req := httptest.NewRequest("GET", "/api/v1/runs/nonexistent", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestReport_NotFound(t *testing.T) {
	cfg := &config.Config{AppPort: "8080"}
	srv := api.NewServer(cfg, db.NewMemoryStore())

	req := httptest.NewRequest("GET", "/api/v1/runs/nonexistent/report", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestRerun(t *testing.T) {
	cfg := &config.Config{AppPort: "8080"}
	srv := api.NewServer(cfg, db.NewMemoryStore())

	// Buat run awal
	body := `{"project_path":"/tmp/app","requirements":"login","mode":"simple"}`
	req := httptest.NewRequest("POST", "/api/v1/runs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	var createResp map[string]string
	json.NewDecoder(w.Body).Decode(&createResp)
	origID := createResp["run_id"]

	// Rerun
	req = httptest.NewRequest("POST", "/api/v1/runs/"+origID+"/rerun", nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", w.Code, w.Body.String())
	}

	var rerunResp map[string]string
	json.NewDecoder(w.Body).Decode(&rerunResp)
	if rerunResp["run_id"] == "" || rerunResp["run_id"] == origID {
		t.Fatalf("expected new run_id different from original, got %q", rerunResp["run_id"])
	}
}

func TestRerun_NotFound(t *testing.T) {
	cfg := &config.Config{AppPort: "8080"}
	srv := api.NewServer(cfg, db.NewMemoryStore())

	req := httptest.NewRequest("POST", "/api/v1/runs/nonexistent/rerun", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
