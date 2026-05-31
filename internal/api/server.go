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
	"github.com/go-go-golems/gotest-agent/internal/recordings"
	"github.com/go-go-golems/gotest-agent/internal/report"
	"github.com/go-go-golems/gotest-agent/internal/visual"
	"github.com/go-go-golems/gotest-agent/internal/webhook"
	"github.com/google/uuid"
)

type Server struct {
	router     *chi.Mux
	cfg        *config.Config
	store      db.RunStore
	events     *events.Store
	recordings *recordings.Store
	visuals    *visual.Store
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
	}
	s.routes()
	return s
}

func (s *Server) Events() *events.Store         { return s.events }
func (s *Server) Recordings() *recordings.Store  { return s.recordings }
func (s *Server) Visuals() *visual.Store         { return s.visuals }

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
