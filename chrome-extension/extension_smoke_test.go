package chrome_extension

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
)

func TestChromeExtensionRecordFlowSmoke(t *testing.T) {
	if os.Getenv("GOTEST_CHROME_EXTENSION_SMOKE") != "1" {
		t.Skip("set GOTEST_CHROME_EXTENSION_SMOKE=1 to run live Chrome extension browser smoke test")
	}

	backend := newRecordingBackendFixture(t)
	defer backend.Close()

	extensionDir, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("extension abs path: %v", err)
	}
	userDataDir := t.TempDir()

	if err := playwright.Install(); err != nil {
		t.Fatalf("install Playwright browsers: %v", err)
	}
	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("start Playwright: %v", err)
	}
	defer pw.Stop()

	ctx, err := pw.Chromium.LaunchPersistentContext(userDataDir, playwright.BrowserTypeLaunchPersistentContextOptions{
		Headless: playwright.Bool(false),
		Args: []string{
			"--disable-extensions-except=" + extensionDir,
			"--load-extension=" + extensionDir,
		},
		IgnoreDefaultArgs: []string{"--disable-extensions"},
	})
	if err != nil {
		t.Fatalf("launch persistent Chromium context with extension: %v", err)
	}
	defer ctx.Close()

	extensionID := waitForExtensionID(t, ctx)
	popup, err := ctx.NewPage()
	if err != nil {
		t.Fatalf("new popup page: %v", err)
	}
	if _, err := popup.Goto(fmt.Sprintf("chrome-extension://%s/popup.html", extensionID), playwright.PageGotoOptions{Timeout: playwright.Float(5000)}); err != nil {
		t.Fatalf("open extension popup: %v", err)
	}

	mustFill(t, popup, "#backendUrl", backend.URL)
	mustFill(t, popup, "#apiKey", recordingFixtureAPIKey)
	mustClick(t, popup, "#saveSettings")
	waitForText(t, popup, "#saveSettings", "Saved!", 5*time.Second)

	mustFill(t, popup, "#sessionName", "Smoke Recording")
	mustFill(t, popup, "#baseUrl", backend.URL+"/app")
	mustClick(t, popup, "#startBtn")
	waitForText(t, popup, "#status", "Recording", 8*time.Second)

	app, err := ctx.NewPage()
	if err != nil {
		t.Fatalf("new app page: %v", err)
	}
	if _, err := app.Goto(backend.URL+"/app", playwright.PageGotoOptions{Timeout: playwright.Float(5000)}); err != nil {
		t.Fatalf("open fixture app: %v", err)
	}
	mustFill(t, app, "#email", "smoke@example.com")
	mustClick(t, app, "#submit")

	waitForCondition(t, 8*time.Second, func() bool {
		return backend.eventCount() >= 2
	}, "recording events to reach backend fixture")

	mustClick(t, popup, "#stopBtn")
	waitForText(t, popup, "#status", "Idle", 8*time.Second)

	if got := backend.sessionCount(); got != 1 {
		t.Fatalf("expected exactly one recording session, got %d", got)
	}
	if got := backend.eventCount(); got < 2 {
		t.Fatalf("expected at least two recorded events, got %d", got)
	}
	if !backend.completed() {
		t.Fatal("expected recording session to be marked completed after stop")
	}
	if !backend.sawAPIKey() {
		t.Fatal("expected extension to authenticate backend calls with X-Api-Key")
	}
}

const recordingFixtureAPIKey = "extension-smoke-key"

type recordingBackendFixture struct {
	*httptest.Server
	mu          sync.Mutex
	sessions    map[string]map[string]any
	events      []map[string]any
	completedID string
	apiKeySeen  bool
}

func newRecordingBackendFixture(t *testing.T) *recordingBackendFixture {
	t.Helper()
	fixture := &recordingBackendFixture{sessions: map[string]map[string]any{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html>
<html>
  <body>
    <form id="login-form">
      <input id="email" data-testid="email" name="email" />
      <button id="submit" type="button">Submit</button>
    </form>
  </body>
</html>`))
	})
	mux.HandleFunc("/api/v1/recording-sessions", fixture.handleSessions)
	mux.HandleFunc("/api/v1/recording-sessions/rec-1", fixture.handleSession)
	mux.HandleFunc("/api/v1/recording-sessions/rec-1/events", fixture.handleEvents)
	fixture.Server = httptest.NewServer(mux)
	return fixture
}

func (f *recordingBackendFixture) handleSessions(w http.ResponseWriter, r *http.Request) {
	f.recordAPIKey(r)
	switch r.Method {
	case http.MethodPost:
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		body["id"] = "rec-1"
		body["status"] = "recording"
		body["created_at"] = time.Now().UTC().Format(time.RFC3339)
		body["updated_at"] = body["created_at"]
		f.mu.Lock()
		f.sessions["rec-1"] = body
		f.mu.Unlock()
		writeJSONFixture(w, http.StatusCreated, body)
	case http.MethodGet:
		f.mu.Lock()
		defer f.mu.Unlock()
		list := make([]map[string]any, 0, len(f.sessions))
		for _, session := range f.sessions {
			list = append(list, session)
		}
		writeJSONFixture(w, http.StatusOK, list)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (f *recordingBackendFixture) handleSession(w http.ResponseWriter, r *http.Request) {
	f.recordAPIKey(r)
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	f.mu.Lock()
	defer f.mu.Unlock()
	if session, ok := f.sessions["rec-1"]; ok {
		for k, v := range body {
			session[k] = v
		}
		if session["status"] == "completed" {
			f.completedID = "rec-1"
		}
		writeJSONFixture(w, http.StatusOK, session)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func (f *recordingBackendFixture) handleEvents(w http.ResponseWriter, r *http.Request) {
	f.recordAPIKey(r)
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	f.mu.Lock()
	f.events = append(f.events, body)
	f.mu.Unlock()
	writeJSONFixture(w, http.StatusCreated, map[string]any{"id": fmt.Sprintf("evt-%d", len(f.events))})
}

func (f *recordingBackendFixture) recordAPIKey(r *http.Request) {
	if r.Header.Get("X-Api-Key") == recordingFixtureAPIKey {
		f.mu.Lock()
		f.apiKeySeen = true
		f.mu.Unlock()
	}
}

func (f *recordingBackendFixture) sessionCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.sessions)
}

func (f *recordingBackendFixture) eventCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.events)
}

func (f *recordingBackendFixture) completed() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.completedID == "rec-1"
}

func (f *recordingBackendFixture) sawAPIKey() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.apiKeySeen
}

func writeJSONFixture(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func waitForExtensionID(t *testing.T, ctx playwright.BrowserContext) string {
	t.Helper()
	var extensionID string
	waitForCondition(t, 10*time.Second, func() bool {
		for _, worker := range ctx.ServiceWorkers() {
			workerURL := worker.URL()
			if strings.HasPrefix(workerURL, "chrome-extension://") && strings.Contains(workerURL, "/background.js") {
				trimmed := strings.TrimPrefix(workerURL, "chrome-extension://")
				extensionID = strings.SplitN(trimmed, "/", 2)[0]
				return extensionID != ""
			}
		}
		return false
	}, "extension service worker")
	return extensionID
}

func mustFill(t *testing.T, page playwright.Page, selector string, value string) {
	t.Helper()
	if err := page.Locator(selector).Fill(value, playwright.LocatorFillOptions{Timeout: playwright.Float(5000)}); err != nil {
		t.Fatalf("fill %s: %v", selector, err)
	}
}

func mustClick(t *testing.T, page playwright.Page, selector string) {
	t.Helper()
	if err := page.Locator(selector).Click(playwright.LocatorClickOptions{Timeout: playwright.Float(5000)}); err != nil {
		t.Fatalf("click %s: %v", selector, err)
	}
}

func waitForText(t *testing.T, page playwright.Page, selector string, text string, timeout time.Duration) {
	t.Helper()
	waitForCondition(t, timeout, func() bool {
		content, err := page.Locator(selector).TextContent(playwright.LocatorTextContentOptions{Timeout: playwright.Float(500)})
		return err == nil && strings.Contains(content, text)
	}, fmt.Sprintf("%s text to contain %q", selector, text))
}

func waitForCondition(t *testing.T, timeout time.Duration, condition func() bool, description string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", description)
}
