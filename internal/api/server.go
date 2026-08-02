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
	"github.com/go-go-golems/gotest-agent/internal/ai"
	"github.com/go-go-golems/gotest-agent/internal/appmetrics"
	"github.com/go-go-golems/gotest-agent/internal/audit"
	"github.com/go-go-golems/gotest-agent/internal/auth"
	"github.com/go-go-golems/gotest-agent/internal/config"
	"github.com/go-go-golems/gotest-agent/internal/db"
	"github.com/go-go-golems/gotest-agent/internal/drift"
	"github.com/go-go-golems/gotest-agent/internal/events"
	"github.com/go-go-golems/gotest-agent/internal/llmprofile"
	"github.com/go-go-golems/gotest-agent/internal/notify"
	"github.com/go-go-golems/gotest-agent/internal/planning"
	"github.com/go-go-golems/gotest-agent/internal/project"
	"github.com/go-go-golems/gotest-agent/internal/recordings"
	"github.com/go-go-golems/gotest-agent/internal/release"
	"github.com/go-go-golems/gotest-agent/internal/schedule"
	"github.com/go-go-golems/gotest-agent/internal/steel"
	"github.com/go-go-golems/gotest-agent/internal/tracing"
	"github.com/go-go-golems/gotest-agent/internal/visual"
	"github.com/go-go-golems/gotest-agent/internal/webhook"
	"github.com/go-go-golems/gotest-agent/internal/workflow"
	"github.com/google/uuid"
)

type Server struct {
	router        *chi.Mux
	cfg           *config.Config
	store         db.RunStore
	settings      *db.SettingsStore
	projects      project.Store
	planning      planning.Store
	events        *events.Store
	recordings    *recordings.Store
	visuals       *visual.Store
	schedules     schedule.Repository
	releases      *release.Store
	notifs        *notify.Store
	drifts        *drift.Store
	driftTests    *drift.GeneratedTestStore
	driftDetector *drift.Detector
	webhooks      *webhook.RegistrationStore
	reviews       *workflow.ReviewStore
	suites        *workflow.SuiteStore
	auditLog      *audit.Store
	keyStore      *auth.KeyStore
	oidcManager   *auth.OIDCManager
	jwtAuth       *auth.Auth
	steel         *steel.Client                 // Steel Browser client (remote headless browser via CDP); nil if not configured
	llmProfiles   *llmprofile.Store             // LLM provider profiles (multi-provider)
	userStore     *auth.UserStore               // dashboard users (email+password login)
	runSem        chan struct{}                 // concurrency cap for run goroutines (AUDIT S-01)
	enqueueRun    func(runID string) error      // optional durable-queue enqueuer (Redis/Asynq); nil = in-process execution
	runCancels    map[string]context.CancelFunc // active run cancellation functions
	cancelsMu     sync.RWMutex
	metrics       *appmetrics.Metrics                 // Prometheus-format application metrics
	aiClientFn    func(ctx context.Context) ai.Client // test hook for drift generation; nil uses aiClient
}

func NewServer(cfg *config.Config, store db.RunStore, settingsStore *db.SettingsStore) *Server {
	appm := appmetrics.New()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(tracing.HTTPMiddleware())              // distributed tracing for all HTTP requests
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

	// Init Steel Browser client (remote headless browser) when configured.
	// Used by SteelRunner to execute UI tests via CDP without a local browser.
	var steelClient *steel.Client
	if cfg.SteelAPIURL != "" {
		steelClient = steel.NewClient(cfg.SteelAPIURL, cfg.SteelAPIKey)
	}

	s := &Server{
		router:      r,
		cfg:         cfg,
		store:       store,
		settings:    settingsStore,
		projects:    projectStore,
		planning:    planningStore,
		events:      evtStore,
		recordings:  recordings.NewStore(),
		visuals:     visual.NewStore(),
		schedules:   scheduleStore,
		releases:    release.NewStore(),
		notifs:      notify.NewStore(),
		drifts:      drift.NewStore(),
		driftTests:  drift.NewGeneratedTestStore(),
		webhooks:    webhook.NewRegistrationStore(),
		reviews:     workflow.NewReviewStore(),
		suites:      workflow.NewSuiteStore(),
		auditLog:    audit.NewStore(),
		keyStore:    auth.NewKeyStore(),
		oidcManager: auth.NewOIDCManager(),
		jwtAuth:     jwtAuth,
		steel:       steelClient,
		llmProfiles: llmprofile.NewStore(),
		userStore:   auth.NewUserStore(),
		runSem:      make(chan struct{}, cfg.MaxConcurrentRuns),
		runCancels:  make(map[string]context.CancelFunc),
		metrics:     appm,
	}
	if pgStore, ok := store.(*db.Store); ok {
		pool := pgStore.Pool()
		s.recordings.EnableDB(pool)
		s.webhooks.EnableDB(pool)
		s.drifts.EnableDB(pool)
		s.driftTests.EnableDB(pool)
		s.releases.EnableDB(pool)
		s.reviews.EnableDB(pool)
		s.suites.EnableDB(pool)
		s.auditLog.EnableDB(pool)
		s.llmProfiles.EnableDB(pool)
		s.keyStore.EnableDB(pool)
		s.userStore.EnableDB(pool)
	}
	s.driftDetector = drift.NewDetector(s.drifts)

	// Bootstrap admin default untuk login email+password (first-run setup).
	// Password default memakai API_KEY; bisa diganti lewat manajemen user.
	s.userStore.SeedDefaultAdmin("admin@gotest.local", cfg.APIKey, "Administrator")
	s.keyStore.SeedDefaultKey()
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
				// Echo the request's Origin header rather than "*" so
				// credentials: include is compatible (wildcard + credentials
				// is blocked by browsers per the Fetch spec).
				origin := strings.TrimSuffix(r.Header.Get("Origin"), "/")
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Add("Vary", "Origin")
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
	s.router.Get("/metrics", s.handlePrometheus)
	s.router.Get("/api/v1/stream", s.handleGlobalStream) // SSE: public, outside auth group // Prometheus scrape endpoint (no auth)

	// Auth: login endpoint outside API key auth — login is how you get the JWT
	s.router.Post("/api/v1/auth/login", s.handleLogin)
	s.router.Post("/api/v1/auth/logout", s.handleLogout)

	// OIDC/SSO: public endpoints for OAuth2 flow (no auth required)
	s.router.Get("/auth/oidc/providers", s.handleOIDCProviders)
	s.router.Get("/auth/oidc/login", s.handleOIDCLogin)
	s.router.Get("/auth/oidc/callback", s.handleOIDCCallback)

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
		// New record/playback sessions (Phase 2). /recordings remains backward-compatible screenshot metadata.
		r.Post("/recording-sessions", s.handleCreateRecordingSession)
		r.Get("/recording-sessions", s.handleListRecordingSessions)
		r.Get("/recording-sessions/{id}", s.handleGetRecordingSession)
		r.Post("/recording-sessions/{id}/events", s.handleAddRecordingEvent)
		r.Get("/recording-sessions/{id}/events", s.handleListRecordingEvents)
		r.Post("/recording-sessions/{id}/generate", s.handleGenerateTestFromRecording)
		r.Delete("/recording-sessions/{id}", s.handleDeleteRecordingSession)
		r.Patch("/recording-sessions/{id}", s.handleUpdateRecordingSession)
		r.Get("/runs/{id}/visual", s.handleGetVisualArtifacts)
		r.Get("/runs/{id}/video", s.handleGetVideoMetadata)
		r.Delete("/runs/{id}", s.handleDeleteRun)
		r.Get("/recordings", s.handleListAllRecordings)
		// Global live stream (also available at /api/v1/stream for public SSE access)
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
		r.Get("/notifications/types", s.handleListNotificationTypes)
		// Drift detection (Phase 3)
		r.Get("/drifts", s.handleListDrifts)
		r.Patch("/drifts/{id}", s.handleUpdateDriftStatus)
		r.Post("/drifts/{id}/generate-test", s.handleGenerateDriftTest)
		r.Get("/drifts/{id}/auto-generate", s.handleAutoGenerateDriftTest)
		r.Get("/drifts/{id}/generated-tests", s.handleListDriftTests)
		r.Patch("/generated-tests/{id}", s.handleUpdateDriftTestStatus)
		// Webhook registration (Phase 3 continuous sync)
		r.Post("/webhooks/register", s.handleRegisterWebhook)
		r.Get("/webhooks", s.handleListWebhooks)
		r.Get("/webhooks/{id}", s.handleGetWebhook)
		r.Patch("/webhooks/{id}/status", s.handleUpdateWebhookStatus)
		r.Delete("/webhooks/{id}", s.handleDeleteWebhook)
		r.Post("/webhooks/{id}/sync", s.handleSyncWebhook)
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
		r.Get("/intelligence/quality", s.handleTestQuality)
		r.Get("/intelligence/redundancy", s.handleTestRedundancy)
		r.Post("/intelligence/review", s.handleReviewTestCode)
		r.Get("/runs/{id}/review", s.handleReviewTestRun)
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
		// Settings (admin-only for modification, all roles for view outside this group)
		r.Get("/settings", s.handleGetSettings)
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireRole(auth.RoleAdmin))
			r.Put("/settings", s.handleUpdateSettings)
			r.Get("/ai/providers", s.handleListAIProviders)
			r.Post("/ai/test-provider", s.handleTestAIProvider)
			r.Post("/ai/models", s.handleListProviderModels)
			// LLM provider profiles (multi-provider)
			r.Get("/ai/profiles", s.handleListLLMProfiles)
			r.Post("/ai/profiles", s.handleCreateLLMProfile)
			r.Put("/ai/profiles/{id}", s.handleUpdateLLMProfile)
			r.Delete("/ai/profiles/{id}", s.handleDeleteLLMProfile)
			r.Post("/ai/profiles/{id}/activate", s.handleActivateLLMProfile)
			r.Post("/ai/profiles/{id}/test", s.handleTestLLMProfile)
			// User management (admin-only)
			r.Get("/users", s.handleListUsers)
			r.Post("/users", s.handleCreateUser)
			r.Put("/users/{id}", s.handleUpdateUser)
			r.Delete("/users/{id}", s.handleDeleteUser)
			// Audit log (admin-only)
			r.Get("/audit-log", s.handleListAuditLog)
			r.Get("/audit-log/users/{actorID}", s.handleListAuditLogByActor)
			r.Get("/audit-log/{resource}/{resourceID}", s.handleListAuditLogByResource)
			// API Key management (admin-only)
			r.Post("/keys", s.handleCreateAPIKey)
			r.Get("/keys", s.handleListAPIKeys)
			r.Post("/keys/{id}/revoke", s.handleRevokeAPIKey)
			r.Delete("/keys/{id}", s.handleDeleteAPIKey)
		})
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
		// Prefer AI test generation (clone, parse, synthesize plan); fall
		// back to a plain auto-triggered run when unavailable or failing.
		generated := s.processPushWithTestGen(event)
		// Phase 3: detect code/test drift. Runs after the test-gen clone
		// attempt so on-disk test checks see the repository.
		s.detectDriftFromPush(event)
		if generated {
			return
		}
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
