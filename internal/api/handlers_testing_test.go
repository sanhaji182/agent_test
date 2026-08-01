package api_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/api"
	"github.com/go-go-golems/gotest-agent/internal/config"
	"github.com/go-go-golems/gotest-agent/internal/db"
)

func TestExportCodeHonorsRequestedFramework(t *testing.T) {
	store := db.NewMemoryStore()
	run := &agent.TestRun{
		ID:          "export-run-1",
		ProjectPath: "https://example.com",
		State:       agent.StateDone,
		TestFiles: []agent.TestFile{{
			Name:    "login.json",
			Content: `[{"action":"goto","url":"https://example.com"},{"action":"assert","selector":"h1","assert":"visible"}]`,
		}},
	}
	if err := store.CreateRun(context.Background(), run); err != nil {
		t.Fatalf("create run: %v", err)
	}
	srv := api.NewServer(&config.Config{}, store, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/runs/export-run-1/export-code?language=appium", nil)
	srv.ServeHTTP(w, req)
	assertStatus(t, w, 200)

	var out struct {
		Target    string            `json:"target"`
		Language  string            `json:"language"`
		Framework string            `json:"framework"`
		Code      string            `json:"code"`
		Scripts   map[string]string `json:"scripts"`
	}
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode export response: %v", err)
	}
	if out.Target != "appium" || out.Framework != "Appium WebdriverIO" || out.Language != "javascript" {
		t.Fatalf("unexpected export metadata: %+v", out)
	}
	if _, ok := out.Scripts["login.appium.mjs"]; !ok {
		t.Fatalf("missing appium script: %+v", out.Scripts)
	}
	if out.Code == "" || !containsAll(out.Code, "webdriverio", "appium:automationName") {
		t.Fatalf("unexpected export code: %s", out.Code)
	}
}

func containsAll(value string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(value, part) {
			return false
		}
	}
	return true
}

func TestAdvancedTestingEndpoints_BlockUnsafeBrowserURLs(t *testing.T) {
	srv := newTestServer()
	cases := []struct {
		name string
		path string
		body string
	}{
		{name: "explore metadata", path: "/api/v1/testing/explore", body: `{"url":"http://169.254.169.254/latest/meta-data","max_depth":1}`},
		{name: "performance localhost", path: "/api/v1/testing/performance", body: `{"url":"http://localhost:3000"}`},
		{name: "accessibility private ip", path: "/api/v1/testing/accessibility", body: `{"url":"http://10.0.0.5"}`},
		{name: "visual local tld", path: "/api/v1/testing/visual-regression", body: `{"url":"http://app.local","browsers":["chromium"],"viewports":["desktop"]}`},
		{name: "full audit disallowed scheme", path: "/api/v1/testing/audit", body: `{"url":"ftp://example.com/file"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := post(srv, tc.path, tc.body)
			assertStatus(t, w, 400)

			var out map[string]string
			if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
				t.Fatalf("decode error response: %v", err)
			}
			if out["error"] != "unsafe browser URL" {
				t.Fatalf("expected unsafe browser URL error, got %+v", out)
			}
		})
	}
}
