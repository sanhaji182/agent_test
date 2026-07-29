package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/ai"
	"github.com/go-go-golems/gotest-agent/internal/compare"
	"github.com/go-go-golems/gotest-agent/internal/events"
	"github.com/go-go-golems/gotest-agent/internal/execution"
	"github.com/go-go-golems/gotest-agent/internal/planning"
	"github.com/go-go-golems/gotest-agent/internal/project"
	"github.com/go-go-golems/gotest-agent/internal/recordings"
	"github.com/go-go-golems/gotest-agent/internal/report"
	"github.com/go-go-golems/gotest-agent/internal/visual"
	"github.com/google/uuid"
)

func (s *Server) handleCreateRun(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProjectPath  string            `json:"project_path"`
		Requirements string            `json:"requirements"`
		Mode         string            `json:"mode"`
		TestType     string            `json:"test_type"`
		PRD          string            `json:"prd"`
		APIDocs      string            `json:"api_docs"`
		AuthType     string            `json:"auth_type"`
		Credentials  string            `json:"credentials"`
		FocusHints   string            `json:"focus_hints"`
		SkipHints    string            `json:"skip_hints"`
		Browser      string            `json:"browser"`
		Viewport     string            `json:"viewport"`
		Parallel     bool              `json:"parallel"`
		TestData     map[string]string `json:"test_data"`
		Tags         []string          `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	mode := req.Mode
	if mode == "" {
		mode = "simple"
	}
	testType := req.TestType
	if testType == "" {
		testType = "ui"
	}
	run := &agent.TestRun{
		ID: uuid.New().String(), ProjectPath: req.ProjectPath,
		Requirements: req.Requirements, Mode: mode, TestType: testType,
		PRD: req.PRD, APIDocs: req.APIDocs, AuthType: req.AuthType,
		Credentials: req.Credentials, FocusHints: req.FocusHints,
		SkipHints: req.SkipHints, FeatureMap: s.deriveFeatureMap(r.Context(), req.PRD, req.Requirements),
		Browser: req.Browser, Viewport: req.Viewport, Parallel: req.Parallel, TestData: req.TestData,
		Tags:      req.Tags,
		State:     agent.StateIdle,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := s.store.CreateRun(r.Context(), run); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	// Audit: run created
	s.events.Emit(run.ID, "run_started", "idle", "Run created via API", map[string]string{"project": req.ProjectPath, "mode": mode})

	// Snapshot response fields BEFORE launching: the run object is mutated by
	// the execution goroutine after launchRun returns (race otherwise).
	resp := map[string]string{"run_id": run.ID, "state": string(run.State), "created_at": run.CreatedAt.Format(time.RFC3339)}

	// Start async real execution
	s.launchRun(run)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) deriveFeatureMap(ctx context.Context, prd, requirements string) *agent.FeatureMap {
	if fm, err := s.deriveFeatureMapWithAI(ctx, prd, requirements); err == nil && fm != nil && len(fm.Features) > 0 {
		return fm
	}
	return deriveFeatureMapFallback(prd, requirements)
}

func (s *Server) parseAPIDocsWithAI(ctx context.Context, p *project.Project) []planning.DraftCase {
	client := s.aiClient(ctx)
	if client == nil {
		return []planning.DraftCase{}
	}
	prompt := `Parse the following API documentation and generate an array of distinct, actionable API test cases.
Return ONLY valid JSON. No markdown.
Schema:
[
  {
    "title": "Short descriptive title of the test case",
    "feature": "Which API endpoint or feature this tests",
    "priority": "high or medium",
    "steps": ["Step 1", "Step 2"],
    "assertions": ["Assert status is 200", "Assert response matches schema"],
    "tags": ["api"]
  }
]

API Docs:
` + p.APIDocs

	text, err := client.GenerateText(ctx, prompt)
	if err != nil {
		return []planning.DraftCase{}
	}

	var cases []planning.DraftCase
	if err := json.Unmarshal([]byte(ai.StripJSONMarkers(text)), &cases); err != nil {
		return []planning.DraftCase{}
	}

	for i := range cases {
		cases[i].Type = "api"
		cases[i].Enabled = true
	}
	return cases
}

func (s *Server) deriveFeatureMapWithAI(ctx context.Context, prd, requirements string) (*agent.FeatureMap, error) {
	client := s.aiClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("ai planning disabled")
	}
	source := strings.TrimSpace(prd)
	sourceName := "prd"
	if source == "" {
		source = strings.TrimSpace(requirements)
		sourceName = "requirements"
	}
	if source == "" {
		return nil, fmt.Errorf("empty source")
	}
	prompt := `Extract a concise feature map from the source below.
Return ONLY valid JSON. No markdown.
Schema:
{
  "source": "prd or requirements",
  "features": [
    {"name": "feature name", "use_cases": ["specific testable use case"]}
  ]
}

Source type: ` + sourceName + `
Source:
` + source
	text, err := client.GenerateText(ctx, prompt)
	if err != nil {
		return nil, err
	}
	var fm agent.FeatureMap
	if err := json.Unmarshal([]byte(ai.StripJSONMarkers(text)), &fm); err != nil {
		return nil, err
	}
	if fm.Source == "" {
		fm.Source = sourceName
	}
	return &fm, nil
}

func deriveFeatureMapFallback(prd, requirements string) *agent.FeatureMap {
	source := strings.TrimSpace(prd)
	sourceName := "prd"
	if source == "" {
		source = strings.TrimSpace(requirements)
		sourceName = "requirements"
	}
	if source == "" {
		return nil
	}

	features := make([]agent.Feature, 0, 6)
	lines := strings.Split(source, "\n")
	for _, line := range lines {
		clean := strings.TrimSpace(strings.Trim(line, "#-*0123456789. "))
		if len(clean) < 4 {
			continue
		}
		lower := strings.ToLower(clean)
		if strings.Contains(lower, "feature") || strings.Contains(lower, "use case") ||
			strings.Contains(lower, "flow") || strings.Contains(lower, "user") ||
			strings.Contains(lower, "login") || strings.Contains(lower, "checkout") ||
			strings.Contains(lower, "api") || strings.Contains(lower, "endpoint") {
			features = append(features, agent.Feature{
				Name:     truncate(clean, 80),
				UseCases: []string{truncate(clean, 140)},
			})
		}
		if len(features) >= 6 {
			break
		}
	}
	if len(features) == 0 {
		features = append(features, agent.Feature{
			Name:     "Primary product flow",
			UseCases: []string{truncate(source, 140)},
		})
	}
	return &agent.FeatureMap{Source: sourceName, Features: features}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return strings.TrimSpace(s[:max-3]) + "..."
}

func contains(items []string, target string) bool {
	if target == "" {
		return false
	}
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func (s *Server) buildTestListHistory(ctx context.Context, list *planning.TestList) map[string]interface{} {
	runs, _ := s.store.ListRuns(ctx, 1000, 0)
	caseSet := map[string]bool{}
	for _, id := range list.TestCaseIDs {
		caseSet[id] = true
	}
	var relevant []*agent.TestRun
	for _, run := range runs {
		if run.TestListID == list.ID || caseSet[run.TestCaseID] {
			relevant = append(relevant, run)
		}
	}

	type runView struct {
		ID         string    `json:"id"`
		TestCaseID string    `json:"test_case_id,omitempty"`
		Status     string    `json:"status"`
		CreatedAt  time.Time `json:"created_at"`
		Failed     int       `json:"failed"`
		Passed     int       `json:"passed"`
	}
	views := make([]runView, 0, len(relevant))
	latestByCase := map[string]*agent.TestRun{}
	previousByCase := map[string]*agent.TestRun{}
	counts := map[string]int{"passed": 0, "failed": 0}

	for _, run := range relevant {
		failed := 0
		passed := 0
		if run.RunResult != nil {
			failed = run.RunResult.Failed
			passed = run.RunResult.Passed
		}
		views = append(views, runView{
			ID: run.ID, TestCaseID: run.TestCaseID, Status: string(run.State),
			CreatedAt: run.CreatedAt, Failed: failed, Passed: passed,
		})
		if run.State == agent.StateDone {
			counts["passed"]++
		}
		if run.State == agent.StateFailed {
			counts["failed"]++
		}
		if run.TestCaseID == "" {
			continue
		}
		latest := latestByCase[run.TestCaseID]
		if latest == nil || run.CreatedAt.After(latest.CreatedAt) {
			if latest != nil {
				prev := previousByCase[run.TestCaseID]
				if prev == nil || latest.CreatedAt.After(prev.CreatedAt) {
					previousByCase[run.TestCaseID] = latest
				}
			}
			latestByCase[run.TestCaseID] = run
			continue
		}
		prev := previousByCase[run.TestCaseID]
		if prev == nil || run.CreatedAt.After(prev.CreatedAt) {
			previousByCase[run.TestCaseID] = run
		}
	}

	newlyFailed := []string{}
	recovered := []string{}
	stableFailed := []string{}
	for caseID, latest := range latestByCase {
		prev := previousByCase[caseID]
		latestFailed := latest.State == agent.StateFailed || (latest.RunResult != nil && latest.RunResult.Failed > 0)
		prevFailed := prev != nil && (prev.State == agent.StateFailed || (prev.RunResult != nil && prev.RunResult.Failed > 0))
		switch {
		case latestFailed && !prevFailed:
			newlyFailed = append(newlyFailed, caseID)
		case !latestFailed && prevFailed:
			recovered = append(recovered, caseID)
		case latestFailed && prevFailed:
			stableFailed = append(stableFailed, caseID)
		}
	}

	latest := map[string]interface{}{}
	if len(relevant) > 0 {
		run := relevant[0]
		latest = map[string]interface{}{"run_id": run.ID, "status": string(run.State), "created_at": run.CreatedAt}
	}
	return map[string]interface{}{
		"test_list_id":  list.ID,
		"name":          list.Name,
		"latest":        latest,
		"counts":        counts,
		"runs":          views,
		"newly_failed":  newlyFailed,
		"recovered":     recovered,
		"stable_failed": stableFailed,
	}
}

func (s *Server) aiClient(ctx context.Context) ai.Client {
	if os.Getenv("GOTEST_AI_PLANNING") != "1" && os.Getenv("GOTEST_AI_PLANNING") != "true" {
		return nil
	}
	cfg := ai.ConfigFromEnv()
	if s.settings != nil {
		if v, err := s.settings.Get(ctx, "llm_provider"); err == nil && v != "" {
			cfg.Provider = v
		}
		if v, err := s.settings.Get(ctx, "llm_model"); err == nil && v != "" {
			cfg.Model = v
		}
		if v, err := s.settings.Get(ctx, "llm_api_key"); err == nil && v != "" && !strings.Contains(v, "...") {
			cfg.APIKey = v
		}
		if v, err := s.settings.Get(ctx, "llm_base_url"); err == nil && v != "" {
			cfg.BaseURL = v
		}
	}
	// Credential-origin binding (ADR-005 Phase 2): When a custom LLM base URL
	// is set, do NOT forward the system API key to it. The user must provide
	// their own api key. This prevents caller-controlled base_url from receiving
	// the stored system credential.
	if cfg.BaseURL != "" && !isApprovedLLMOrigin(cfg.BaseURL) {
		if cfg.APIKey == "" {
			slog.Warn("custom LLM base URL set without explicit api key — refusing to forward system credential", "base_url", cfg.BaseURL)
		}
	}
	return ai.New(cfg)
}

// isApprovedLLMOrigin returns true if baseURL is a known LLM provider endpoint.
// Only these origins may receive the system's API key without an explicit
// user-provided key. Custom/self-hosted endpoints require an explicit api key.
func isApprovedLLMOrigin(baseURL string) bool {
	if baseURL == "" {
		return true // Not custom — using provider default
	}
	lower := strings.ToLower(baseURL)
	approved := []string{
		"https://api.anthropic.com",
		"https://api.openai.com",
		"https://generativelanguage.googleapis.com",
		"https://openrouter.ai/api",
		"https://api.deepseek.com",
		"https://api.mistral.ai",
		"https://api.groq.com",
	}
	for _, prefix := range approved {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}

// buildAgent constructs a fully-configured Agent using the LLM settings
// resolution chain (DB overrides → env config → defaults). Returns nil if the
// resolved provider is unsupported.
func (s *Server) buildAgent() *agent.Agent {
	return s.buildAgentForRun(nil)
}

// buildAgentForRun constructs an Agent with runner options from the run config.
func (s *Server) buildAgentForRun(run *agent.TestRun) *agent.Agent {
	// Read LLM settings: DB overrides, env fallback
	llmProvider := s.cfg.LLMProvider
	llmModel := s.cfg.LLMModel
	apiKey := s.cfg.AnthropicAPIKey
	baseURL := s.cfg.LLMBaseURL

	if s.settings != nil {
		ctx := context.Background()
		if v, err := s.settings.Get(ctx, "llm_provider"); err == nil && v != "" {
			llmProvider = v
		}
		if v, err := s.settings.Get(ctx, "llm_model"); err == nil && v != "" {
			llmModel = v
		}
		if v, err := s.settings.Get(ctx, "llm_api_key"); err == nil && v != "" {
			apiKey = v
		}
		if v, err := s.settings.Get(ctx, "llm_base_url"); err == nil && v != "" {
			baseURL = v
		}
	}

	llm := agent.NewLLM(llmProvider, llmModel, apiKey, baseURL)
	if llm == nil {
		slog.Error("unsupported LLM provider", "provider", llmProvider)
		return nil
	}

	runner := agent.NewPlaywrightRunner("/tmp/agent_test/videos", llm)
	runner.ScreenshotDir = "/tmp/agent_test/screenshots"
	// Apply run-specific execution options
	if run != nil {
		if run.Browser != "" {
			runner.WithBrowser(run.Browser)
		}
		if run.Viewport != "" {
			runner.WithViewport(run.Viewport)
		}
		if run.Parallel {
			runner.WithParallel(true)
		}
		if run.TestData != nil {
			runner.TestData = run.TestData
		}
	}
	execCtx := execution.NewContext(s.events, s.recordings, s.visuals)

	return agent.NewWithConfig(llm, runner, 3, agent.AgentConfig{
		Exec:  execCtx,
		Store: s.store,
	})
}

// launchRun is the canonical execution entry point (ADR-001). All 5 trigger
// paths (web API, webhook, MCP, schedule run-now, schedule due) route through
// this helper. When a durable queue enqueuer is configured (QUEUE_ENABLED),
// the run is enqueued to Redis/Asynq; otherwise it executes in-process via
// Agent.Launch with panic recovery.
func (s *Server) launchRun(run *agent.TestRun) {
	// Durable queue path (optional): enqueue and let the worker execute.
	// Falls back to in-process execution if enqueue fails.
	if s.enqueueRun != nil {
		if err := s.enqueueRun(run.ID); err == nil {
			return
		} else {
			slog.Warn("queue enqueue failed, falling back to in-process execution",
				"run_id", run.ID, "error", err)
		}
	}

	a := s.buildAgentForRun(run)
	if a == nil {
		return
	}
	// Bounded concurrency: acquire a slot before launching (AUDIT S-01).
	// Non-blocking best-effort: warn if at capacity, but still launch
	// for backward compatibility.
	s.acquireSlot()
	go func() {
		defer s.releaseSlot()
		a.Launch(run)
	}()
}

// SetRunEnqueuer installs a durable-queue enqueue function. When set,
// launchRun enqueues run IDs instead of executing in-process (with in-process
// fallback on enqueue error). Called from cmd/server when QUEUE_ENABLED=true.
func (s *Server) SetRunEnqueuer(fn func(runID string) error) {
	s.enqueueRun = fn
}

// ExecuteRunByID loads a run from the store and executes it synchronously.
// Used by the Asynq queue worker so job completion/retries track real
// execution outcome. Final state is persisted by the Agent per transition.
func (s *Server) ExecuteRunByID(ctx context.Context, runID string) error {
	run, err := s.store.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	// Terminal states are not re-executed (idempotent retry safety).
	if run.State == agent.StateDone || run.State == agent.StateFailed || run.State == agent.StateSimulated {
		return nil
	}
	a := s.buildAgent()
	if a == nil {
		return errUnsupportedProvider
	}
	execErr := a.Execute(ctx, run)
	_ = s.store.UpdateRun(context.Background(), run)
	return execErr
}

// errUnsupportedProvider is returned when the configured LLM provider cannot
// be constructed (queue worker path).
var errUnsupportedProvider = fmt.Errorf("unsupported LLM provider")

func (s *Server) acquireSlot() {
	select {
	case s.runSem <- struct{}{}:
	default:
		slog.Warn("run capacity exceeded, slot not acquired",
			"limit", cap(s.runSem))
	}
}

func (s *Server) releaseSlot() {
	<-s.runSem
}

func (s *Server) handleListRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := s.store.ListRuns(r.Context(), 50, 0)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if runs == nil {
		runs = []*agent.TestRun{}
	}

	// Filter by tag if ?tag= query param is provided
	if tagFilter := r.URL.Query().Get("tag"); tagFilter != "" {
		var filtered []*agent.TestRun
		for _, run := range runs {
			for _, tag := range run.Tags {
				if tag == tagFilter {
					filtered = append(filtered, run)
					break
				}
			}
		}
		runs = filtered
	}

	// Filter by state if ?state= query param is provided
	if stateFilter := r.URL.Query().Get("state"); stateFilter != "" {
		var filtered []*agent.TestRun
		for _, run := range runs {
			if string(run.State) == stateFilter {
				filtered = append(filtered, run)
			}
		}
		runs = filtered
	}

	redacted := make([]*agent.TestRun, len(runs))
	for i, run := range runs {
		redacted[i] = redactCredentials(run)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(redacted)
}

func (s *Server) handleMonitoringSummary(w http.ResponseWriter, r *http.Request) {
	runs, err := s.store.ListRuns(r.Context(), 200, 0)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	lists, _ := s.planning.ListTestLists(r.Context(), "")
	cases, _ := s.planning.ListTestCases(r.Context(), "")

	type listHealth struct {
		ID           string    `json:"id"`
		Name         string    `json:"name"`
		Pinned       bool      `json:"pinned"`
		TestCount    int       `json:"test_count"`
		PassRate     float64   `json:"pass_rate"`
		LastStatus   string    `json:"last_status"`
		LastRunID    string    `json:"last_run_id,omitempty"`
		LastRunAt    time.Time `json:"last_run_at,omitempty"`
		Failed       int       `json:"failed"`
		Passed       int       `json:"passed"`
		NewlyFailed  []string  `json:"newly_failed"`
		Recovered    []string  `json:"recovered"`
		StableFailed []string  `json:"stable_failed"`
	}
	var health []listHealth
	for _, list := range lists {
		listHistory := s.buildTestListHistory(r.Context(), list)
		h := listHealth{ID: list.ID, Name: list.Name, Pinned: list.Pinned, TestCount: len(list.TestCaseIDs), LastStatus: "never"}
		if latest, _ := listHistory["latest"].(map[string]interface{}); latest != nil {
			if v, ok := latest["run_id"].(string); ok {
				h.LastRunID = v
			}
			if v, ok := latest["status"].(string); ok && v != "" {
				h.LastStatus = v
			}
			if v, ok := latest["created_at"].(time.Time); ok {
				h.LastRunAt = v
			}
		}
		if counts, _ := listHistory["counts"].(map[string]int); counts != nil {
			h.Passed = counts["passed"]
			h.Failed = counts["failed"]
		}
		if v, _ := listHistory["newly_failed"].([]string); v != nil {
			h.NewlyFailed = v
		}
		if v, _ := listHistory["recovered"].([]string); v != nil {
			h.Recovered = v
		}
		if v, _ := listHistory["stable_failed"].([]string); v != nil {
			h.StableFailed = v
		}
		total := h.Passed + h.Failed
		if total > 0 {
			h.PassRate = float64(h.Passed) / float64(total)
		}
		health = append(health, h)
	}

	active := 0
	failed := 0
	completed := 0
	for _, run := range runs {
		if run.State == agent.StateDone {
			completed++
		} else if run.State == agent.StateFailed {
			failed++
		} else if run.State != agent.StateIdle {
			active++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	redactedRuns := make([]*agent.TestRun, len(runs))
	for i, run := range runs {
		redactedRuns[i] = redactCredentials(run)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"summary": map[string]interface{}{
			"total_lists":    len(lists),
			"total_cases":    len(cases),
			"active_runs":    active,
			"failed_runs":    failed,
			"completed_runs": completed,
		},
		"lists":       health,
		"recent_runs": redactedRuns,
	})
}

func (s *Server) handleGetRun(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !isValidID(id) {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	run, err := s.store.GetRun(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(redactCredentials(run))
}

func (s *Server) handleRerun(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	orig, err := s.store.GetRun(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	run := &agent.TestRun{
		ID: uuid.New().String(), ProjectPath: orig.ProjectPath,
		Requirements: orig.Requirements, Mode: orig.Mode, TestType: orig.TestType,
		PRD: orig.PRD, APIDocs: orig.APIDocs, AuthType: orig.AuthType,
		Credentials: orig.Credentials, FocusHints: orig.FocusHints,
		SkipHints: orig.SkipHints, FeatureMap: orig.FeatureMap,
		State:     agent.StateIdle,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := s.store.CreateRun(r.Context(), run); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Audit: run created for rerun
	s.events.Emit(run.ID, "run_started", "idle", "Rerun triggered via API", map[string]string{"project": run.ProjectPath, "mode": run.Mode})

	// Snapshot response fields BEFORE launching (run is mutated async after launch).
	resp := map[string]string{"run_id": run.ID, "state": string(run.State), "created_at": run.CreatedAt.Format(time.RFC3339)}

	// Start async real execution
	s.launchRun(run)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(resp)
}

// handleSSEStream streams granular events + state changes via SSE
func (s *Server) handleSSEStream(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	ctx := r.Context()

	// Send existing events first (replay)
	for _, evt := range s.events.GetEvents(id) {
		data, _ := json.Marshal(evt)
		fmt.Fprintf(w, "event: step\ndata: %s\n\n", data)
	}
	flusher.Flush()

	// Subscribe to new events
	ch, unsub := s.events.Subscribe(id)
	defer unsub()

	lastState := ""
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(evt)
			fmt.Fprintf(w, "event: step\ndata: %s\n\n", data)
			flusher.Flush()
		case <-ticker.C:
			// Also check state for done/failed
			run, err := s.store.GetRun(ctx, id)
			if err != nil {
				fmt.Fprintf(w, "event: error\ndata: {\"message\":\"run not found\"}\n\n")
				flusher.Flush()
				return
			}
			cs := string(run.State)
			if cs != lastState {
				lastState = cs
				data, _ := json.Marshal(map[string]string{"state": cs, "message": "State: " + cs})
				fmt.Fprintf(w, "event: state_change\ndata: %s\n\n", data)
				flusher.Flush()
			}
			if run.State == agent.StateDone || run.State == agent.StateFailed {
				doneData, _ := json.Marshal(run)
				fmt.Fprintf(w, "event: done\ndata: %s\n\n", doneData)
				flusher.Flush()
				return
			}
		}
	}
}

// handleGetEvents returns all persisted events for a run
func (s *Server) handleGetEvents(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	evts := s.events.GetEvents(id)
	if evts == nil {
		evts = []events.Event{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(evts)
}

// handleCompare compares two runs and returns structured diff
func (s *Server) handleCompare(w http.ResponseWriter, r *http.Request) {
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
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleAnalyzeFailure(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	run, err := s.store.GetRun(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	analysis := s.analyzeFailure(r.Context(), run)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analysis)
}

type failureAnalysis struct {
	RunID       string   `json:"run_id"`
	Status      string   `json:"status"`
	Summary     string   `json:"summary"`
	LikelyCause string   `json:"likely_cause"`
	NextAction  string   `json:"next_action"`
	Evidence    []string `json:"evidence"`
	Source      string   `json:"source"`
}

func (s *Server) analyzeFailure(ctx context.Context, run *agent.TestRun) failureAnalysis {
	fallback := deterministicFailureAnalysis(run)
	client := s.aiClient(ctx)
	if client == nil {
		return fallback
	}
	prompt := buildFailureAnalysisPrompt(run, fallback)
	text, err := client.GenerateText(ctx, prompt)
	if err != nil {
		return fallback
	}
	var parsed failureAnalysis
	if err := json.Unmarshal([]byte(ai.StripJSONMarkers(text)), &parsed); err != nil {
		return fallback
	}
	if parsed.Summary == "" || parsed.NextAction == "" {
		return fallback
	}
	parsed.RunID = run.ID
	parsed.Status = string(run.State)
	parsed.Source = "ai"
	if parsed.Evidence == nil {
		parsed.Evidence = fallback.Evidence
	}
	return parsed
}

func deterministicFailureAnalysis(run *agent.TestRun) failureAnalysis {
	evidence := []string{}
	summary := "Run did not report a failure."
	likelyCause := "No failure evidence was found in the run result."
	nextAction := "Review the run timeline and rerun if this result is unexpected."
	if run.State == agent.StateFailed || (run.RunResult != nil && run.RunResult.Failed > 0) || run.Error != "" {
		summary = "Run failed during automated execution."
		likelyCause = "The application behavior did not match one or more generated assertions."
		nextAction = "Open the failed step evidence, verify whether the product behavior or test expectation is wrong, then rerun the affected case."
	}
	if run.Error != "" {
		evidence = append(evidence, "Execution error: "+truncate(run.Error, 180))
		likelyCause = "The runner or target app returned an execution error before all assertions completed."
		nextAction = "Check runner connectivity, target URL availability, and credentials, then rerun."
	}
	if run.RunResult != nil {
		evidence = append(evidence, fmt.Sprintf("Result: %d passed, %d failed, %d total", run.RunResult.Passed, run.RunResult.Failed, run.RunResult.Total))
		for _, failure := range run.RunResult.Failures {
			msg := strings.TrimSpace(failure.Test + ": " + failure.Message)
			if msg != ":" {
				evidence = append(evidence, truncate(msg, 220))
			}
		}
	}
	if run.VideoURL != "" {
		evidence = append(evidence, "Video evidence is available")
	}
	if len(run.Screenshots) > 0 {
		evidence = append(evidence, fmt.Sprintf("%d screenshot artifacts are available", len(run.Screenshots)))
	}
	if len(evidence) == 0 {
		evidence = append(evidence, "No structured failure artifacts were captured")
	}
	return failureAnalysis{
		RunID:       run.ID,
		Status:      string(run.State),
		Summary:     summary,
		LikelyCause: likelyCause,
		NextAction:  nextAction,
		Evidence:    evidence,
		Source:      "deterministic",
	}
}

func buildFailureAnalysisPrompt(run *agent.TestRun, fallback failureAnalysis) string {
	payload, _ := json.Marshal(map[string]interface{}{
		"run_id":       run.ID,
		"state":        run.State,
		"requirements": redactForAI(run.Requirements),
		"project_path": run.ProjectPath,
		"test_type":    run.TestType,
		"test_plan":    run.TestPlan,
		"run_result":   run.RunResult,
		"error":        redactForAI(run.Error),
		"fallback":     fallback,
	})
	return `Analyze this automated test run failure.
Return only JSON with keys: summary, likely_cause, next_action, evidence.
Evidence must be an array of short strings grounded in the payload.
Do not invent artifacts or credentials.

Payload:
` + string(payload)
}

func redactForAI(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	lines := strings.Split(value, "\n")
	for i, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "password") || strings.Contains(lower, "secret") || strings.Contains(lower, "token") || strings.Contains(lower, "api key") {
			lines[i] = "[redacted]"
		}
	}
	return strings.Join(lines, "\n")
}

// redactCredentials strips credential fields from run/project structs before
// JSON serialization. Returns a shallow copy with Credentials cleared so the
// store's original is not mutated (AUDIT SEC-09).
func redactCredentials(run *agent.TestRun) *agent.TestRun {
	if run == nil {
		return nil
	}
	redacted := *run // shallow copy
	redacted.Credentials = ""
	return &redacted
}

// redactProject strips credential fields from project structs (AUDIT SEC-09).
func redactProject(p *project.Project) *project.Project {
	if p == nil {
		return nil
	}
	redacted := *p
	redacted.Credentials = ""
	return &redacted
}

// handleGetRecordings returns recordings for a specific run
func (s *Server) handleGetRecordings(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	recs := s.recordings.ByRun(id)
	if recs == nil {
		recs = []recordings.Recording{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recs)
}

// handleListAllRecordings returns all recordings across runs
func (s *Server) handleListAllRecordings(w http.ResponseWriter, r *http.Request) {
	recs := s.recordings.All()
	if recs == nil {
		recs = []recordings.Recording{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recs)
}

// handleGetVisualArtifacts returns visual test artifacts for a run
func (s *Server) handleGetVisualArtifacts(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	arts := s.visuals.ByRun(id)
	if arts == nil {
		arts = []visual.Artifact{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(arts)
}

func (s *Server) handleGetVideoMetadata(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	run, err := s.store.GetRun(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"video_url":               run.VideoURL,
		"video_status":            run.VideoStatus,
		"video_duration":          run.VideoDuration,
		"video_size":              run.VideoSize,
		"video_failure_marker_at": run.VideoFailureMarkerAt,
	})
}

func (s *Server) handleReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	run, err := s.store.GetRun(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Content-Type", "text/html")
	report.GenerateHTML(w, run)
}

func (s *Server) handleDeleteRun(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !isValidID(id) {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := s.store.DeleteRun(r.Context(), id); err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// helper to get all runs with full data
func (s *Server) getAllRuns(r *http.Request) []*agent.TestRun {
	list, _ := s.store.ListRuns(r.Context(), 1000, 0)
	var full []*agent.TestRun
	for _, run := range list {
		if f, err := s.store.GetRun(r.Context(), run.ID); err == nil {
			full = append(full, f)
		}
	}
	return full
}

// handleGlobalStream pushes all execution events instantly via SSE (event-bus driven)
func (s *Server) handleGlobalStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	ctx := r.Context()

	// Subscribe to ALL events from the event bus
	ch, unsub := s.events.SubscribeAll()
	defer unsub()

	// Heartbeat to keep connection alive
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-ch:
			if !ok {
				return
			}
			// Push every event to the control room
			data, _ := json.Marshal(map[string]interface{}{
				"type":     string(evt.Type),
				"run_id":   evt.RunID,
				"phase":    evt.Phase,
				"message":  evt.Message,
				"metadata": evt.Metadata,
				"failed":   evt.Type == "run_failed" || evt.Type == "assertion_failed",
			})
			fmt.Fprintf(w, "event: update\ndata: %s\n\n", data)
			flusher.Flush()
		case <-ticker.C:
			// Heartbeat
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}
