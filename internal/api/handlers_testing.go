package api

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/playwright-community/playwright-go"
)

// handleExploratoryTest runs AI-driven autonomous exploration of a URL.
// POST /api/v1/testing/explore
func (s *Server) handleExploratoryTest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL      string `json:"url"`
		MaxDepth int    `json:"max_depth"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.URL == "" {
		writeJSONError(w, http.StatusBadRequest, "url is required")
		return
	}
	if req.MaxDepth == 0 {
		req.MaxDepth = 3
	}

	// Launch Playwright and run exploratory test
	pw, err := playwright.Run()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "could not start playwright")
		return
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "could not launch browser")
		return
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "could not create page")
		return
	}
	defer page.Close()

	runner := agent.NewPlaywrightRunner("/tmp/agent_test/videos", nil)
	runner.ScreenshotDir = "/tmp/agent_test/screenshots"

	result, err := runner.RunExploratoryTest(r.Context(), page, req.URL, req.MaxDepth)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// handlePerformanceAudit collects Core Web Vitals for a URL.
// POST /api/v1/testing/performance
func (s *Server) handlePerformanceAudit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.URL == "" {
		writeJSONError(w, http.StatusBadRequest, "url is required")
		return
	}

	pw, err := playwright.Run()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "could not start playwright")
		return
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "could not launch browser")
		return
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "could not create page")
		return
	}
	defer page.Close()

	_, err = page.Goto(req.URL)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "could not navigate to URL: "+err.Error())
		return
	}
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})

	metrics, err := agent.CollectPerformanceMetrics(page)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "could not collect metrics: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, metrics)
}

// handleAccessibilityAudit runs axe-core accessibility audit on a URL.
// POST /api/v1/testing/accessibility
func (s *Server) handleAccessibilityAudit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.URL == "" {
		writeJSONError(w, http.StatusBadRequest, "url is required")
		return
	}

	pw, err := playwright.Run()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "could not start playwright")
		return
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "could not launch browser")
		return
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "could not create page")
		return
	}
	defer page.Close()

	_, err = page.Goto(req.URL)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "could not navigate to URL: "+err.Error())
		return
	}
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})

	violations, err := agent.RunAccessibilityAudit(page)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "accessibility audit failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"url":              req.URL,
		"violations":       violations,
		"total_violations": len(violations),
		"passed":           len(violations) == 0,
	})
}

// handleVisualRegression captures screenshots across browsers/viewports and compares them.
// POST /api/v1/testing/visual-regression
func (s *Server) handleVisualRegression(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL       string   `json:"url"`
		Browsers  []string `json:"browsers"`  // ["chromium", "firefox", "webkit"]
		Viewports []string `json:"viewports"` // ["desktop", "iphone-14"]
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.URL == "" {
		writeJSONError(w, http.StatusBadRequest, "url is required")
		return
	}
	if len(req.Browsers) == 0 {
		req.Browsers = []string{"chromium"}
	}
	if len(req.Viewports) == 0 {
		req.Viewports = []string{"desktop"}
	}

	pw, err := playwright.Run()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "could not start playwright")
		return
	}
	defer pw.Stop()

	type capture struct {
		Browser  string `json:"browser"`
		Viewport string `json:"viewport"`
		Width    int    `json:"width"`
		Height   int    `json:"height"`
		Hash     string `json:"screenshot_hash"`
		Size     int    `json:"screenshot_bytes"`
	}
	var captures []capture

	for _, browserName := range req.Browsers {
		var bt playwright.BrowserType
		switch browserName {
		case "firefox":
			bt = pw.Firefox
		case "webkit":
			bt = pw.WebKit
		default:
			bt = pw.Chromium
			browserName = "chromium"
		}

		browser, err := bt.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(true)})
		if err != nil {
			captures = append(captures, capture{Browser: browserName, Viewport: "error"})
			continue
		}

		for _, vpName := range req.Viewports {
			vp, ok := agent.ViewportPresets[vpName]
			if !ok {
				vp = agent.ViewportPresets["desktop"]
				vpName = "desktop"
			}

			bCtx, err := browser.NewContext(playwright.BrowserNewContextOptions{
				Viewport: &playwright.Size{Width: vp.Width, Height: vp.Height},
				IsMobile: playwright.Bool(vp.IsMobile),
			})
			if err != nil {
				continue
			}

			page, err := bCtx.NewPage()
			if err != nil {
				bCtx.Close()
				continue
			}

			_, navErr := page.Goto(req.URL)
			if navErr == nil {
				page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
					State: playwright.LoadStateNetworkidle,
				})
			}

			ss, ssErr := page.Screenshot(playwright.PageScreenshotOptions{
				FullPage: playwright.Bool(true),
			})

			c := capture{
				Browser:  browserName,
				Viewport: vpName,
				Width:    vp.Width,
				Height:   vp.Height,
			}
			if ssErr == nil && len(ss) > 0 {
				c.Size = len(ss)
				c.Hash = fmt.Sprintf("%x", sha256.Sum256(ss))
			}
			captures = append(captures, c)

			page.Close()
			bCtx.Close()
		}
		browser.Close()
	}

	// Compare: group by viewport, check if hashes differ across browsers
	type comparison struct {
		Viewport    string   `json:"viewport"`
		Identical   bool     `json:"identical"`
		Browsers    []string `json:"browsers"`
		UniqueHashes int     `json:"unique_hashes"`
	}
	byViewport := map[string][]capture{}
	for _, c := range captures {
		byViewport[c.Viewport] = append(byViewport[c.Viewport], c)
	}
	var comparisons []comparison
	for vp, caps := range byViewport {
		hashes := map[string]bool{}
		var browsers []string
		for _, c := range caps {
			if c.Hash != "" {
				hashes[c.Hash] = true
			}
			browsers = append(browsers, c.Browser)
		}
		comparisons = append(comparisons, comparison{
			Viewport:     vp,
			Identical:    len(hashes) <= 1,
			Browsers:     browsers,
			UniqueHashes: len(hashes),
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"url":         req.URL,
		"captures":    captures,
		"comparisons": comparisons,
		"total":       len(captures),
	})
}

// handleFullAudit runs performance, accessibility, and visual checks in one call.
// POST /api/v1/testing/audit
func (s *Server) handleFullAudit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL      string `json:"url"`
		Viewport string `json:"viewport"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.URL == "" {
		writeJSONError(w, http.StatusBadRequest, "url is required")
		return
	}

	pw, err := playwright.Run()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "could not start playwright")
		return
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(true)})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "could not launch browser")
		return
	}
	defer browser.Close()

	// Resolve viewport
	vp := agent.ViewportPresets["desktop"]
	if req.Viewport != "" {
		if p, ok := agent.ViewportPresets[req.Viewport]; ok {
			vp = p
		}
	}

	bCtx, err := browser.NewContext(playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{Width: vp.Width, Height: vp.Height},
		IsMobile: playwright.Bool(vp.IsMobile),
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "could not create context")
		return
	}
	defer bCtx.Close()

	page, err := bCtx.NewPage()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "could not create page")
		return
	}
	defer page.Close()

	_, err = page.Goto(req.URL)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "could not navigate: "+err.Error())
		return
	}
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})

	result := map[string]interface{}{
		"url":      req.URL,
		"viewport": vp.Name,
	}

	// 1. Performance metrics
	if metrics, err := agent.CollectPerformanceMetrics(page); err == nil {
		result["performance"] = metrics
	} else {
		result["performance_error"] = err.Error()
	}

	// 2. Accessibility audit
	if violations, err := agent.RunAccessibilityAudit(page); err == nil {
		result["accessibility"] = map[string]interface{}{
			"violations": violations,
			"count":      len(violations),
			"passed":     len(violations) == 0,
		}
	} else {
		result["accessibility_error"] = err.Error()
	}

	// 3. Screenshot evidence
	if ss, err := page.Screenshot(playwright.PageScreenshotOptions{FullPage: playwright.Bool(true)}); err == nil {
		result["screenshot_hash"] = fmt.Sprintf("%x", sha256.Sum256(ss))
		result["screenshot_bytes"] = len(ss)
	}

	// 4. Page metadata
	if title, err := page.Title(); err == nil {
		result["title"] = title
	}
	result["final_url"] = page.URL()

	writeJSON(w, http.StatusOK, result)
}

// handleExportCode converts a run's generated action JSON into runnable
// Playwright TypeScript test files.
// GET /api/v1/runs/{id}/export-code
func (s *Server) handleExportCode(w http.ResponseWriter, r *http.Request) {
	runID := chi.URLParam(r, "id")
	if !isValidID(runID) {
		writeJSONError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	run, err := s.store.GetRun(r.Context(), runID)
	if err != nil || run == nil {
		writeJSONError(w, http.StatusNotFound, "run not found")
		return
	}
	if len(run.TestFiles) == 0 {
		writeJSONError(w, http.StatusConflict, "run has no generated test files yet")
		return
	}

	opts := agent.ExportOptions{AddWaits: true, Timeout: 5000}
	scripts := agent.ExportAllScripts(run.TestFiles, opts)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"run_id":  runID,
		"count":   len(scripts),
		"scripts": scripts,
	})
}

// registerAdvancedTestingRoutes adds Phase 2+3 testing endpoints.
func (s *Server) registerAdvancedTestingRoutes(r chi.Router) {
	r.Post("/testing/explore", s.handleExploratoryTest)
	r.Post("/testing/performance", s.handlePerformanceAudit)
	r.Post("/testing/accessibility", s.handleAccessibilityAudit)
	r.Post("/testing/visual-regression", s.handleVisualRegression)
	r.Post("/testing/audit", s.handleFullAudit)
	r.Get("/runs/{id}/export-code", s.handleExportCode)
}
