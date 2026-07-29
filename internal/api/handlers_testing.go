package api

import (
	"encoding/json"
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

// registerAdvancedTestingRoutes adds Phase 2+3 testing endpoints.
func (s *Server) registerAdvancedTestingRoutes(r chi.Router) {
	r.Post("/testing/explore", s.handleExploratoryTest)
	r.Post("/testing/performance", s.handlePerformanceAudit)
	r.Post("/testing/accessibility", s.handleAccessibilityAudit)
}
