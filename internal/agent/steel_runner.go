package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/browser"
	"github.com/go-go-golems/gotest-agent/internal/steel"
	"github.com/playwright-community/playwright-go"
)

// steelDriverOnce memastikan driver Playwright di-install sekali saja untuk
// seluruh SteelRunner (driver saja, tanpa browser — Steel sediakan browser remote).
var steelDriverOnce sync.Once
var steelDriverErr error

// SteelRunner executes tests using Steel Browser cloud sessions (ADR-002).
// It connects Playwright to a remote browser via CDP URL provided by Steel.
type SteelRunner struct {
	client        *steel.Client
	llm           LLM
	ScreenshotDir string
	TestData      map[string]string
	AllowedHosts  []string
	Parallel      bool
}

// NewSteelRunner creates a runner that uses Steel Browser cloud for execution.
func NewSteelRunner(client *steel.Client, llm LLM) *SteelRunner {
	return &SteelRunner{client: client, llm: llm}
}

// WithAllowedHosts configures explicit browser egress allowlist entries.
func (r *SteelRunner) WithAllowedHosts(hosts ...string) *SteelRunner {
	r.AllowedHosts = append(r.AllowedHosts, hosts...)
	return r
}

// WithParallel enables concurrent test file execution.
func (r *SteelRunner) WithParallel(parallel bool) *SteelRunner {
	r.Parallel = parallel
	return r
}

// Run executes test files using a Steel cloud browser session.
func (r *SteelRunner) Run(ctx context.Context, testFiles []TestFile, projectURL string) (*RunResult, error) {
	// Ensure the screenshot directory exists so failure screenshots can be written.
	if r.ScreenshotDir != "" {
		_ = os.MkdirAll(r.ScreenshotDir, 0o755)
	}
	// Install Playwright driver saja (skip browser — Steel sediakan browser remote).
	steelDriverOnce.Do(func() {
		steelDriverErr = playwright.Install(&playwright.RunOptions{SkipInstallBrowsers: true})
	})
	if steelDriverErr != nil {
		return nil, fmt.Errorf("steel: install driver: %w", steelDriverErr)
	}

	// Create a Steel browser session
	session, err := r.client.CreateSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("steel: create session: %w", err)
	}
	defer r.client.DestroySession(ctx, session.ID)

	// Install driver Playwright saja (tanpa browser) sebelum memulai — Steel
	// menyediakan browser remote via CDP, jadi browser lokal tidak diperlukan.
	steelDriverOnce.Do(func() {
		steelDriverErr = playwright.Install(&playwright.RunOptions{SkipInstallBrowsers: true})
	})
	if steelDriverErr != nil {
		return nil, fmt.Errorf("steel: install driver: %w", steelDriverErr)
	}

	// Connect Playwright to the remote browser via CDP
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("steel: could not start playwright: %w", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.ConnectOverCDP(session.CDPURL)
	if err != nil {
		return nil, fmt.Errorf("steel: connect CDP %s: %w", session.CDPURL, err)
	}
	defer browser.Close()

	// Execute test files concurrently when enabled (each in its own context).
	if r.Parallel && len(testFiles) > 1 {
		return r.runParallel(ctx, browser, testFiles)
	}

	// Use existing context or create new one
	contexts := browser.Contexts()
	var bCtx playwright.BrowserContext
	if len(contexts) > 0 {
		bCtx = contexts[0]
	} else {
		bCtx, err = browser.NewContext()
		if err != nil {
			return nil, fmt.Errorf("steel: create context: %w", err)
		}
	}

	if err := InstallContextEgressGuard(bCtx, r.AllowedHosts...); err != nil {
		return nil, fmt.Errorf("steel: install browser egress guard: %w", err)
	}

	page, err := bCtx.NewPage()
	if err != nil {
		return nil, fmt.Errorf("steel: create page: %w", err)
	}
	defer page.Close()

	// Execute actions using shared logic with PlaywrightRunner
	localRunner := &PlaywrightRunner{
		ScreenshotDir: r.ScreenshotDir,
		llm:           r.llm,
		TestData:      r.TestData,
		AllowedHosts:  r.AllowedHosts,
	}

	totalActions := 0
	successfulActions := 0
	failedActions := 0
	var failures []Failure
	var stepShots []string

	for fileIdx := range testFiles {
		tf := testFiles[fileIdx]
		var actions []BrowserAction
		if err := json.Unmarshal([]byte(tf.Content), &actions); err != nil {
			continue
		}
		actionsChanged := false

		for i := 0; i < len(actions); i++ {
			totalActions++
			before := actions[i]
			err := localRunner.executeAction(ctx, page, &actions[i])
			// Simpan aksi hasil self-healing kembali ke slice (lihat playwright_runner).
			if actions[i] != before {
				actionsChanged = true
			}

			// Capture a screenshot after every step (success or failure) so runs
			// always produce visual evidence — not only when something fails.
			if r.ScreenshotDir != "" {
				stepName := fmt.Sprintf("step_%s_%d_%d", tf.Name, i, time.Now().UnixMilli())
				ssPath := r.ScreenshotDir + "/" + stepName + ".png"
				if ss, ssErr := page.Screenshot(playwright.PageScreenshotOptions{Path: playwright.String(ssPath)}); ssErr == nil && len(ss) > 0 {
					stepShots = append(stepShots, "/screenshots/"+stepName+".png")
				}
			}

			if err != nil && r.ScreenshotDir != "" {
				ssName := fmt.Sprintf("steel_fail_%s_%d_%d.png", tf.Name, i, time.Now().UnixMilli())
				ssPath := r.ScreenshotDir + "/" + ssName
				if ss, ssErr := page.Screenshot(playwright.PageScreenshotOptions{Path: playwright.String(ssPath)}); ssErr == nil && len(ss) > 0 {
					failures = append(failures, Failure{
						Test:       fmt.Sprintf("%s:action_%d", tf.Name, i),
						Message:    err.Error(),
						Screenshot: "/screenshots/" + ssName,
					})
				}
			}

			if err != nil {
				failedActions++
			} else {
				successfulActions++
			}

			time.Sleep(500 * time.Millisecond)
		}

		// Tulis kembali aksi hasil healing ke testFiles (untuk auto-save test case).
		if actionsChanged {
			if b, err := json.Marshal(actions); err == nil {
				testFiles[fileIdx].Content = string(b)
			}
		}
	}

	return &RunResult{
		Passed:      successfulActions,
		Failed:      failedActions,
		Total:       totalActions,
		Failures:    failures,
		Screenshots: stepShots,
	}, nil
}

// runParallel executes test files concurrently, each in its own browser context
// within the shared Steel session. This cuts wall-clock time roughly by the
// number of test files (bounded by the slowest file).
func (r *SteelRunner) runParallel(ctx context.Context, browser playwright.Browser, testFiles []TestFile) (*RunResult, error) {
	type fileResult struct {
		passed   int
		failed   int
		total    int
		failures []Failure
	}
	results := make([]fileResult, len(testFiles))
	var wg sync.WaitGroup

	localRunner := &PlaywrightRunner{
		ScreenshotDir: r.ScreenshotDir,
		llm:           r.llm,
		TestData:      r.TestData,
		AllowedHosts:  r.AllowedHosts,
	}

	for idx, tf := range testFiles {
		wg.Add(1)
		go func(i int, tf TestFile) {
			defer wg.Done()

			bCtx, err := browser.NewContext()
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
				if err := localRunner.executeAction(ctx, page, &a); err != nil {
					res.failed++
					res.failures = append(res.failures, Failure{
						Test:    fmt.Sprintf("%s:action_%d", tf.Name, j),
						Message: err.Error(),
					})
				} else {
					res.passed++
				}
				time.Sleep(500 * time.Millisecond)
			}
			results[i] = res
		}(idx, tf)
	}
	wg.Wait()

	agg := &RunResult{}
	for _, res := range results {
		agg.Passed += res.passed
		agg.Failed += res.failed
		agg.Total += res.total
		agg.Failures = append(agg.Failures, res.failures...)
	}
	return agg, nil
}

// --- Phase 3: Performance Metrics Collection ---

// PerformanceMetrics captures Core Web Vitals and timing data via CDP.
type PerformanceMetrics struct {
	FCP  float64 `json:"fcp_ms"`  // First Contentful Paint
	LCP  float64 `json:"lcp_ms"`  // Largest Contentful Paint
	CLS  float64 `json:"cls"`     // Cumulative Layout Shift
	TTFB float64 `json:"ttfb_ms"` // Time to First Byte
	DOM  float64 `json:"dom_ms"`  // DOM Content Loaded
	Load float64 `json:"load_ms"` // Full page load
}

// CollectPerformanceMetrics gathers Core Web Vitals from the current page.
func CollectPerformanceMetrics(page playwright.Page) (*PerformanceMetrics, error) {
	// Get navigation timing
	result, err := page.Evaluate(`() => {
		const nav = performance.getEntriesByType('navigation')[0];
		const paint = performance.getEntriesByType('paint');
		const fcp = paint.find(e => e.name === 'first-contentful-paint');
		return {
			ttfb: nav ? nav.responseStart - nav.requestStart : 0,
			dom: nav ? nav.domContentLoadedEventEnd - nav.startTime : 0,
			load: nav ? nav.loadEventEnd - nav.startTime : 0,
			fcp: fcp ? fcp.startTime : 0
		};
	}`)
	if err != nil {
		return nil, fmt.Errorf("collect metrics: %w", err)
	}

	metrics := &PerformanceMetrics{}
	if m, ok := result.(map[string]interface{}); ok {
		metrics.TTFB = toFloat(m["ttfb"])
		metrics.DOM = toFloat(m["dom"])
		metrics.Load = toFloat(m["load"])
		metrics.FCP = toFloat(m["fcp"])
	}

	// LCP via PerformanceObserver (already observed entries)
	lcpResult, _ := page.Evaluate(`() => {
		return new Promise(resolve => {
			new PerformanceObserver(list => {
				const entries = list.getEntries();
				resolve(entries.length > 0 ? entries[entries.length-1].startTime : 0);
			}).observe({type: 'largest-contentful-paint', buffered: true});
			setTimeout(() => resolve(0), 3000);
		});
	}`)
	metrics.LCP = toFloat(lcpResult)

	return metrics, nil
}

// --- Phase 3: Accessibility Testing ---

// AccessibilityViolation represents a single a11y issue found on the page.
type AccessibilityViolation struct {
	Rule   string `json:"rule"`
	Impact string `json:"impact"` // minor, moderate, serious, critical
	Help   string `json:"help"`
	Target string `json:"target"` // CSS selector of violating element
}

// RunAccessibilityAudit performs a basic accessibility check using ARIA queries.
// For full axe-core integration, inject the axe-core script via page.Evaluate.
func RunAccessibilityAudit(page playwright.Page) ([]AccessibilityViolation, error) {
	// Inject axe-core from CDN and run audit
	result, err := page.Evaluate(`() => {
		return new Promise((resolve, reject) => {
			const script = document.createElement('script');
			script.src = 'https://cdnjs.cloudflare.com/ajax/libs/axe-core/4.9.1/axe.min.js';
			script.onload = () => {
				axe.run(document, {
					runOnly: ['wcag2a', 'wcag2aa', 'best-practice']
				}).then(results => {
					resolve(results.violations.map(v => ({
						rule: v.id,
						impact: v.impact,
						help: v.help,
						target: v.nodes.length > 0 ? v.nodes[0].target.join(' ') : ''
					})));
				}).catch(reject);
			};
			script.onerror = () => reject(new Error('failed to load axe-core'));
			document.head.appendChild(script);
		});
	}`)
	if err != nil {
		return nil, fmt.Errorf("accessibility audit: %w", err)
	}

	var violations []AccessibilityViolation
	if raw, ok := result.([]interface{}); ok {
		for _, item := range raw {
			if m, ok := item.(map[string]interface{}); ok {
				violations = append(violations, AccessibilityViolation{
					Rule:   fmt.Sprintf("%v", m["rule"]),
					Impact: fmt.Sprintf("%v", m["impact"]),
					Help:   fmt.Sprintf("%v", m["help"]),
					Target: fmt.Sprintf("%v", m["target"]),
				})
			}
		}
	}
	return violations, nil
}

func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	}
	return 0
}

// --- Phase 3: AI-Driven Exploratory Testing ---

// ExploratoryResult captures findings from autonomous page exploration.
type ExploratoryResult struct {
	PagesVisited  int      `json:"pages_visited"`
	ActionsTried  int      `json:"actions_tried"`
	ErrorsFound   []string `json:"errors_found"`
	ConsoleErrors []string `json:"console_errors"`
	BrokenLinks   []string `json:"broken_links"`
	Screenshots   []string `json:"screenshots"`
}

// RunExploratoryTest performs AI-driven autonomous exploration of a page.
// It discovers interactive elements, tries them, and reports issues.
func (r *PlaywrightRunner) RunExploratoryTest(ctx context.Context, page playwright.Page, startURL string, maxDepth int) (*ExploratoryResult, error) {
	result := &ExploratoryResult{}

	// Collect console errors via injected listener
	page.Evaluate(`() => {
		window.__consoleErrors = [];
		const origError = console.error;
		console.error = (...args) => { window.__consoleErrors.push(args.join(' ')); origError(...args); };
	}`)

	// Navigate to start
	if err := InstallPageEgressGuard(page, r.AllowedHosts...); err != nil {
		return nil, fmt.Errorf("exploratory: install browser egress guard: %w", err)
	}
	if !IsSafeBrowserURL(startURL, r.AllowedHosts...) {
		return nil, fmt.Errorf("exploratory: unsafe URL %q", startURL)
	}
	_, err := page.Goto(startURL)
	if err != nil {
		return nil, fmt.Errorf("exploratory: goto %s: %w", startURL, err)
	}
	result.PagesVisited++

	// Discover and interact with elements
	for depth := 0; depth < maxDepth; depth++ {
		// Use CDP accessibility tree snapshot for richer element discovery
		snapshot, sErr := browser.GetPageSnapshotFromPlaywright(ctx, page)
		if sErr != nil {
			continue
		}
		interactive := browser.FindElementsByRoleCDP(snapshot, "link")
		interactive = append(interactive, browser.FindElementsByRoleCDP(snapshot, "button")...)
		interactive = append(interactive, browser.FindElementsByRoleCDP(snapshot, "textbox")...)

		limit := len(interactive)
		if limit > 10 {
			limit = 10
		}
		for i := 0; i < limit; i++ {
			el := interactive[i]
			// Skip text inputs for clicking — they need fill, not click
			if el.Tag == "input" && el.Role == "textbox" {
				continue
			}
			result.ActionsTried++

			// Try clicking using the element's text as locator fallback
			var clickErr error
			if el.Text != "" {
				clickErr = page.GetByText(el.Text).First().Click(playwright.LocatorClickOptions{Timeout: playwright.Float(3000)})
			} else {
				clickErr = page.Locator("a[href]").First().Click(playwright.LocatorClickOptions{Timeout: playwright.Float(3000)})
			}
			if clickErr != nil {
				result.ErrorsFound = append(result.ErrorsFound, fmt.Sprintf("click %s [@%s]: %v", el.Role, el.Ref, clickErr))
			}

			// Check for JS errors after action
			time.Sleep(500 * time.Millisecond)

			// Screenshot evidence
			if r.ScreenshotDir != "" {
				ssName := fmt.Sprintf("explore_%d_%d.png", depth, result.ActionsTried)
				ssPath := r.ScreenshotDir + "/" + ssName
				if _, ssErr := page.Screenshot(playwright.PageScreenshotOptions{Path: playwright.String(ssPath)}); ssErr == nil {
					result.Screenshots = append(result.Screenshots, "/screenshots/"+ssName)
				}
			}

			// Navigate back if we left the page
			currentURL := page.URL()
			if !strings.HasPrefix(currentURL, startURL) {
				result.PagesVisited++
				page.GoBack()
				time.Sleep(500 * time.Millisecond)
			}
		}
	}

	return result, nil
}
