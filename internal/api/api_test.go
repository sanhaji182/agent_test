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
	srv := api.NewServer(&config.Config{APIKey: "secret"}, db.NewMemoryStore(), nil)
	w := get(srv, "/api/v1/runs")
	assertStatus(t, w, 401)
}

func TestAPIKeyAuth_ValidKey(t *testing.T) {
	srv := api.NewServer(&config.Config{APIKey: "secret"}, db.NewMemoryStore(), nil)
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

func TestCreateRun_WithTestSpriteMetadata(t *testing.T) {
	srv := newTestServer()
	body := `{
		"project_path":"https://app.example.com",
		"requirements":"test login and checkout",
		"test_type":"ui",
		"prd":"Feature: Login\nUse case: Returning users can sign in\nFeature: Checkout\nUse case: Buyers can pay",
		"auth_type":"login",
		"credentials":"use seeded test account",
		"focus_hints":"checkout happy path",
		"skip_hints":"do not submit real payment"
	}`
	w := post(srv, "/api/v1/runs", body)
	assertStatus(t, w, 202)

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)

	w = get(srv, "/api/v1/runs/"+resp["run_id"])
	assertStatus(t, w, 200)

	var run map[string]interface{}
	json.NewDecoder(w.Body).Decode(&run)
	if run["test_type"] != "ui" {
		t.Fatalf("expected test_type=ui, got %v", run["test_type"])
	}
	if run["auth_type"] != "login" {
		t.Fatalf("expected auth_type=login, got %v", run["auth_type"])
	}
	if run["feature_map"] == nil {
		t.Fatal("expected feature_map")
	}
}

func TestCreateRun_ModelOverrideStored(t *testing.T) {
	srv := newTestServer()
	w := post(srv, "/api/v1/runs", `{"project_path":"/tmp/test","requirements":"login","model":"qoder"}`)
	assertStatus(t, w, 202)

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)

	w = get(srv, "/api/v1/runs/"+resp["run_id"])
	assertStatus(t, w, 200)
	var run map[string]interface{}
	json.NewDecoder(w.Body).Decode(&run)
	if run["model_override"] != "qoder" {
		t.Fatalf("expected model_override=qoder, got %v", run["model_override"])
	}
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

func TestAnalyzeFailure(t *testing.T) {
	srv := newTestServer()
	w := post(srv, "/api/v1/runs", `{"project_path":"/tmp/app","requirements":"checkout failure"}`)
	assertStatus(t, w, 202)
	var cr map[string]string
	json.NewDecoder(w.Body).Decode(&cr)

	w = postNoBody(srv, "/api/v1/runs/"+cr["run_id"]+"/analyze-failure")
	assertStatus(t, w, 200)
	var analysis map[string]interface{}
	json.NewDecoder(w.Body).Decode(&analysis)
	if analysis["run_id"] != cr["run_id"] || analysis["summary"] == "" || analysis["next_action"] == "" {
		t.Fatalf("expected failure analysis payload, got %v", analysis)
	}
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
	if len(evts) < 1 {
		t.Fatalf("expected at least 1 audit event, got %d", len(evts))
	}
}

func TestGetEvents_WithEmit(t *testing.T) {
	srv := newTestServer()
	w := post(srv, "/api/v1/runs", `{"project_path":"/tmp/x"}`)
	var cr map[string]string
	json.NewDecoder(w.Body).Decode(&cr)
	runID := cr["run_id"]

	srv.Events().Emit(runID, events.AnalysisStarted, "analyzing", "Analyzing", nil)
	srv.Events().Emit(runID, events.AnalysisCompleted, "analyzing", "Done", nil)

	w = get(srv, "/api/v1/runs/"+runID+"/events")
	assertStatus(t, w, 200)

	var evts []events.Event
	json.NewDecoder(w.Body).Decode(&evts)
	if len(evts) < 3 {
		t.Fatalf("expected at least 3 events (1 audit + 2 manual), got %d", len(evts))
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
	return api.NewServer(&config.Config{}, db.NewMemoryStore(), nil)
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

func patch(srv *api.Server, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("PATCH", path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	return w
}

func del(srv *api.Server, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("DELETE", path, nil)
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
