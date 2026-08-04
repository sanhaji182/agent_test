package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/ai"
	"github.com/go-go-golems/gotest-agent/internal/compare"
	"github.com/go-go-golems/gotest-agent/internal/intelligence"
	"github.com/go-go-golems/gotest-agent/internal/junit"
	"github.com/go-go-golems/gotest-agent/internal/notify"
	"github.com/go-go-golems/gotest-agent/internal/release"
	"github.com/go-go-golems/gotest-agent/internal/schedule"
	"github.com/go-go-golems/gotest-agent/internal/workflow"
	"github.com/google/uuid"
)

// --- Demo seed ---

func (s *Server) handleDemoSeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	now := time.Now()

	// Create sample runs with realistic data
	runs := []struct {
		req            string
		state          agent.State
		passed, failed int
	}{
		{"test login and signup flows", agent.StateDone, 5, 0},
		{"test checkout and payment", agent.StateFailed, 3, 1}, // Modified to match failed coupon scenario
		{"test user profile and settings", agent.StateDone, 4, 0},
		{"regression: homepage and navigation", agent.StateDone, 8, 0},
		{"test API endpoints", agent.StateDone, 6, 1},
	}

	var ids []string
	for i, r := range runs {
		run := &agent.TestRun{
			ID: uuid.New().String(), ProjectPath: "https://demostore.com",
			Requirements: r.req, Mode: "standard", State: r.state,
			CreatedAt: now.Add(time.Duration(-i) * time.Hour), UpdatedAt: now,
		}

		if r.state == agent.StateDone || r.state == agent.StateFailed {
			fin := now.Add(time.Duration(-i)*time.Hour + 45*time.Second)
			run.FinishedAt = &fin
			run.RunResult = &agent.RunResult{Passed: r.passed, Failed: r.failed, Total: r.passed + r.failed}

			// Inject rich data for the checkout scenario to match walkthrough
			if r.req == "test checkout and payment" {
				run.Requirements = "Lakukan proses checkout sebagai Guest:\n1. Cari produk \"Wireless Mouse\".\n2. Tambahkan ke keranjang.\n3. Masuk ke halaman checkout.\n4. Masukkan kupon \"PROMO50\".\n5. Verifikasi bahwa Total Harga terpotong sesuai diskon kupon.\n6. Selesaikan pemesanan."
				run.RunResult.Failures = []agent.Failure{{
					Test:       "Verifikasi Total Harga",
					Message:    "Expected price $25, but found $50. Coupon PROMO50 failed to apply.",
					Screenshot: "https://images.unsplash.com/photo-1555421689-491a97ff2040?w=800&q=80", // dummy screenshot
				}}
				run.VideoURL = "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4" // Dummy public video for replay simulation
				run.VideoStatus = "completed"
				run.VideoDuration = 15.0
				run.VideoFailureMarkerAt = 12.5

				run.TestPlan = &agent.TestPlan{
					Summary: "Test Guest Checkout Flow with Coupon",
					Scenarios: []agent.Scenario{{
						Name: "Guest Checkout", Priority: "high",
						Steps: []string{
							"Navigate to https://demostore.com",
							"Search for \"Wireless Mouse\"",
							"Add item to cart",
							"Proceed to checkout",
							"Apply coupon \"PROMO50\"",
							"Verify total price discount",
						},
					}},
				}
			} else {
				if r.failed > 0 {
					run.RunResult.Failures = []agent.Failure{{Test: "API check", Message: "Timeout on /api/users"}}
				}
				run.TestPlan = &agent.TestPlan{Summary: r.req, Scenarios: []agent.Scenario{{Name: r.req, Priority: "high", Steps: []string{"Navigate to page", "Fill form", "Submit", "Verify result"}}}}
			}
		}
		s.store.CreateRun(ctx, run)
		ids = append(ids, run.ID)

		// Emit rich events for the checkout run to simulate execution timeline
		if r.req == "test checkout and payment" {
			s.events.Emit(run.ID, "run_started", "idle", "Run started: E-Commerce Coupon Validation", nil)
			s.events.Emit(run.ID, "analysis_completed", "analyzing", "Analyzed demostore.com DOM structure", nil)
			s.events.Emit(run.ID, "plan_generated", "plan_generated", "Generated Test Plan for checkout flow", nil)
			s.events.Emit(run.ID, "step_started", "running", "Navigating to checkout page", map[string]string{"step": "Proceed to checkout", "timestamp_ms": fmt.Sprintf("%d", (now.UnixMilli() - 10000))})
			s.events.Emit(run.ID, "step_started", "running", "Applying coupon PROMO50", map[string]string{"step": "Apply coupon \"PROMO50\"", "timestamp_ms": fmt.Sprintf("%d", (now.UnixMilli() - 5000))})
			s.events.Emit(run.ID, "step_started", "running", "Verifying total price", map[string]string{"step": "Verify total price discount", "timestamp_ms": fmt.Sprintf("%d", (now.UnixMilli() - 2500))})
			s.events.Emit(run.ID, "assertion_failed", "running", "Expected price $25, but found $50. Coupon PROMO50 failed to apply.", map[string]string{"expected": "$25", "actual": "$50"})
			s.events.Emit(run.ID, "run_failed", "failed", "Run failed during verification", nil)
		} else {
			s.events.Emit(run.ID, "run_started", "idle", "Run started", nil)
			s.events.Emit(run.ID, "analysis_completed", "analyzing", "Analysis complete", nil)
			s.events.Emit(run.ID, "plan_generated", "plan_generated", "Generated test plan", nil)
			if r.state == agent.StateDone {
				s.events.Emit(run.ID, "run_completed", "done", "Run completed", nil)
			} else if r.state == agent.StateFailed {
				s.events.Emit(run.ID, "run_failed", "failed", "Run failed", nil)
			}
		}
	}

	// Create a sample schedule
	s.schedules.Create(&schedule.Schedule{
		Name: "Nightly Regression", ProjectPath: "/demo/app",
		Requirements: "full regression", Frequency: "daily",
		Environment: "staging", BaseURL: "http://staging.demo.com",
		Enabled: true, NextRunAt: now.Add(12 * time.Hour),
	})

	// Create sample releases (history: completed & cancelled + active)
	s.releases.Create(&release.Release{
		Name: "v2.1.0", Version: "2.1.0", ProjectID: "demo",
		Status: "active", RunIDs: ids[:3],
	})
	s.releases.Create(&release.Release{
		Name: "v2.0.3", Version: "2.0.3", ProjectID: "demo",
		Status: "completed", RunIDs: ids[3:],
	})
	s.releases.Create(&release.Release{
		Name: "v2.0.2", Version: "2.0.2", ProjectID: "demo",
		Status: "cancelled", RunIDs: ids[3:4],
	})

	// Create sample test suites so the Suites page isn't empty
	s.suites.Create(&workflow.Suite{
		Name: "Smoke: Login & Registration", ProjectID: "demo",
		Environment: "staging", Tags: []string{"smoke", "critical", "auth"},
		Pinned: true, RunIDs: ids[:2],
	})
	s.suites.Create(&workflow.Suite{
		Name: "E2E Checkout Flow", ProjectID: "demo",
		Environment: "staging", Tags: []string{"e2e", "checkout", "payment"},
		Pinned: true, RunIDs: ids[1:3],
	})
	s.suites.Create(&workflow.Suite{
		Name: "API Regression", ProjectID: "demo",
		Environment: "production", Tags: []string{"api", "regression"},
		Pinned: false, RunIDs: ids[3:],
	})

	// Create sample reviews so the Reviews page isn't empty
	pendingRev := s.reviews.Create(&workflow.Review{RunID: ids[2], Type: "test_plan"})
	s.reviews.Create(&workflow.Review{RunID: ids[0], Type: "test_scripts"})
	approvedRev := s.reviews.Create(&workflow.Review{RunID: ids[3], Type: "test_plan"})
	s.reviews.Approve(approvedRev.ID, "QA Lead", "Plan looks solid, approved for execution.")
	rejectedRev := s.reviews.Create(&workflow.Review{RunID: ids[4], Type: "fix_suggestion"})
	s.reviews.Reject(rejectedRev.ID, "QA Lead", "Fix suggestion misses the coupon discount case.")
	_ = pendingRev

	// Create sample notifications so the Alerts page isn't empty
	s.notifs.Add(notify.Notification{RunID: ids[1], Type: "failure", Message: "Coupon PROMO50 failed to apply: expected $25, found $50."})
	s.notifs.Add(notify.Notification{RunID: ids[4], Type: "flake", Message: "Flaky test detected: API check timed out on /api/users in 2 of last 5 runs."})
	s.notifs.Add(notify.Notification{RunID: ids[3], Type: "degradation", Message: "Homepage load time increased from 1.2s to 3.8s (p95)."})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Demo data seeded",
		"runs":    len(ids),
	})
}

// --- Export handlers ---

func (s *Server) handleExportRun(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	run, err := s.store.GetRun(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=run-"+id[:8]+".json")
	json.NewEncoder(w).Encode(redactCredentials(run))
}

func (s *Server) handleExportCompare(w http.ResponseWriter, r *http.Request) {
	idA := chi.URLParam(r, "id")
	idB := chi.URLParam(r, "otherId")
	runA, err := s.store.GetRun(r.Context(), idA)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "run A not found")
		return
	}
	runB, err := s.store.GetRun(r.Context(), idB)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "run B not found")
		return
	}
	result := compare.Compare(redactCredentials(runA), redactCredentials(runB))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=compare-"+idA[:8]+"-vs-"+idB[:8]+".json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleExportRisk(w http.ResponseWriter, r *http.Request) {
	runs := s.getAllRuns(r)
	scheds := s.schedules.List()
	risks := intelligence.ComputeRisk(runs, scheds)
	if risks == nil {
		risks = []intelligence.RiskItem{}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=risk-report.json")
	json.NewEncoder(w).Encode(risks)
}

func (s *Server) handleExportConfidence(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	rel, ok := s.releases.Get(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	var runs []*agent.TestRun
	for _, rid := range rel.RunIDs {
		if run, err := s.store.GetRun(r.Context(), rid); err == nil {
			runs = append(runs, run)
		}
	}
	allRuns := s.getAllRuns(r)
	risks := intelligence.ComputeRisk(allRuns, nil)
	conf := intelligence.ComputeConfidence(runs, risks)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=confidence-"+id[:8]+".json")
	json.NewEncoder(w).Encode(conf)
}

// ─── Settings API ───────────────────────────────────────────────────────

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	if s.settings == nil {
		json.NewEncoder(w).Encode(map[string]string{})
		return
	}
	settings, err := s.settings.GetAll(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	// Mask the API key for security
	if apiKey, ok := settings["llm_api_key"]; ok && len(apiKey) > 8 {
		settings["llm_api_key"] = apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
	}
	if fbKey, ok := settings["llm_fallback_api_key"]; ok && len(fbKey) > 8 {
		settings["llm_fallback_api_key"] = fbKey[:4] + "..." + fbKey[len(fbKey)-4:]
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

func (s *Server) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	if s.settings == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "settings not available")
		return
	}
	var payload map[string]string
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	// Whitelist allowed keys
	allowed := map[string]bool{
		"llm_provider": true, "llm_model": true, "llm_api_key": true,
		"llm_base_url":    true,
		"llm_temperature": true, "llm_max_tokens": true,
		"llm_fallback_provider": true, "llm_fallback_model": true,
		"llm_fallback_api_key": true, "llm_fallback_base_url": true,
		"browser_headless": true, "browser_timeout": true,
		"max_fix_attempts": true,
	}
	filtered := make(map[string]string)
	for k, v := range payload {
		if allowed[k] {
			filtered[k] = v
		}
	}
	if err := s.settings.SetMany(r.Context(), filtered); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to save settings")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleListAIProviders(w http.ResponseWriter, r *http.Request) {
	providers := []map[string]interface{}{
		{
			"id":   "anthropic",
			"name": "Anthropic",
			"models": []string{
				"claude-opus-4.8",
				"claude-sonnet-4.6",
				"claude-haiku-4.5",
			},
		},
		{
			"id":   "openai",
			"name": "OpenAI",
			"models": []string{
				"gpt-5.5",
				"gpt-5.5-pro",
				"gpt-5.4",
				"gpt-5.4-mini",
				"gpt-5.4-nano",
			},
		},
		{
			"id":   "google",
			"name": "Google Gemini",
			"models": []string{
				"gemini-3.5-pro",
				"gemini-3.5-flash",
				"gemini-3.1-pro",
				"gemini-3.1-flash-lite",
			},
		},
		{
			"id":   "deepseek",
			"name": "DeepSeek",
			"models": []string{
				"deepseek-v4-pro",
				"deepseek-v4-flash",
				"deepseek-r1",
			},
		},
		{
			"id":   "local",
			"name": "Local (Ollama / vLLM)",
			"models": []string{
				"llama-4-maverick",
				"llama-4-scout",
				"qwen-3.7-max",
				"deepseek-r1",
			},
		},
		{
			"id":     "custom",
			"name":   "Custom (OpenAI-Compatible API)",
			"models": []string{},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(providers)
}

func (s *Server) handleTestAIProvider(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
		APIKey   string `json:"api_key"`
		BaseURL  string `json:"base_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Fall back to stored settings for any field not provided. The UI masks the
	// saved API key, so it sends an empty key to mean "use the saved one".
	if s.settings != nil {
		ctx := r.Context()
		if payload.APIKey == "" {
			if v, err := s.settings.Get(ctx, "llm_api_key"); err == nil {
				payload.APIKey = v
			}
		}
		if payload.Provider == "" {
			if v, err := s.settings.Get(ctx, "llm_provider"); err == nil {
				payload.Provider = v
			}
		}
		if payload.Model == "" {
			if v, err := s.settings.Get(ctx, "llm_model"); err == nil {
				payload.Model = v
			}
		}
		if payload.BaseURL == "" {
			if v, err := s.settings.Get(ctx, "llm_base_url"); err == nil {
				payload.BaseURL = v
			}
		}
	}

	// Bound the test so a slow/hanging provider can't pin the request (and the
	// UI's Test button) indefinitely.
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()

	// Create a temporary client just for the test
	client := agent.NewLLM(payload.Provider, payload.Model, payload.APIKey, payload.BaseURL)
	if client == nil {
		writeJSONError(w, http.StatusBadRequest, "provider not supported for test")
		return
	}

	// Just a simple health check or analyze call to see if it responds
	_, err := client.AnalyzeCodebase(ctx, "ping")
	if err != nil {
		errMsg := err.Error()
		if errors.Is(err, context.DeadlineExceeded) {
			errMsg = "connection timed out — provider did not respond within 25 seconds"
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// handleListProviderModels mengambil daftar model yang tersedia dari endpoint
// /models provider (format OpenAI-compatible). Memungkinkan dashboard mengisi
// dropdown model secara dinamis untuk provider custom.
func (s *Server) handleListProviderModels(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Provider string `json:"provider"`
		APIKey   string `json:"api_key"`
		BaseURL  string `json:"base_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	provider := strings.ToLower(strings.TrimSpace(payload.Provider))
	baseURL := strings.TrimRight(strings.TrimSpace(payload.BaseURL), "/")
	if baseURL == "" {
		baseURL = ai.DefaultOpenAICompatibleBaseURL(provider)
	}

	models, err := fetchProviderModels(r.Context(), baseURL, payload.APIKey)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"models": []string{},
			"error":  err.Error(),
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"models": models,
	})
}

// fetchProviderModels memanggil endpoint GET {baseURL}/models (OpenAI-compatible)
// dan mengembalikan daftar model ID.
func fetchProviderModels(ctx context.Context, baseURL, apiKey string) ([]string, error) {
	url := baseURL + "/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch models: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch models: status %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	var parsed struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("fetch models: parse: %w", err)
	}
	models := make([]string, 0, len(parsed.Data))
	for _, m := range parsed.Data {
		if m.ID != "" {
			models = append(models, m.ID)
		}
	}
	return models, nil
}

// handleExportJUnit exports a run's results as JUnit XML for CI/CD integration.
// GET /api/v1/runs/{id}/export-junit
func (s *Server) handleExportJUnit(w http.ResponseWriter, r *http.Request) {
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

	data, err := junit.GenerateXML(run)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate JUnit XML: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=junit-%s.xml", runID))
	w.Write(data)
}
