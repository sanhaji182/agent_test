package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/config"
	"github.com/go-go-golems/gotest-agent/internal/db"
	"github.com/go-go-golems/gotest-agent/internal/report"
	"github.com/go-go-golems/gotest-agent/internal/webhook"
	"github.com/google/uuid"
)

type Server struct {
	router *chi.Mux
	cfg    *config.Config
	store  db.RunStore
}

func NewServer(cfg *config.Config, store db.RunStore) *Server {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	s := &Server{router: r, cfg: cfg, store: store}
	s.routes()
	return s
}

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
		r.Get("/runs/{id}/report", s.handleReport)
		r.Delete("/runs/{id}", s.handleDeleteRun)
	})

	wh := webhook.NewGitHubHandler(s.cfg.APIKey, func(event webhook.PushEvent) {
		_ = event
	})
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
		ID:           uuid.New().String(),
		ProjectPath:  req.ProjectPath,
		Requirements: req.Requirements,
		Mode:         mode,
		State:        agent.StateIdle,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.store.CreateRun(r.Context(), run); err != nil {
		http.Error(w, "store error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"run_id":     run.ID,
		"state":      string(run.State),
		"created_at": run.CreatedAt.Format(time.RFC3339),
	})
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

// handleRerun membuat run baru dengan menyalin konfigurasi run lama (project, requirements, mode)
func (s *Server) handleRerun(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	orig, err := s.store.GetRun(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	run := &agent.TestRun{
		ID:           uuid.New().String(),
		ProjectPath:  orig.ProjectPath,
		Requirements: orig.Requirements,
		Mode:         orig.Mode,
		State:        agent.StateIdle,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.store.CreateRun(r.Context(), run); err != nil {
		http.Error(w, "store error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"run_id":     run.ID,
		"state":      string(run.State),
		"created_at": run.CreatedAt.Format(time.RFC3339),
	})
}

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
	lastState := ""

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Second):
		}

		run, err := s.store.GetRun(ctx, id)
		if err != nil {
			fmt.Fprintf(w, "event: error\ndata: {\"message\":\"run not found\"}\n\n")
			flusher.Flush()
			return
		}

		currentState := string(run.State)
		if currentState != lastState {
			lastState = currentState
			data, _ := json.Marshal(map[string]string{
				"state":   currentState,
				"message": "State changed to " + currentState,
			})
			fmt.Fprintf(w, "event: state_change\ndata: %s\n\n", data)
			flusher.Flush()

			if run.State == agent.StateDone || run.State == agent.StateFailed {
				doneData, _ := json.Marshal(run)
				fmt.Fprintf(w, "event: done\ndata: %s\n\n", doneData)
				flusher.Flush()
				return
			}
		}
	}
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
