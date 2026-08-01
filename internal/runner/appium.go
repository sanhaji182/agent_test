// Package runner provides the native mobile/desktop test execution layer via Appium.
// The AppiumRunner connects to an Appium server (local or remote) using the W3C WebDriver
// protocol over HTTP. It works with Android emulators, iOS simulators, and real devices.
//
// Requirements:
//   - Appium server running (local or remote)
//   - Device/simulator configured in the Appium server
//   - APPIUM_URL env var (default: http://127.0.0.1:4723)
//
// Device profiles are configured via APPIUM_CAPABILITIES or inline in the run request.
package runner

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/google/uuid"
)

// AppiumRunner executes mobile/desktop tests via an Appium server.
type AppiumRunner struct {
	ServerURL    string         // Appium server URL (default: http://127.0.0.1:4723)
	Capabilities map[string]any // Appium desired capabilities
	HTTPClient   *http.Client
	SessionID    string // active session ID after Start()
}

// AppiumSession represents an active Appium session.
type AppiumSession struct {
	ID           string         `json:"sessionId"`
	Capabilities map[string]any `json:"capabilities"`
}

// AppiumElement represents a found element.
type AppiumElement struct {
	ID string `json:"ELEMENT"` // or "element-6066-11e4-a52e-4f735466cecf" in W3C
}

// DeviceProfile defines a pre-configured mobile device.
type DeviceProfile struct {
	Name            string `json:"name"`
	PlatformName    string `json:"platformName"` // "Android" or "iOS"
	DeviceName      string `json:"deviceName"`   // emulator/simulator name
	PlatformVersion string `json:"platformVersion"`
	App             string `json:"app,omitempty"`         // path to .apk or .app
	BrowserName     string `json:"browserName,omitempty"` // for mobile web testing
	AutomationName  string `json:"automationName"`        // "UiAutomator2" or "XCUITest"
}

// Pre-defined device profiles for common testing scenarios.
var DeviceProfiles = map[string]DeviceProfile{
	"pixel-7": {
		Name: "Pixel 7", PlatformName: "Android", DeviceName: "Pixel_7_API_34",
		PlatformVersion: "14.0", AutomationName: "UiAutomator2", BrowserName: "Chrome",
	},
	"pixel-7-app": {
		Name: "Pixel 7 (Native App)", PlatformName: "Android", DeviceName: "Pixel_7_API_34",
		PlatformVersion: "14.0", AutomationName: "UiAutomator2",
	},
	"iphone-15": {
		Name: "iPhone 15", PlatformName: "iOS", DeviceName: "iPhone 15",
		PlatformVersion: "17.0", AutomationName: "XCUITest", BrowserName: "Safari",
	},
	"iphone-15-app": {
		Name: "iPhone 15 (Native App)", PlatformName: "iOS", DeviceName: "iPhone 15",
		PlatformVersion: "17.0", AutomationName: "XCUITest",
	},
	"desktop-chrome": {
		Name: "Desktop Chrome", PlatformName: "Windows", DeviceName: "Windows",
		PlatformVersion: "11", AutomationName: "UiAutomator2", BrowserName: "Chrome",
	},
}

// NewAppiumRunner creates a new Appium runner with default settings.
func NewAppiumRunner() *AppiumRunner {
	serverURL := os.Getenv("APPIUM_URL")
	if serverURL == "" {
		serverURL = "http://127.0.0.1:4723"
	}
	return &AppiumRunner{
		ServerURL:  strings.TrimRight(serverURL, "/"),
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// WithDevice sets capabilities from a device profile.
func (r *AppiumRunner) WithDevice(profile DeviceProfile) *AppiumRunner {
	if r.Capabilities == nil {
		r.Capabilities = make(map[string]any)
	}
	r.Capabilities["platformName"] = profile.PlatformName
	r.Capabilities["appium:deviceName"] = profile.DeviceName
	r.Capabilities["appium:platformVersion"] = profile.PlatformVersion
	r.Capabilities["appium:automationName"] = profile.AutomationName
	if profile.BrowserName != "" {
		r.Capabilities["browserName"] = profile.BrowserName
	}
	if profile.App != "" {
		r.Capabilities["appium:app"] = profile.App
	}
	return r
}

// WithApp sets the app path for native app testing.
func (r *AppiumRunner) WithApp(appPath string) *AppiumRunner {
	if r.Capabilities == nil {
		r.Capabilities = make(map[string]any)
	}
	r.Capabilities["appium:app"] = appPath
	return r
}

// Start creates a new Appium session.
func (r *AppiumRunner) Start(ctx context.Context) error {
	if r.Capabilities == nil {
		return fmt.Errorf("no capabilities configured — use WithDevice() or set capabilities directly")
	}

	body := map[string]any{
		"capabilities": map[string]any{
			"alwaysMatch": r.Capabilities,
		},
	}

	var session AppiumSession
	if err := r.post(ctx, "/session", body, &session); err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	r.SessionID = session.ID
	return nil
}

// Stop closes the active Appium session.
func (r *AppiumRunner) Stop(ctx context.Context) error {
	if r.SessionID == "" {
		return nil
	}
	err := r.delete(ctx, "/session/"+r.SessionID)
	r.SessionID = ""
	return err
}

// Run executes test actions via Appium, following the same interface pattern as PlaywrightRunner.
func (r *AppiumRunner) Run(ctx context.Context, testFiles []agent.TestFile, targetURL string) (*agent.RunResult, error) {
	if err := r.Start(ctx); err != nil {
		return nil, fmt.Errorf("appium start: %w", err)
	}
	defer r.Stop(ctx)

	// Navigate to target URL if provided (for mobile web)
	if targetURL != "" {
		if err := r.Navigate(ctx, targetURL); err != nil {
			return nil, fmt.Errorf("navigate: %w", err)
		}
	}

	result := &agent.RunResult{}

	for _, tf := range testFiles {
		fileResult := r.executeTestFile(ctx, tf)
		result.Total += fileResult.Total
		result.Passed += fileResult.Passed
		result.Failed += fileResult.Failed
		if fileResult.Failed > 0 && len(fileResult.Failures) > 0 {
			result.Failures = append(result.Failures, fileResult.Failures...)
		}
	}

	return result, nil
}

func (r *AppiumRunner) executeTestFile(ctx context.Context, tf agent.TestFile) agent.RunResult {
	var actions []agent.BrowserAction
	if err := json.Unmarshal([]byte(tf.Content), &actions); err != nil {
		return agent.RunResult{Failed: 1, Total: 1, Failures: []agent.Failure{{
			Test: tf.Name, Message: fmt.Sprintf("parse actions: %v", err),
		}}}
	}

	result := agent.RunResult{}
	for _, action := range actions {
		result.Total++
		err := r.executeAction(ctx, action)
		if err != nil {
			result.Failed++
			result.Failures = append(result.Failures, agent.Failure{
				Test:    tf.Name,
				Message: fmt.Sprintf("action %s: %v", action.Action, err),
			})
		} else {
			result.Passed++
		}
	}
	return result
}

func (r *AppiumRunner) executeAction(ctx context.Context, a agent.BrowserAction) error {
	switch a.Action {
	case "goto":
		return r.Navigate(ctx, a.URL)
	case "click":
		return r.Click(ctx, a.Selector, a.Text)
	case "fill":
		return r.SendKeys(ctx, a.Selector, a.Value)
	case "screenshot":
		_, err := r.Screenshot(ctx)
		return err
	case "wait":
		time.Sleep(time.Duration(a.Ms) * time.Millisecond)
		return nil
	case "assert":
		return r.assertElement(ctx, a)
	default:
		return fmt.Errorf("unsupported action for Appium: %s", a.Action)
	}
}

func (r *AppiumRunner) assertElement(ctx context.Context, a agent.BrowserAction) error {
	switch a.Assert {
	case "visible":
		_, err := r.FindElement(ctx, a.Selector, a.Text)
		return err
	case "text_contains":
		el, err := r.FindElement(ctx, a.Selector, a.Text)
		if err != nil {
			return err
		}
		text, err := r.GetText(ctx, el.ID)
		if err != nil {
			return err
		}
		if !strings.Contains(strings.ToLower(text), strings.ToLower(a.Text)) {
			return fmt.Errorf("expected text %q not found in element text %q", a.Text, text)
		}
		return nil
	default:
		return fmt.Errorf("unsupported assert for Appium: %s", a.Assert)
	}
}

// Navigate to a URL (for mobile web testing).
func (r *AppiumRunner) Navigate(ctx context.Context, url string) error {
	body := map[string]string{"url": url}
	return r.post(ctx, "/session/"+r.SessionID+"/url", body, nil)
}

// FindElement finds an element by selector or text content.
func (r *AppiumRunner) FindElement(ctx context.Context, selector, text string) (*AppiumElement, error) {
	var strategy, value string

	switch {
	case selector != "" && strings.HasPrefix(selector, "#"):
		strategy, value = "id", strings.TrimPrefix(selector, "#")
	case selector != "" && strings.HasPrefix(selector, "."):
		strategy, value = "class name", strings.TrimPrefix(selector, ".")
	case selector != "" && strings.HasPrefix(selector, "//"):
		strategy, value = "xpath", selector
	case text != "":
		strategy, value = "accessibility id", text
	default:
		strategy, value = "xpath", "//*"
	}

	body := map[string]string{
		"using": strategy,
		"value": value,
	}

	var el struct {
		Value []struct {
			ElementID string `json:"element-6066-11e4-a52e-4f735466cecf"`
			ELEMENT   string `json:"ELEMENT"`
		} `json:"value"`
	}

	if err := r.post(ctx, "/session/"+r.SessionID+"/elements", body, &el); err != nil {
		return nil, err
	}

	if len(el.Value) == 0 {
		return nil, fmt.Errorf("element not found: %s=%s", strategy, value)
	}

	elem := el.Value[0]
	id := elem.ElementID
	if id == "" {
		id = elem.ELEMENT // fallback to JSONWP key
	}
	return &AppiumElement{ID: id}, nil
}

// Click clicks an element.
func (r *AppiumRunner) Click(ctx context.Context, selector, text string) error {
	el, err := r.FindElement(ctx, selector, text)
	if err != nil {
		return err
	}
	return r.post(ctx, "/session/"+r.SessionID+"/element/"+el.ID+"/click", nil, nil)
}

// SendKeys sends text to an input element.
func (r *AppiumRunner) SendKeys(ctx context.Context, selector, value string) error {
	el, err := r.FindElement(ctx, selector, "")
	if err != nil {
		return err
	}
	body := map[string]any{
		"text":  value,
		"value": []string{value},
	}
	return r.post(ctx, "/session/"+r.SessionID+"/element/"+el.ID+"/value", body, nil)
}

// GetText retrieves the text content of an element.
func (r *AppiumRunner) GetText(ctx context.Context, elementID string) (string, error) {
	var resp struct {
		Value string `json:"value"`
	}
	if err := r.get(ctx, "/session/"+r.SessionID+"/element/"+elementID+"/text", &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// Screenshot takes a screenshot and returns it as base64.
func (r *AppiumRunner) Screenshot(ctx context.Context) (string, error) {
	var resp struct {
		Value string `json:"value"`
	}
	if err := r.get(ctx, "/session/"+r.SessionID+"/screenshot", &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// ScreenshotToFile takes a screenshot and saves it to a file path.
func (r *AppiumRunner) ScreenshotToFile(ctx context.Context, path string) error {
	b64, err := r.Screenshot(ctx)
	if err != nil {
		return err
	}
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return fmt.Errorf("decode screenshot: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// ExecuteScript runs JavaScript in the mobile browser context.
func (r *AppiumRunner) ExecuteScript(ctx context.Context, script string, args ...any) (any, error) {
	body := map[string]any{
		"script": script,
		"args":   args,
	}
	var resp struct {
		Value any `json:"value"`
	}
	if err := r.post(ctx, "/session/"+r.SessionID+"/execute/sync", body, &resp); err != nil {
		return nil, err
	}
	return resp.Value, nil
}

// post sends a POST request to the Appium server.
func (r *AppiumRunner) post(ctx context.Context, path string, body any, result any) error {
	return r.request(ctx, http.MethodPost, path, body, result)
}

// get sends a GET request to the Appium server.
func (r *AppiumRunner) get(ctx context.Context, path string, result any) error {
	return r.request(ctx, http.MethodGet, path, nil, result)
}

// delete sends a DELETE request to the Appium server.
func (r *AppiumRunner) delete(ctx context.Context, path string) error {
	return r.request(ctx, http.MethodDelete, path, nil, nil)
}

func (r *AppiumRunner) request(ctx context.Context, method, path string, body, result any) error {
	url := r.ServerURL + path

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// Check for WebDriver errors
	var errResp struct {
		Value struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		} `json:"value"`
	}
	if json.Unmarshal(respBody, &errResp) == nil && errResp.Value.Error != "" {
		return fmt.Errorf("appium error: %s — %s", errResp.Value.Error, errResp.Value.Message)
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w\nbody: %s", err, string(respBody))
		}
	}

	return nil
}

// Ensure AppiumRunner implements Runner interface.
var _ agent.Runner = (*AppiumRunner)(nil)

func init() {
	// Register device profile names for use in test runs
	_ = uuid.New()
}
