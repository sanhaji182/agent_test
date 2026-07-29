package api_test

// End-to-end product smoke test (real LLM + real Playwright).
//
// Gated twice so it never runs in normal suites:
//  1. Requires the marker file .e2e-enable at the repo root (gitignored path).
//  2. Requires LLM credentials in the repo-root .env.
//
// Run: go test ./internal/api/ -run TestE2E_ProductSmoke -v -timeout 15m
//
// This is the empirical answer to "does the core product loop work":
// create run -> LLM analyzes + generates plan/scripts -> Playwright executes
// -> artifacts (events, test files, run_result, HTML report) appear.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/api"
	"github.com/go-go-golems/gotest-agent/internal/config"
	"github.com/go-go-golems/gotest-agent/internal/db"
)

func loadDotEnv(t *testing.T) map[string]string {
	t.Helper()
	// Repo root is two levels up from internal/api.
	data, err := os.ReadFile(filepath.Join("..", "..", ".env"))
	if err != nil {
		t.Skipf("no .env at repo root: %v", err)
	}
	env := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if k, v, ok := strings.Cut(line, "="); ok {
			env[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return env
}

func TestE2E_ProductSmoke(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E in -short mode")
	}
	if _, err := os.Stat(filepath.Join("..", "..", ".e2e-enable")); err != nil {
		t.Skip("E2E disabled: create .e2e-enable at repo root to run")
	}
	env := loadDotEnv(t)
	apiKey := env["LLM_API_KEY"]
	if apiKey == "" {
		apiKey = env["ANTHROPIC_API_KEY"]
	}
	if apiKey == "" {
		t.Skip("no LLM_API_KEY/ANTHROPIC_API_KEY in .env")
	}

	cfg := config.Load()
	// Playwright's default CDN (azureedge.net) is retired. If the run reaches
	// execution and needs the browser driver, set PLAYWRIGHT_DOWNLOAD_HOST to a
	// working mirror (or pre-install the driver) before running this test —
	// playwright.Install() honors that env var natively.
	cfg.LLMProvider = env["LLM_PROVIDER"]
	cfg.LLMBaseURL = env["LLM_BASE_URL"]
	cfg.LLMModel = env["LLM_MODEL"]
	cfg.AnthropicAPIKey = apiKey // buildAgent passes this as the api key
	cfg.AppEnv = "development"

	srv := api.NewServer(cfg, db.NewMemoryStore(), nil)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	// --- Step 1: provider test-connection (ADR-006 manual check) ---
	t.Run("provider test-connection", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"provider": cfg.LLMProvider,
			"model":    cfg.LLMModel,
			"api_key":  apiKey,
			"base_url": cfg.LLMBaseURL,
		})
		resp, err := http.Post(ts.URL+"/api/v1/ai/test-provider", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("test-provider request: %v", err)
		}
		defer resp.Body.Close()
		var out map[string]any
		json.NewDecoder(resp.Body).Decode(&out)
		if ok, _ := out["success"].(bool); !ok {
			t.Fatalf("test-provider failed: %+v", out)
		}
		t.Logf("provider %s/%s reachable ✓", cfg.LLMProvider, cfg.LLMModel)
	})

	// --- Step 2: full run against a real page ---
	target := os.Getenv("E2E_TARGET_URL")
	if target == "" {
		target = "https://example.com"
	}

	body, _ := json.Marshal(map[string]string{
		"project_path": target,
		"requirements": "Open the page, verify it loads, check the main heading is visible, and confirm at least one link is present.",
		"mode":         "simple",
	})
	resp, err := http.Post(ts.URL+"/api/v1/runs", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("create run: %v", err)
	}
	var created map[string]string
	json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()
	runID := created["run_id"]
	if runID == "" {
		t.Fatalf("no run_id in response: %+v", created)
	}
	t.Logf("run created: %s (target %s)", runID, target)

	// --- Step 3: poll to terminal state (Playwright download can be slow) ---
	deadline := time.Now().Add(8 * time.Minute)
	state := ""
	var run map[string]any
	for time.Now().Before(deadline) {
		r, err := http.Get(ts.URL + "/api/v1/runs/" + runID)
		if err == nil {
			run = map[string]any{}
			json.NewDecoder(r.Body).Decode(&run)
			r.Body.Close()
			state, _ = run["state"].(string)
			if state == "done" || state == "failed" || state == "simulated" {
				break
			}
		}
		time.Sleep(3 * time.Second)
	}
	t.Logf("terminal state: %q", state)

	// --- Step 4: artifacts ---
	evResp, err := http.Get(ts.URL + "/api/v1/runs/" + runID + "/events")
	if err != nil {
		t.Fatalf("get events: %v", err)
	}
	var events []map[string]any
	json.NewDecoder(evResp.Body).Decode(&events)
	evResp.Body.Close()
	t.Logf("events emitted: %d", len(events))
	for _, e := range events {
		t.Logf("  [%v] %v — %v", e["type"], e["phase"], e["message"])
	}

	repResp, err := http.Get(ts.URL + "/api/v1/runs/" + runID + "/report")
	if err != nil {
		t.Fatalf("get report: %v", err)
	}
	reportHTML := new(bytes.Buffer)
	reportHTML.ReadFrom(repResp.Body)
	repResp.Body.Close()

	// --- Assertions ---
	if state != "done" && state != "failed" {
		t.Errorf("run did not reach done/failed, state=%q (stuck or simulated)", state)
	}
	if len(events) < 3 {
		t.Errorf("too few events (%d) — pipeline likely short-circuited", len(events))
	}
	if tf, ok := run["test_files"].([]any); !ok || len(tf) == 0 {
		t.Errorf("no test_files generated — LLM generation step failed")
	} else {
		t.Logf("test files generated: %d", len(tf))
		if content, ok := tf[0].(map[string]any)["content"].(string); ok {
			preview := content
			if len(preview) > 400 {
				preview = preview[:400] + "..."
			}
			t.Logf("first file preview:\n%s", preview)
		}
	}
	if !strings.Contains(reportHTML.String(), "<html") {
		t.Errorf("HTML report did not render")
	}
	if rr, ok := run["run_result"].(map[string]any); ok {
		t.Logf("run_result: passed=%v failed=%v total=%v", rr["passed"], rr["failed"], rr["total"])
	} else if state == "done" {
		t.Errorf("state=done but no run_result")
	}
	if errMsg, _ := run["error"].(string); errMsg != "" {
		t.Logf("run error field: %s", errMsg)
	}
	fmt.Println("E2E smoke complete — state:", state)
}
