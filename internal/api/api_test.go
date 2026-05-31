package api_test

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/api"
	"github.com/go-go-golems/gotest-agent/internal/config"
	"github.com/go-go-golems/gotest-agent/internal/db"
	"github.com/go-go-golems/gotest-agent/internal/events"
)

func TestHealthEndpoint(t *testing.T) {
	srv := newTestServer()
	w := get(srv, "/health")
	assertStatus(t, w, 200)
}

func TestAPIKeyAuth_NoKey(t *testing.T) {
	srv := api.NewServer(&config.Config{APIKey: "secret"}, db.NewMemoryStore())
	w := get(srv, "/api/v1/runs")
	assertStatus(t, w, 401)
}

func TestAPIKeyAuth_ValidKey(t *testing.T) {
	srv := api.NewServer(&config.Config{APIKey: "secret"}, db.NewMemoryStore())
	req := httptest.NewRequest("GET", "/api/v1/runs", nil)
	req.Header.Set("X-Api-Key", "secret")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	assertStatus(t, w, 200)
}

func TestCreateAndGetRun(t *testing.T) {
	srv := newTestServer()
	w := post(srv, "/api/v1/runs", `{"project_path":"/tmp/test","requirements":"login"}`)
	assertStatus(t, w, 202)

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	runID := resp["run_id"]
	if runID == "" {
		t.Fatal("expected run_id")
	}

	w = get(srv, "/api/v1/runs/"+runID)
	assertStatus(t, w, 200)
}

func TestListRuns(t *testing.T) {
	srv := newTestServer()
	post(srv, "/api/v1/runs", `{"project_path":"/tmp/a"}`)
	post(srv, "/api/v1/runs", `{"project_path":"/tmp/b"}`)

	w := get(srv, "/api/v1/runs")
	assertStatus(t, w, 200)

	var runs []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&runs)
	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs))
	}
}

func TestGetRun_NotFound(t *testing.T) {
	srv := newTestServer()
	w := get(srv, "/api/v1/runs/nonexistent")
	assertStatus(t, w, 404)
}

func TestReport_NotFound(t *testing.T) {
	srv := newTestServer()
	w := get(srv, "/api/v1/runs/nonexistent/report")
	assertStatus(t, w, 404)
}

func TestRerun(t *testing.T) {
	srv := newTestServer()
	w := post(srv, "/api/v1/runs", `{"project_path":"/tmp/app","requirements":"login","mode":"simple"}`)
	var cr map[string]string
	json.NewDecoder(w.Body).Decode(&cr)

	w = postNoBody(srv, "/api/v1/runs/"+cr["run_id"]+"/rerun")
	assertStatus(t, w, 202)

	var rr map[string]string
	json.NewDecoder(w.Body).Decode(&rr)
	if rr["run_id"] == "" || rr["run_id"] == cr["run_id"] {
		t.Fatal("expected new run_id")
	}
}

func TestRerun_NotFound(t *testing.T) {
	srv := newTestServer()
	w := postNoBody(srv, "/api/v1/runs/nonexistent/rerun")
	assertStatus(t, w, 404)
}

// --- Events ---

func TestGetEvents_Empty(t *testing.T) {
	srv := newTestServer()
	w := post(srv, "/api/v1/runs", `{"project_path":"/tmp/x"}`)
	var cr map[string]string
	json.NewDecoder(w.Body).Decode(&cr)

	w = get(srv, "/api/v1/runs/"+cr["run_id"]+"/events")
	assertStatus(t, w, 200)

	var evts []events.Event
	json.NewDecoder(w.Body).Decode(&evts)
	if len(evts) != 0 {
		t.Fatalf("expected 0 events, got %d", len(evts))
	}
}

func TestGetEvents_WithEmit(t *testing.T) {
	srv := newTestServer()
	w := post(srv, "/api/v1/runs", `{"project_path":"/tmp/x"}`)
	var cr map[string]string
	json.NewDecoder(w.Body).Decode(&cr)
	runID := cr["run_id"]

	// Emit events via the store
	srv.Events().Emit(runID, events.RunStarted, "idle", "Run started", nil)
	srv.Events().Emit(runID, events.AnalysisStarted, "analyzing", "Analyzing", nil)

	w = get(srv, "/api/v1/runs/"+runID+"/events")
	assertStatus(t, w, 200)

	var evts []events.Event
	json.NewDecoder(w.Body).Decode(&evts)
	if len(evts) != 2 {
		t.Fatalf("expected 2 events, got %d", len(evts))
	}
	if evts[0].Type != events.RunStarted {
		t.Fatalf("expected run_started, got %s", evts[0].Type)
	}
}

// --- Compare ---

func TestCompare(t *testing.T) {
	srv := newTestServer()
	w := post(srv, "/api/v1/runs", `{"project_path":"/tmp/a"}`)
	var cr1 map[string]string
	json.NewDecoder(w.Body).Decode(&cr1)

	w = post(srv, "/api/v1/runs", `{"project_path":"/tmp/b"}`)
	var cr2 map[string]string
	json.NewDecoder(w.Body).Decode(&cr2)

	w = get(srv, "/api/v1/runs/"+cr1["run_id"]+"/compare/"+cr2["run_id"])
	assertStatus(t, w, 200)

	var result map[string]interface{}
	json.NewDecoder(w.Body).Decode(&result)
	if result["run_a"] != cr1["run_id"] {
		t.Fatalf("expected run_a=%s, got %v", cr1["run_id"], result["run_a"])
	}
	if result["summary"] == nil || result["summary"] == "" {
		t.Fatal("expected non-empty summary")
	}
}

func TestCompare_NotFound(t *testing.T) {
	srv := newTestServer()
	w := get(srv, "/api/v1/runs/aaa/compare/bbb")
	assertStatus(t, w, 404)
}

// --- Recordings ---

func TestRecordings_Empty(t *testing.T) {
	srv := newTestServer()
	w := post(srv, "/api/v1/runs", `{"project_path":"/tmp/x"}`)
	var cr map[string]string
	json.NewDecoder(w.Body).Decode(&cr)

	w = get(srv, "/api/v1/runs/"+cr["run_id"]+"/recordings")
	assertStatus(t, w, 200)

	var recs []interface{}
	json.NewDecoder(w.Body).Decode(&recs)
	if len(recs) != 0 {
		t.Fatalf("expected 0 recordings, got %d", len(recs))
	}
}

func TestRecordings_All(t *testing.T) {
	srv := newTestServer()
	w := get(srv, "/api/v1/recordings")
	assertStatus(t, w, 200)
}

// --- Visual ---

func TestVisualArtifacts_Empty(t *testing.T) {
	srv := newTestServer()
	w := post(srv, "/api/v1/runs", `{"project_path":"/tmp/x"}`)
	var cr map[string]string
	json.NewDecoder(w.Body).Decode(&cr)

	w = get(srv, "/api/v1/runs/"+cr["run_id"]+"/visual")
	assertStatus(t, w, 200)

	var arts []interface{}
	json.NewDecoder(w.Body).Decode(&arts)
	if len(arts) != 0 {
		t.Fatalf("expected 0 artifacts, got %d", len(arts))
	}
}

// --- Helpers ---

func newTestServer() *api.Server {
	return api.NewServer(&config.Config{}, db.NewMemoryStore())
}

func get(srv *api.Server, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", path, nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	return w
}

func post(srv *api.Server, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	return w
}

func postNoBody(srv *api.Server, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", path, nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	return w
}

func assertStatus(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if w.Code != expected {
		t.Fatalf("expected %d, got %d: %s", expected, w.Code, w.Body.String())
	}
}
