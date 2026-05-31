package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/compare"
	"github.com/go-go-golems/gotest-agent/internal/config"
	"github.com/go-go-golems/gotest-agent/internal/db"
	"github.com/go-go-golems/gotest-agent/internal/events"
	"github.com/go-go-golems/gotest-agent/internal/intelligence"
	"github.com/go-go-golems/gotest-agent/internal/metrics"
	"github.com/go-go-golems/gotest-agent/internal/notify"
	"github.com/go-go-golems/gotest-agent/internal/recordings"
	"github.com/go-go-golems/gotest-agent/internal/release"
	"github.com/go-go-golems/gotest-agent/internal/report"
	"github.com/go-go-golems/gotest-agent/internal/schedule"
	"github.com/go-go-golems/gotest-agent/internal/visual"
	"github.com/go-go-golems/gotest-agent/internal/webhook"
	"github.com/go-go-golems/gotest-agent/internal/workflow"
	"github.com/google/uuid"
)

type Server struct {
	router     *chi.Mux
	cfg        *config.Config
	store      db.RunStore
	events     *events.Store
	recordings *recordings.Store
	visuals    *visual.Store
	schedules  *schedule.Store
	releases   *release.Store
	notifs     *notify.Store
	reviews    *workflow.ReviewStore
	suites     *workflow.SuiteStore
}

func NewServer(cfg *config.Config, store db.RunStore) *Server {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	s := &Server{
		router:     r,
		cfg:        cfg,
		store:      store,
		events:     events.NewStore(),
		recordings: recordings.NewStore(),
		visuals:    visual.NewStore(),
		schedules:  schedule.NewStore(),
		releases:   release.NewStore(),
		notifs:     notify.NewStore(),
		reviews:    workflow.NewReviewStore(),
		suites:     workflow.NewSuiteStore(),
	}
	s.routes()
	return s
}

func (s *Server) Events() *events.Store         { return s.events }
func (s *Server) Recordings() *recordings.Store  { return s.recordings }
func (s *Server) Visuals() *visual.Store         { return s.visuals }
func (s *Server) Schedules() *schedule.Store     { return s.schedules }
func (s *Server) Releases() *release.Store       { return s.releases }
func (s *Server) Notifications() *notify.Store   { return s.notifs }

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Api-Key")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) routes() {
	s.router.Get("/health", s.handleHealth)

	s.router.Route("/api/v1", func(r chi.Router) {
		r.Use(s.apiKeyAuth)
		// Runs
		r.Post("/runs", s.handleCreateRun)
		r.Get("/runs", s.handleListRuns)
		r.Get("/runs/{id}", s.handleGetRun)
		r.Post("/runs/{id}/rerun", s.handleRerun)
		r.Get("/runs/{id}/stream", s.handleSSEStream)
		r.Get("/runs/{id}/events", s.handleGetEvents)
		r.Get("/runs/{id}/report", s.handleReport)
		r.Get("/runs/{id}/compare/{otherId}", s.handleCompare)
		r.Get("/runs/{id}/recordings", s.handleGetRecordings)
		r.Get("/runs/{id}/visual", s.handleGetVisualArtifacts)
		r.Delete("/runs/{id}", s.handleDeleteRun)
		r.Get("/recordings", s.handleListAllRecordings)
		// Schedules
		r.Post("/schedules", s.handleCreateSchedule)
		r.Get("/schedules", s.handleListSchedules)
		r.Get("/schedules/{id}", s.handleGetSchedule)
		r.Patch("/schedules/{id}", s.handleUpdateSchedule)
		r.Delete("/schedules/{id}", s.handleDeleteSchedule)
		r.Post("/schedules/{id}/run-now", s.handleRunNow)
		// Releases
		r.Post("/releases", s.handleCreateRelease)
		r.Get("/releases", s.handleListReleases)
		r.Get("/releases/{id}", s.handleGetRelease)
		r.Patch("/releases/{id}", s.handleUpdateRelease)
		r.Get("/releases/{id}/summary", s.handleReleaseSummary)
		// Notifications
		r.Get("/notifications", s.handleListNotifications)
		// Metrics
		r.Get("/metrics/summary", s.handleMetricsSummary)
		r.Get("/metrics/hotspots", s.handleMetricsHotspots)
		r.Get("/metrics/flaky", s.handleMetricsFlaky)
		r.Get("/metrics/trend", s.handleMetricsTrend)
		r.Get("/metrics/risk", s.handleMetricsRisk)
		r.Get("/metrics/recommendations", s.handleMetricsRecommendations)
		// Intelligence
		r.Get("/releases/{id}/confidence", s.handleReleaseConfidence)
		r.Get("/releases/{id}/risk", s.handleReleaseRisk)
		r.Post("/suite-selection", s.handleSuiteSelection)
		// Reviews
		r.Post("/reviews", s.handleCreateReview)
		r.Get("/runs/{id}/reviews", s.handleGetRunReviews)
		r.Post("/reviews/{id}/approve", s.handleApproveReview)
		r.Post("/reviews/{id}/reject", s.handleRejectReview)
		// Suites
		r.Post("/suites", s.handleCreateSuite)
		r.Get("/suites", s.handleListSuites)
		r.Get("/suites/{id}", s.handleGetSuite)
		r.Delete("/suites/{id}", s.handleDeleteSuite)
		// Alert rules
		r.Post("/alert-rules/evaluate", s.handleEvaluateAlertRules)
	})

	wh := webhook.NewGitHubHandler(s.cfg.APIKey, func(event webhook.PushEvent) { _ = event })
	s.router.Post("/api/v1/webhooks/github", wh.ServeHTTP)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleCreateRun(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProjectPath  string `json:"project_path"`
		Requirements string `json:"requirements"`
		Mode         string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	mode := req.Mode
	if mode == "" {
		mode = "simple"
	}
	run := &agent.TestRun{
		ID: uuid.New().String(), ProjectPath: req.ProjectPath,
		Requirements: req.Requirements, Mode: mode, State: agent.StateIdle,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := s.store.CreateRun(r.Context(), run); err != nil {
		http.Error(w, "store error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"run_id": run.ID, "state": string(run.State), "created_at": run.CreatedAt.Format(time.RFC3339)})
}

func (s *Server) handleListRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := s.store.ListRuns(r.Context(), 50, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if runs == nil {
		runs = []*agent.TestRun{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(runs)
}

func (s *Server) handleGetRun(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	run, err := s.store.GetRun(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(run)
}

func (s *Server) handleRerun(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	orig, err := s.store.GetRun(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	run := &agent.TestRun{
		ID: uuid.New().String(), ProjectPath: orig.ProjectPath,
		Requirements: orig.Requirements, Mode: orig.Mode, State: agent.StateIdle,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := s.store.CreateRun(r.Context(), run); err != nil {
		http.Error(w, "store error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"run_id": run.ID, "state": string(run.State), "created_at": run.CreatedAt.Format(time.RFC3339)})
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
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
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
		http.Error(w, "run A not found", http.StatusNotFound)
		return
	}
	runB, err := s.store.GetRun(r.Context(), idB)
	if err != nil {
		http.Error(w, "run B not found", http.StatusNotFound)
		return
	}

	result := compare.Compare(runA, runB)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
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

func (s *Server) handleReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	run, err := s.store.GetRun(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	report.GenerateHTML(w, run)
}

func (s *Server) handleDeleteRun(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) apiKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.APIKey == "" {
			next.ServeHTTP(w, r)
			return
		}
		key := r.Header.Get("X-Api-Key")
		if key != s.cfg.APIKey {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// --- Schedule handlers ---

func (s *Server) handleCreateSchedule(w http.ResponseWriter, r *http.Request) {
	var sch schedule.Schedule
	if err := json.NewDecoder(r.Body).Decode(&sch); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if sch.Enabled {
		sch.NextRunAt = schedule.CalcNextRun(sch.Frequency, sch.CronExpr, time.Now())
	}
	result := s.schedules.Create(&sch)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleListSchedules(w http.ResponseWriter, r *http.Request) {
	list := s.schedules.List()
	if list == nil {
		list = []*schedule.Schedule{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (s *Server) handleGetSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sch, ok := s.schedules.Get(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sch)
}

func (s *Server) handleUpdateSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var patch map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ok := s.schedules.Update(id, func(sch *schedule.Schedule) {
		if v, ok := patch["enabled"]; ok {
			sch.Enabled = v.(bool)
		}
		if v, ok := patch["name"]; ok {
			sch.Name = v.(string)
		}
		if v, ok := patch["webhook_url"]; ok {
			sch.WebhookURL = v.(string)
		}
		if v, ok := patch["notify_on_fail"]; ok {
			sch.NotifyOnFail = v.(bool)
		}
		if v, ok := patch["environment"]; ok {
			sch.Environment = v.(string)
		}
		if v, ok := patch["base_url"]; ok {
			sch.BaseURL = v.(string)
		}
	})
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	sch, _ := s.schedules.Get(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sch)
}

func (s *Server) handleDeleteSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !s.schedules.Delete(id) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleRunNow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sch, ok := s.schedules.Get(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	// Create a run from the schedule config
	run := &agent.TestRun{
		ID: uuid.New().String(), ProjectPath: sch.ProjectPath,
		Requirements: sch.Requirements, Mode: sch.Mode, State: agent.StateIdle,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := s.store.CreateRun(r.Context(), run); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Update schedule last run
	now := time.Now()
	s.schedules.Update(id, func(sc *schedule.Schedule) {
		sc.LastRunAt = &now
		sc.LastRunID = run.ID
		sc.LastRunStatus = string(run.State)
		sc.NextRunAt = schedule.CalcNextRun(sc.Frequency, sc.CronExpr, now)
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"run_id": run.ID, "state": string(run.State)})
}

// --- Release handlers ---

func (s *Server) handleCreateRelease(w http.ResponseWriter, r *http.Request) {
	var rel release.Release
	if err := json.NewDecoder(r.Body).Decode(&rel); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	result := s.releases.Create(&rel)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleListReleases(w http.ResponseWriter, r *http.Request) {
	list := s.releases.List()
	if list == nil {
		list = []*release.Release{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (s *Server) handleGetRelease(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	rel, ok := s.releases.Get(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rel)
}

func (s *Server) handleUpdateRelease(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var patch map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ok := s.releases.Update(id, func(rel *release.Release) {
		if v, ok := patch["status"]; ok {
			rel.Status = v.(string)
		}
		if v, ok := patch["name"]; ok {
			rel.Name = v.(string)
		}
	})
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	rel, _ := s.releases.Get(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rel)
}

func (s *Server) handleReleaseSummary(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	rel, ok := s.releases.Get(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	// Gather runs for this release
	var runs []*agent.TestRun
	for _, rid := range rel.RunIDs {
		if run, err := s.store.GetRun(r.Context(), rid); err == nil {
			runs = append(runs, run)
		}
	}
	sum := release.Summarize(rel, runs)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sum)
}

// --- Notification handlers ---

func (s *Server) handleListNotifications(w http.ResponseWriter, r *http.Request) {
	list := s.notifs.List()
	if list == nil {
		list = []notify.Notification{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// --- Metrics handlers ---

func (s *Server) handleMetricsSummary(w http.ResponseWriter, r *http.Request) {
	runs, _ := s.store.ListRuns(r.Context(), 1000, 0)
	sum := metrics.ComputeSummary(runs)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sum)
}

func (s *Server) handleMetricsHotspots(w http.ResponseWriter, r *http.Request) {
	runs, _ := s.store.ListRuns(r.Context(), 1000, 0)
	// Need full run data for hotspots
	var fullRuns []*agent.TestRun
	for _, run := range runs {
		if full, err := s.store.GetRun(r.Context(), run.ID); err == nil {
			fullRuns = append(fullRuns, full)
		}
	}
	hotspots := metrics.ComputeHotspots(fullRuns, 10)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hotspots)
}

func (s *Server) handleMetricsFlaky(w http.ResponseWriter, r *http.Request) {
	runs, _ := s.store.ListRuns(r.Context(), 1000, 0)
	var fullRuns []*agent.TestRun
	for _, run := range runs {
		if full, err := s.store.GetRun(r.Context(), run.ID); err == nil {
			fullRuns = append(fullRuns, full)
		}
	}
	flaky := metrics.DetectFlaky(fullRuns)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flaky)
}

func (s *Server) handleMetricsTrend(w http.ResponseWriter, r *http.Request) {
	runs, _ := s.store.ListRuns(r.Context(), 1000, 0)
	var fullRuns []*agent.TestRun
	for _, run := range runs {
		if full, err := s.store.GetRun(r.Context(), run.ID); err == nil {
			fullRuns = append(fullRuns, full)
		}
	}
	trend := metrics.ComputeTrend(fullRuns)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trend)
}

// --- Intelligence handlers ---

func (s *Server) handleMetricsRisk(w http.ResponseWriter, r *http.Request) {
	runs := s.getAllRuns(r)
	scheds := s.schedules.List()
	risks := intelligence.ComputeRisk(runs, scheds)
	if risks == nil {
		risks = []intelligence.RiskItem{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(risks)
}

func (s *Server) handleMetricsRecommendations(w http.ResponseWriter, r *http.Request) {
	runs := s.getAllRuns(r)
	scheds := s.schedules.List()
	risks := intelligence.ComputeRisk(runs, scheds)
	recs := intelligence.GenerateRecommendations(risks)
	if recs == nil {
		recs = []intelligence.Recommendation{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recs)
}

func (s *Server) handleReleaseConfidence(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	rel, ok := s.releases.Get(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
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
	json.NewEncoder(w).Encode(conf)
}

func (s *Server) handleReleaseRisk(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	rel, ok := s.releases.Get(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	var runs []*agent.TestRun
	for _, rid := range rel.RunIDs {
		if run, err := s.store.GetRun(r.Context(), rid); err == nil {
			runs = append(runs, run)
		}
	}
	risks := intelligence.ComputeRisk(runs, nil)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(risks)
}

func (s *Server) handleSuiteSelection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Mode      string   `json:"mode"`
		AllTests  []string `json:"all_tests"`
		FlakyTests []string `json:"flaky_tests"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	runs := s.getAllRuns(r)
	risks := intelligence.ComputeRisk(runs, nil)
	sel := intelligence.SelectSuite(intelligence.SelectionMode(req.Mode), req.AllTests, risks, req.FlakyTests)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sel)
}

// --- Review handlers ---

func (s *Server) handleCreateReview(w http.ResponseWriter, r *http.Request) {
	var rev workflow.Review
	if err := json.NewDecoder(r.Body).Decode(&rev); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	result := s.reviews.Create(&rev)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleGetRunReviews(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	revs := s.reviews.ByRun(id)
	if revs == nil {
		revs = []*workflow.Review{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(revs)
}

func (s *Server) handleApproveReview(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Reviewer string `json:"reviewer"`
		Comment  string `json:"comment"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if !s.reviews.Approve(id, req.Reviewer, req.Comment) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	rev, _ := s.reviews.Get(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rev)
}

func (s *Server) handleRejectReview(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Reviewer string `json:"reviewer"`
		Comment  string `json:"comment"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if !s.reviews.Reject(id, req.Reviewer, req.Comment) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	rev, _ := s.reviews.Get(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rev)
}

// --- Suite handlers ---

func (s *Server) handleCreateSuite(w http.ResponseWriter, r *http.Request) {
	var suite workflow.Suite
	if err := json.NewDecoder(r.Body).Decode(&suite); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	result := s.suites.Create(&suite)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleListSuites(w http.ResponseWriter, r *http.Request) {
	list := s.suites.List()
	if list == nil {
		list = []*workflow.Suite{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (s *Server) handleGetSuite(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	suite, ok := s.suites.Get(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suite)
}

func (s *Server) handleDeleteSuite(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !s.suites.Delete(id) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Alert rules handler ---

func (s *Server) handleEvaluateAlertRules(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Rules []intelligence.AlertRule `json:"rules"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	runs := s.getAllRuns(r)
	sum := metrics.ComputeSummary(runs)
	allRuns := s.getAllRuns(r)
	risks := intelligence.ComputeRisk(allRuns, s.schedules.List())
	avgRisk := 0.0
	if len(risks) > 0 {
		for _, ri := range risks[:min(len(risks), 5)] {
			avgRisk += ri.RiskScore
		}
		avgRisk /= float64(min(len(risks), 5))
	}
	triggers := intelligence.EvaluateAlertRules(req.Rules, sum.PassRate, sum.TotalFailed, avgRisk)
	if triggers == nil {
		triggers = []intelligence.AlertTrigger{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(triggers)
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
