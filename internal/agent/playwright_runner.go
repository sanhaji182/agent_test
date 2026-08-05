package agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/browser"
	"github.com/playwright-community/playwright-go"
)

var playwrightInstallOnce sync.Once
var playwrightInstallErr error

type PlaywrightRunner struct {
	VideoDir      string
	ScreenshotDir string
	llm           LLM
	BrowserType   string // "chromium" (default), "firefox", "webkit"
	Parallel      bool   // execute test files concurrently
	Viewport      *ViewportPreset
	TestData      map[string]string // parameterized test data (template key → value)
	AllowedHosts  []string          // explicit browser egress allowlist for local/private targets
	// ReplayMode menandakan eksekusi rekam-putar deterministik (test case dari
	// recorder). Saat aktif, aksi interaksi menunggu elemen terlihat dulu
	// (best-effort) supaya replay tahan terhadap halaman yang lambat — mirip
	// smart-wait Katalon — tanpa mengubah perilaku run AI biasa.
	ReplayMode bool
}

// ViewportPreset defines browser viewport dimensions for responsive testing.
type ViewportPreset struct {
	Name      string `json:"name"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	IsMobile  bool   `json:"is_mobile"`
	UserAgent string `json:"user_agent,omitempty"`
}

// Built-in viewport presets for mobile/desktop testing.
var ViewportPresets = map[string]ViewportPreset{
	"desktop":        {Name: "desktop", Width: 1280, Height: 720},
	"desktop-hd":     {Name: "desktop-hd", Width: 1920, Height: 1080},
	"ipad":           {Name: "ipad", Width: 768, Height: 1024, IsMobile: true},
	"ipad-landscape": {Name: "ipad-landscape", Width: 1024, Height: 768, IsMobile: true},
	"iphone-14":      {Name: "iphone-14", Width: 390, Height: 844, IsMobile: true},
	"iphone-se":      {Name: "iphone-se", Width: 375, Height: 667, IsMobile: true},
	"pixel-7":        {Name: "pixel-7", Width: 412, Height: 915, IsMobile: true},
	"galaxy-s23":     {Name: "galaxy-s23", Width: 360, Height: 780, IsMobile: true},
}

func NewPlaywrightRunner(videoDir string, llm LLM) *PlaywrightRunner {
	return &PlaywrightRunner{VideoDir: videoDir, llm: llm, BrowserType: "chromium"}
}

// WithBrowser sets the browser engine (chromium, firefox, webkit).
func (r *PlaywrightRunner) WithBrowser(browserType string) *PlaywrightRunner {
	r.BrowserType = browserType
	return r
}

// WithParallel enables concurrent test file execution.
func (r *PlaywrightRunner) WithParallel(parallel bool) *PlaywrightRunner {
	r.Parallel = parallel
	return r
}

// WithViewport sets a viewport preset for responsive testing.
func (r *PlaywrightRunner) WithViewport(preset string) *PlaywrightRunner {
	if vp, ok := ViewportPresets[preset]; ok {
		r.Viewport = &vp
	}
	return r
}

// WithAllowedHosts configures explicit host allowlist entries for browser egress.
// Entries can be exact hosts, comma-separated hosts, URLs, or wildcard suffixes like *.example.test.
func (r *PlaywrightRunner) WithAllowedHosts(hosts ...string) *PlaywrightRunner {
	r.AllowedHosts = append(r.AllowedHosts, hosts...)
	return r
}

type BrowserAction struct {
	Action   string `json:"action"`
	URL      string `json:"url,omitempty"`
	Selector string `json:"selector,omitempty"`
	Value    string `json:"value,omitempty"`
	Key      string `json:"key,omitempty"`
	Assert   string `json:"assert,omitempty"` // "visible", "hidden", "text_contains", "url_contains", "title_contains", "network"
	Text     string `json:"text,omitempty"`   // expected text for assert
	Y        int    `json:"y,omitempty"`
	Ms       int    `json:"ms,omitempty"`
	// Network interception fields
	NetworkURL    string `json:"network_url,omitempty"`    // URL pattern to intercept/assert
	NetworkMethod string `json:"network_method,omitempty"` // GET, POST, etc.
	NetworkStatus int    `json:"network_status,omitempty"` // expected status code
	// Test data parameterization
	Template string `json:"template,omitempty"` // template key for data-driven tests
}

func (r *PlaywrightRunner) Run(ctx context.Context, testFiles []TestFile, projectURL string) (*RunResult, error) {
	// Ensure the screenshot directory exists so failure/step screenshots can be
	// written (Playwright errors with "no such file or directory" otherwise).
	if r.ScreenshotDir != "" {
		_ = os.MkdirAll(r.ScreenshotDir, 0o755)
	}
	// Install Playwright once per process lifetime (AUDIT P-01)
	playwrightInstallOnce.Do(func() {
		playwrightInstallErr = playwright.Install()
	})
	if playwrightInstallErr != nil {
		return nil, fmt.Errorf("playwright not available: %w", playwrightInstallErr)
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %w", err)
	}
	defer pw.Stop()

	// Select browser engine (multi-browser support)
	var browserType playwright.BrowserType
	switch r.BrowserType {
	case "firefox":
		browserType = pw.Firefox
	case "webkit":
		browserType = pw.WebKit
	default:
		browserType = pw.Chromium
	}

	browser, err := browserType.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("could not launch %s browser: %w", r.BrowserType, err)
	}
	defer browser.Close()

	// Watchdog: if the run context is cancelled or times out, force-close the
	// browser so any in-flight (blocking) Playwright operation unblocks and Run
	// returns promptly. Without this, a hung page operation would keep the run
	// stuck in "running" forever, ignoring both cancel requests and the timeout.
	runDone := make(chan struct{})
	defer close(runDone)
	go func() {
		select {
		case <-ctx.Done():
			_ = browser.Close()
		case <-runDone:
		}
	}()

	// Determine viewport dimensions
	vpWidth, vpHeight := 1280, 720
	isMobile := false
	if r.Viewport != nil {
		vpWidth = r.Viewport.Width
		vpHeight = r.Viewport.Height
		isMobile = r.Viewport.IsMobile
	}

	// Setup context with video recording and viewport
	ctxOpts := playwright.BrowserNewContextOptions{
		RecordVideo: &playwright.RecordVideo{
			Dir:  r.VideoDir,
			Size: &playwright.Size{Width: vpWidth, Height: vpHeight},
		},
		Viewport: &playwright.Size{Width: vpWidth, Height: vpHeight},
		IsMobile: playwright.Bool(isMobile),
	}
	if r.Viewport != nil && r.Viewport.UserAgent != "" {
		ctxOpts.UserAgent = playwright.String(r.Viewport.UserAgent)
	}

	bCtx, err := browser.NewContext(ctxOpts)
	if err != nil {
		return nil, fmt.Errorf("could not create context: %w", err)
	}
	defer bCtx.Close()
	if err := InstallContextEgressGuard(bCtx, r.AllowedHosts...); err != nil {
		return nil, fmt.Errorf("could not install browser egress guard: %w", err)
	}

	page, err := bCtx.NewPage()
	if err != nil {
		return nil, fmt.Errorf("could not create page: %w", err)
	}
	defer page.Close()

	time.Sleep(1 * time.Second)

	// Execute test files (parallel or sequential)
	if r.Parallel && len(testFiles) > 1 {
		return r.runParallel(ctx, browser, testFiles, vpWidth, vpHeight, isMobile)
	}
	return r.runSequential(ctx, page, bCtx, testFiles)
}

// runSequential executes test files one by one on a single page.
func (r *PlaywrightRunner) runSequential(ctx context.Context, page playwright.Page, bCtx playwright.BrowserContext, testFiles []TestFile) (*RunResult, error) {
	startTime := time.Now()
	totalActions := 0
	successfulActions := 0
	failedActions := 0
	healedCount := 0
	retriedCount := 0
	var failures []Failure

	for fileIdx := range testFiles {
		tf := testFiles[fileIdx]
		var actions []BrowserAction
		if err := json.Unmarshal([]byte(tf.Content), &actions); err != nil {
			continue
		}
		actionsChanged := false

		for i := 0; i < len(actions); i++ {
			totalActions++
			actionStart := time.Now()
			before := actions[i]
			err := r.executeAction(ctx, page, &actions[i])
			// Jika self-healing mengubah aksi (via pointer), simpan kembali ke slice
			// supaya versi yang sudah diperbaiki ikut tersimpan ke test case.
			if actions[i] != before {
				healedCount++
				actionsChanged = true
			}

			// Simple retry once before declaring failure (reduces flaky results)
			if err != nil {
				time.Sleep(500 * time.Millisecond)
				err = r.executeAction(ctx, page, &actions[i])
				if err == nil {
					retriedCount++
				}
			}

			// Record every failure (with an optional screenshot) so the run result
			// carries actionable detail and the fix loop has real input. Previously
			// failures were only captured when a screenshot succeeded, which left
			// run_result.failures empty (null) despite a non-zero failed count.
			if err != nil {
				failure := Failure{
					Test:       fmt.Sprintf("%s:action_%d", tf.Name, i),
					Message:    err.Error(),
					DurationMs: time.Since(actionStart).Milliseconds(),
				}
				if r.ScreenshotDir != "" {
					ssName := fmt.Sprintf("fail_%s_%d_%d.png", tf.Name, i, time.Now().UnixMilli())
					ssPath := filepath.Join(r.ScreenshotDir, ssName)
					if ss, ssErr := page.Screenshot(playwright.PageScreenshotOptions{Path: playwright.String(ssPath)}); ssErr == nil && len(ss) > 0 {
						failure.Screenshot = "/screenshots/" + ssName
					}
				}
				failures = append(failures, failure)
				failedActions++
			} else {
				successfulActions++
			}

			time.Sleep(500 * time.Millisecond)
		}

		// Tulis kembali aksi yang sudah diperbaiki (healed) ke testFiles supaya
		// run.TestFiles & auto-saved test case membawa versi final — replay
		// deterministik tidak perlu healing lagi.
		if actionsChanged {
			if b, err := json.Marshal(actions); err == nil {
				testFiles[fileIdx].Content = string(b)
			}
		}
	}

	// Close page and context to finalize video
	page.Close()
	bCtx.Close()

	// Find the recorded video path
	video := page.Video()
	var videoPath string
	if video != nil {
		if path, err := video.Path(); err == nil {
			videoPath = "/videos/" + filepath.Base(path)
		}
	}

	return &RunResult{
		Passed:     successfulActions,
		Failed:     failedActions,
		Total:      totalActions,
		Failures:   failures,
		VideoPath:  videoPath,
		DurationMs: time.Since(startTime).Milliseconds(),
		Healed:     healedCount,
		Retried:    retriedCount,
	}, nil
}

// runParallel executes test files concurrently, each in its own browser context.
func (r *PlaywrightRunner) runParallel(ctx context.Context, browser playwright.Browser, testFiles []TestFile, vpWidth, vpHeight int, isMobile bool) (*RunResult, error) {
	type fileResult struct {
		passed  int
		failed  int
		total   int
		failure []Failure
	}

	results := make([]fileResult, len(testFiles))
	var wg sync.WaitGroup

	for idx, tf := range testFiles {
		wg.Add(1)
		go func(i int, tf TestFile) {
			defer wg.Done()

			bCtx, err := browser.NewContext(playwright.BrowserNewContextOptions{
				Viewport: &playwright.Size{Width: vpWidth, Height: vpHeight},
				IsMobile: playwright.Bool(isMobile),
			})
			if err != nil {
				results[i] = fileResult{failed: 1, total: 1}
				return
			}
			defer bCtx.Close()
			if err := InstallContextEgressGuard(bCtx, r.AllowedHosts...); err != nil {
				results[i] = fileResult{failed: 1, total: 1}
				return
			}

			page, err := bCtx.NewPage()
			if err != nil {
				results[i] = fileResult{failed: 1, total: 1}
				return
			}
			defer page.Close()

			var actions []BrowserAction
			if err := json.Unmarshal([]byte(tf.Content), &actions); err != nil {
				return
			}

			res := fileResult{}
			for j := 0; j < len(actions); j++ {
				a := actions[j]
				res.total++
				if err := r.executeAction(ctx, page, &a); err != nil {
					res.failed++
					res.failure = append(res.failure, Failure{
						Test:    fmt.Sprintf("%s:action_%d", tf.Name, j),
						Message: err.Error(),
					})
				} else {
					res.passed++
				}
			}
			results[i] = res
		}(idx, tf)
	}

	wg.Wait()

	// Aggregate results
	total := &RunResult{Failures: []Failure{}}
	for _, res := range results {
		total.Passed += res.passed
		total.Failed += res.failed
		total.Total += res.total
		total.Failures = append(total.Failures, res.failure...)
	}
	return total, nil
}

// executeAction runs a single browser action with self-healing on failure.
func (r *PlaywrightRunner) executeAction(ctx context.Context, page playwright.Page, a *BrowserAction) error {
	// Template expansion: replace {{key}} with test data values
	r.expandTemplate(a)

	// Replay mode (record & playback): tunggu elemen terlihat sebelum interaksi
	// supaya replay tahan halaman lambat — best-effort, gagal = lanjut (error
	// asli dari aksi tetap muncul).
	if r.ReplayMode && a.Selector != "" {
		switch a.Action {
		case "click", "fill", "hover", "select", "press":
			_ = page.Locator(a.Selector).First().WaitFor(playwright.LocatorWaitForOptions{
				State:   playwright.WaitForSelectorStateVisible,
				Timeout: playwright.Float(8000),
			})
		}
	}

	var err error
	switch a.Action {
	case "goto":
		if !IsSafeBrowserURL(a.URL, r.AllowedHosts...) {
			return fmt.Errorf("browser egress blocked: unsafe URL %q", a.URL)
		}
		_, err = page.Goto(a.URL, playwright.PageGotoOptions{Timeout: playwright.Float(60000)})
		if err == nil {
			page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
				State:   playwright.LoadStateNetworkidle,
				Timeout: playwright.Float(15000),
			})
		}
	case "fill":
		err = page.Locator(a.Selector).Fill(a.Value, playwright.LocatorFillOptions{Timeout: playwright.Float(3000)})
	case "click":
		err = page.Locator(a.Selector).First().Click(playwright.LocatorClickOptions{Timeout: playwright.Float(3000)})
	case "scroll":
		_, err = page.Evaluate(fmt.Sprintf("window.scrollBy(0, %d)", a.Y))
	case "wait":
		time.Sleep(time.Duration(a.Ms) * time.Millisecond)
	case "hover":
		err = page.Locator(a.Selector).First().Hover(playwright.LocatorHoverOptions{Timeout: playwright.Float(3000)})
	case "select":
		_, err = page.Locator(a.Selector).First().SelectOption(playwright.SelectOptionValues{Values: &[]string{a.Value}}, playwright.LocatorSelectOptionOptions{Timeout: playwright.Float(3000)})
	case "press":
		if a.Selector != "" {
			err = page.Locator(a.Selector).First().Press(a.Key, playwright.LocatorPressOptions{Timeout: playwright.Float(3000)})
		} else {
			// Tanpa selector: tekan tombol di keyboard (mis. rekam dari extension).
			err = page.Keyboard().Press(a.Key)
		}
	case "screenshot":
		// Capture screenshot for visual testing / evidence
		if r.ScreenshotDir != "" {
			ssName := fmt.Sprintf("step_%d.png", time.Now().UnixMilli())
			ssPath := filepath.Join(r.ScreenshotDir, ssName)
			_, err = page.Screenshot(playwright.PageScreenshotOptions{Path: playwright.String(ssPath)})
		}
	case "network_wait":
		// Wait for a network request matching the URL pattern using ExpectResponse
		timeout := float64(max(a.Ms, 10000))
		resp, waitErr := page.ExpectResponse(func(resp playwright.Response) bool {
			return strings.Contains(resp.URL(), a.NetworkURL)
		}, func() error {
			// No-op trigger: just wait for any pending response
			time.Sleep(100 * time.Millisecond)
			return nil
		}, playwright.PageExpectResponseOptions{Timeout: playwright.Float(timeout)})
		if waitErr != nil {
			err = fmt.Errorf("network_wait: no response matching %q within timeout: %w", a.NetworkURL, waitErr)
		} else if a.NetworkStatus > 0 && resp.Status() != a.NetworkStatus {
			err = fmt.Errorf("network_wait: %q returned status %d, expected %d", a.NetworkURL, resp.Status(), a.NetworkStatus)
		}
	case "assert":
		switch a.Assert {
		case "visible":
			count, _ := page.Locator(a.Selector).Count()
			if count == 0 {
				err = fmt.Errorf("assert failed: %q not found in DOM", a.Selector)
			} else if visible, _ := page.Locator(a.Selector).First().IsVisible(); !visible {
				err = fmt.Errorf("assert failed: %q not visible", a.Selector)
			}
		case "hidden":
			visible, _ := page.Locator(a.Selector).First().IsVisible()
			if visible {
				err = fmt.Errorf("assert failed: %q should be hidden but is visible", a.Selector)
			}
		case "text_contains":
			textContent, _ := page.Locator(a.Selector).First().TextContent()
			if !strings.Contains(textContent, a.Text) {
				err = fmt.Errorf("assert failed: %q text %q does not contain %q", a.Selector, textContent, a.Text)
			}
		case "url_contains":
			if !strings.Contains(page.URL(), a.Text) {
				err = fmt.Errorf("assert failed: URL %q does not contain %q", page.URL(), a.Text)
			}
		case "title_contains":
			title, _ := page.Title()
			if !strings.Contains(title, a.Text) {
				err = fmt.Errorf("assert failed: title %q does not contain %q", title, a.Text)
			}
		case "network":
			// Assert a network request matching pattern occurs
			resp, netErr := page.ExpectResponse(func(resp playwright.Response) bool {
				match := strings.Contains(resp.URL(), a.NetworkURL)
				if a.NetworkMethod != "" {
					match = match && resp.Request().Method() == a.NetworkMethod
				}
				return match
			}, func() error {
				time.Sleep(100 * time.Millisecond)
				return nil
			}, playwright.PageExpectResponseOptions{Timeout: playwright.Float(10000)})
			if netErr != nil {
				err = fmt.Errorf("assert network: no request matching %q %q: %w", a.NetworkMethod, a.NetworkURL, netErr)
			} else if a.NetworkStatus > 0 && resp.Status() != a.NetworkStatus {
				err = fmt.Errorf("assert network: %q returned %d, expected %d", a.NetworkURL, resp.Status(), a.NetworkStatus)
			}
		case "count":
			// Assert element count matches expected (Text field = expected count)
			count, _ := page.Locator(a.Selector).Count()
			expected := 0
			fmt.Sscanf(a.Text, "%d", &expected)
			if count != expected {
				err = fmt.Errorf("assert count: %q found %d elements, expected %d", a.Selector, count, expected)
			}
		case "attribute":
			// Assert element has attribute with expected value
			attr, _ := page.Locator(a.Selector).First().GetAttribute(a.Key)
			if attr != a.Text {
				err = fmt.Errorf("assert attribute: %q[%s] = %q, expected %q", a.Selector, a.Key, attr, a.Text)
			}
		}
	}

	// SELF-HEALING LOOP on failure
	if err != nil && r.llm != nil {
		for attempt := 1; attempt <= 3; attempt++ {
			// Use CDP accessibility tree snapshot for richer, LLM-efficient context
			domSnapshot := getCompactDOMSnapshot(page)

			var imageBase64 string
			if screenshot, errImg := page.Screenshot(playwright.PageScreenshotOptions{
				Type:    playwright.ScreenshotTypeJpeg,
				Quality: playwright.Int(50),
			}); errImg == nil {
				imageBase64 = base64.StdEncoding.EncodeToString(screenshot)
			}

			actionJSON, _ := json.Marshal(a)
			var newActionStr string
			var healErr error
			if imageBase64 != "" {
				newActionStr, healErr = r.llm.HealActionWithVision(ctx, string(actionJSON), fmt.Sprintf("%v", domSnapshot), err.Error(), imageBase64)
			} else {
				newActionStr, healErr = r.llm.HealAction(ctx, string(actionJSON), fmt.Sprintf("%v", domSnapshot), err.Error())
			}

			if healErr != nil {
				break
			}

			var newAction BrowserAction
			if json.Unmarshal([]byte(newActionStr), &newAction) == nil {
				*a = newAction
				// Retry the healed action (without recursion into healing)
				switch a.Action {
				case "goto":
					if !IsSafeBrowserURL(a.URL, r.AllowedHosts...) {
						return fmt.Errorf("browser egress blocked: unsafe URL %q", a.URL)
					}
					_, err = page.Goto(a.URL, playwright.PageGotoOptions{Timeout: playwright.Float(60000)})
				case "fill":
					err = page.Locator(a.Selector).Fill(a.Value, playwright.LocatorFillOptions{Timeout: playwright.Float(3000)})
				case "click":
					err = page.Locator(a.Selector).First().Click(playwright.LocatorClickOptions{Timeout: playwright.Float(3000)})
				case "hover":
					err = page.Locator(a.Selector).First().Hover(playwright.LocatorHoverOptions{Timeout: playwright.Float(3000)})
				case "press":
					err = page.Locator(a.Selector).First().Press(a.Key, playwright.LocatorPressOptions{Timeout: playwright.Float(3000)})
				}
				if err == nil {
					break // Healing successful!
				}
			}
		}
	}

	return err
}

// expandTemplate replaces {{key}} placeholders in action fields with TestData values.
// Enables data-driven testing: same test script, different data sets.
func (r *PlaywrightRunner) expandTemplate(a *BrowserAction) {
	if len(r.TestData) == 0 {
		return
	}
	a.URL = expandString(a.URL, r.TestData)
	a.Selector = expandString(a.Selector, r.TestData)
	a.Value = expandString(a.Value, r.TestData)
	a.Text = expandString(a.Text, r.TestData)
	a.NetworkURL = expandString(a.NetworkURL, r.TestData)
}

// expandString replaces all {{key}} occurrences with values from the data map.
func expandString(s string, data map[string]string) string {
	for k, v := range data {
		s = strings.ReplaceAll(s, "{{"+k+"}}", v)
	}
	return s
}

// getCompactDOMSnapshot extracts a compact accessibility tree snapshot
// from the current page state using the CDP-based browser package.
// Falls back to a basic element list if the browser package is unavailable.
func getCompactDOMSnapshot(page playwright.Page) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try CDP accessibility tree
	if result, err := browser.GetPageSnapshotFromPlaywright(ctx, page); err == nil && result != nil {
		return browser.CDPSnapshotToPrompt(result)
	}

	// Fallback: basic element extraction
	raw, err := page.Evaluate(`() => {
		let result = "";
		document.querySelectorAll("button, input, a, select, [role='button'], h1, h2, h3, h4, h5, h6").forEach(el => {
			if (el.offsetParent !== null) {
				result += el.tagName + (el.id ? "#"+el.id : "") + " '" + (el.innerText || el.value || "").trim() + "'\n";
			}
		});
		return result;
	}`)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%v", raw)
}
