package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/appmetrics"
	"github.com/go-go-golems/gotest-agent/internal/auth"
	"github.com/go-go-golems/gotest-agent/internal/config"
	"github.com/go-go-golems/gotest-agent/internal/db"
	"github.com/go-go-golems/gotest-agent/internal/events"
	"github.com/go-go-golems/gotest-agent/internal/notify"
	"github.com/go-go-golems/gotest-agent/internal/planning"
	"github.com/go-go-golems/gotest-agent/internal/project"
	"github.com/go-go-golems/gotest-agent/internal/recordings"
	"github.com/go-go-golems/gotest-agent/internal/release"
	"github.com/go-go-golems/gotest-agent/internal/schedule"
	"github.com/go-go-golems/gotest-agent/internal/tracing"
	"github.com/go-go-golems/gotest-agent/internal/visual"
	"github.com/go-go-golems/gotest-agent/internal/webhook"
	"github.com/go-go-golems/gotest-agent/internal/workflow"
	"github.com/google/uuid"
)

type Server struct {
	router     *chi.Mux
	cfg        *config.Config
	store      db.RunStore
	settings   *db.SettingsStore
	projects   project.Store
	planning   planning.Store
	events     *events.Store
	recordings *recordings.Store
	visuals    *visual.Store
	schedules  schedule.Repository
	releases   *release.Store
	notifs     *notify.Store
	reviews    *workflow.ReviewStore
	suites     *workflow.SuiteStore
	jwtAuth    *auth.Auth
	runSem     chan struct{}            // concurrency cap for run goroutines (AUDIT S-01)
	enqueueRun func(runID string) error // optional durable-queue enqueuer (Redis/Asynq); nil = in-process execution
	runCancels map[string]context.CancelFunc // active run cancellation functions
	cancelsMu  sync.RWMutex
	metrics    *appmetrics.Metrics // Prometheus-format application metrics
}

func NewServer(cfg *config.Config, store db.RunStore, settingsStore *db.SettingsStore) *Server {
	appm := appmetrics.New()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(tracing.HTTPMiddleware()) // distributed tracing for all HTTP requests
	r.Use(rateLimitMiddleware(100, time.Minute)) // 100 req/min per client IP
	r.Use(newCORSMiddleware(cfg.CORSAllowedOrigins))
	r.Use(instrumentMiddleware(appm)) // count API requests/errors for /metrics

	evtStore := events.NewStore()

	projectStore := project.Store(project.NewMemoryStore())
	planningStore := planning.Store(planning.NewMemoryStore())
	scheduleStore := schedule.Repository(schedule.NewStore())
	if pgStore, ok := store.(*db.Store); ok {
		pool := pgStore.Pool()
		projectStore = project.NewDBStore(pool)
		planningStore = planning.NewDBStore(pool)
		scheduleStore = schedule.NewDBStore(pool)
		evtStore.EnableDB(pool) // ADR-003 Phase 1: persist events to PostgreSQL
	}

	// Init JWT auth for dashboard cookie authentication
	jwtSecret := cfg.JWTSecret
	if jwtSecret == "" {
		jwtSecret = auth.GenerateJWTSecret()
	}
	jwtAuth := auth.New(jwtSecret)

	s := &Server{
		router:     r,
		cfg:        cfg,
		store:      store,
		settings:   settingsStore,
		projects:   projectStore,
		planning:   planningStore,
		events:     evtStore,
		recordings: recordings.NewStore(),
		visuals:    visual.NewStore(),
		schedules:  scheduleStore,
		releases:   release.NewStore(),
		notifs:     notify.NewStore(),
		reviews:    workflow.NewReviewStore(),
		suites:     workflow.NewSuiteStore(),
		jwtAuth:    jwtAuth,
		runSem:     make(chan struct{}, cfg.MaxConcurrentRuns),
		runCancels: make(map[string]context.CancelFunc),
		metrics:    appm,
	}
	s.routes()
	return s
}

func (s *Server) Events() *events.Store          { return s.events }
func (s *Server) Recordings() *recordings.Store  { return s.recordings }
func (s *Server) Visuals() *visual.Store         { return s.visuals }
func (s *Server) Schedules() schedule.Repository { return s.schedules }
func (s *Server) Releases() *release.Store       { return s.releases }
func (s *Server) Notifications() *notify.Store   { return s.notifs }

// newCORSMiddleware builds the CORS middleware from a comma-separated origin
// allowlist (CORS_ALLOWED_ORIGINS). Empty or "*" keeps the historical wildcard
// (development). With an allowlist, only matching Origin headers are echoed
// back (with Vary: Origin) — README security guidance DG-12.
func newCORSMiddleware(allowedOrigins string) func(http.Handler) http.Handler {
	allowedOrigins = strings.TrimSpace(allowedOrigins)
	wildcard := allowedOrigins == "" || allowedOrigins == "*"
	allowed := make(map[string]bool)
	if !wildcard {
		for _, o := range strings.Split(allowedOrigins, ",") {
			if o = strings.TrimSpace(strings.TrimSuffix(o, "/")); o != "" {
				allowed[strings.ToLower(o)] = true
			}
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if wildcard {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				origin := strings.TrimSuffix(r.Header.Get("Origin"), "/")
				if allowed[strings.ToLower(origin)] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				w.Header().Add("Vary", "Origin")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Api-Key")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// isValidID validates that an ID param is safe (no path traversal, reasonable length)
func isValidID(id string) bool {
	if len(id) < 1 || len(id) > 64 {
		return false
	}
	for _, c := range id {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// --- HTTP response helpers ---

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, errorResponse{Error: message})
}

// bodyLimitMiddleware restricts request body size to maxBytes. 1 MiB default.
func bodyLimitMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// safeString extracts a string from map[string]interface{}, returns false on type mismatch.
func safeString(patch map[string]interface{}, key string) (string, bool) {
	v, ok := patch[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// safeBool extracts a bool from map[string]interface{}, returns false on type mismatch.
func safeBool(patch map[string]interface{}, key string) (bool, bool) {
	v, ok := patch[key]
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

func (s *Server) routes() {
	s.router.Get("/health", s.handleHealth)
	s.router.Get("/metrics", s.handlePrometheus) // Prometheus scrape endpoint (no auth)

	// Auth: login endpoint outside API key auth — login is how you get the JWT
	s.router.Post("/api/v1/auth/login", s.handleLogin)
	s.router.Post("/api/v1/auth/logout", s.handleLogout)

	s.router.Route("/api/v1", func(r chi.Router) {
		r.Use(s.apiKeyAuth)
		r.Use(bodyLimitMiddleware(1 << 20)) // 1 MiB body limit
		// Projects
		r.Post("/projects", s.handleCreateProject)
		r.Get("/projects", s.handleListProjects)
		r.Get("/projects/{id}", s.handleGetProject)
		r.Patch("/projects/{id}", s.handleUpdateProject)
		r.Post("/projects/{id}/api-docs", s.handleUploadAPIDocs)
		r.Post("/projects/{id}/parse-api", s.handleParseAPIDocs)
		r.Post("/projects/{id}/extract-features", s.handleExtractProjectFeatures)
		r.Post("/projects/{id}/test-plan", s.handleGenerateProjectTestPlan)
		r.Get("/test-plans/{id}", s.handleGetTestPlan)
		r.Patch("/test-plans/{id}/cases/{caseId}", s.handleUpdateTestPlanCase)
		r.Post("/test-plans/{id}/regenerate", s.handleRegenerateTestPlan)
		r.Post("/test-plans/{id}/approve", s.handleApproveTestPlan)
		r.Get("/test-cases", s.handleListTestCases)
		r.Patch("/test-cases/{id}", s.handleUpdateTestCase)
		r.Get("/test-cases/maintenance", s.handleTestCaseMaintenance)
		r.Get("/test-cases/{id}", s.handleGetTestCase)
		r.Post("/test-cases/{id}/run", s.handleRunTestCase)
		r.Post("/test-cases/{id}/refine", s.handleRefineTestCase)
		r.Get("/test-cases/{id}/proposals", s.handleListTestCaseProposals)
		r.Get("/change-proposals", s.handleListChangeProposals)
		r.Post("/change-proposals/{id}/approve", s.handleApproveChangeProposal)
		r.Post("/change-proposals/{id}/reject", s.handleRejectChangeProposal)
		r.Post("/test-lists", s.handleCreateTestList)
		r.Get("/test-lists", s.handleListTestLists)
		r.Get("/test-lists/{id}", s.handleGetTestList)
		r.Get("/test-lists/{id}/history", s.handleTestListHistory)
		r.Post("/test-lists/{id}/run", s.handleRunTestList)
		// Runs
		r.Post("/runs", s.handleCreateRun)
		r.Get("/runs", s.handleListRuns)
		r.Get("/runs/{id}", s.handleGetRun)
		r.Post("/runs/{id}/rerun", s.handleRerun)
		r.Post("/runs/{id}/cancel", s.handleCancelRun)
		r.Get("/runs/{id}/stream", s.handleSSEStream)
		r.Get("/runs/{id}/events", s.handleGetEvents)
		r.Get("/runs/{id}/report", s.handleReport)
		r.Post("/runs/{id}/analyze-failure", s.handleAnalyzeFailure)
		r.Get("/runs/{id}/compare/{otherId}", s.handleCompare)
		r.Get("/runs/{id}/recordings", s.handleGetRecordings)
		r.Get("/runs/{id}/visual", s.handleGetVisualArtifacts)
		r.Get("/runs/{id}/video", s.handleGetVideoMetadata)
		r.Delete("/runs/{id}", s.handleDeleteRun)
		r.Get("/recordings", s.handleListAllRecordings)
		// Global live stream
		r.Get("/stream", s.handleGlobalStream)
		r.Get("/monitoring/summary", s.handleMonitoringSummary)
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
		r.Get("/releases/{id}/explanation", s.handleReleaseExplanation)
		r.Post("/suite-selection", s.handleSuiteSelection)
		// Reviews
		r.Post("/reviews", s.handleCreateReview)
		r.Get("/runs/{id}/reviews", s.handleGetRunReviews)
		r.Post("/reviews/{id}/approve", s.handleApproveReview)
		r.Post("/reviews/{id}/reject", s.handleRejectReview)
		r.Post("/reviews/{id}/request-changes", s.handleRequestChangesReview)
		r.Get("/reviews", s.handleListAllReviews)
		// Suites
		r.Post("/suites", s.handleCreateSuite)
		r.Get("/suites", s.handleListSuites)
		r.Get("/suites/{id}", s.handleGetSuite)
		r.Delete("/suites/{id}", s.handleDeleteSuite)
		// Alert rules
		r.Post("/alert-rules/evaluate", s.handleEvaluateAlertRules)
		// Settings
		r.Get("/settings", s.handleGetSettings)
		r.Put("/settings", s.handleUpdateSettings)
		r.Get("/ai/providers", s.handleListAIProviders)
		r.Post("/ai/test-provider", s.handleTestAIProvider)
		// Demo
		r.Post("/demo/seed", s.handleDemoSeed)
		// Advanced testing (Phase 2+3)
		s.registerAdvancedTestingRoutes(r)
		// Export
		r.Get("/runs/{id}/export", s.handleExportRun)
		r.Get("/runs/{id}/export-junit", s.handleExportJUnit)
		r.Get("/runs/{id}/compare/{otherId}/export", s.handleExportCompare)
		r.Get("/metrics/risk/export", s.handleExportRisk)
		r.Get("/releases/{id}/confidence/export", s.handleExportConfidence)
	})

	webhookSecret := s.cfg.GitHubWebhookSecret
	if webhookSecret == "" {
		webhookSecret = s.cfg.APIKey // fallback: existing deployments that only set API_KEY
	}
	wh := webhook.NewGitHubHandler(webhookSecret, func(event webhook.PushEvent) {
		// Auto-trigger test run on github push event
		run := &agent.TestRun{
			ID:           uuid.New().String(),
			ProjectPath:  event.Repository.CloneURL,
			Requirements: "Auto-triggered via GitHub push to " + event.Ref,
			Mode:         "simple",
			TestType:     "ui",
			State:        agent.StateIdle,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := s.store.CreateRun(context.Background(), run); err != nil {
			return
		}
		s.events.Emit(run.ID, "run_started", "idle", "Run auto-created via GitHub Webhook", map[string]string{"project": event.Repository.CloneURL, "mode": "simple"})
		s.launchRun(run)
	})
	s.router.Post("/api/v1/webhooks/github", wh.ServeHTTP)

	// Serve video files (behind API key auth)
	s.router.Group(func(r chi.Router) {
		r.Use(s.apiKeyAuth)
		r.Handle("/videos/*", http.StripPrefix("/videos/", http.FileServer(http.Dir("/data/videos"))))
	})
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
