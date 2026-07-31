package api

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/gotest-agent/internal/webhook"
)

// handleListDrifts returns drift records, filterable by ?repository=, ?type=, ?status=.
func (s *Server) handleListDrifts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list := s.drifts.List(q.Get("repository"), q.Get("type"), q.Get("status"))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// handleUpdateDriftStatus updates a drift's status (pending, fixed, ignored).
func (s *Server) handleUpdateDriftStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	updated, err := s.drifts.UpdateStatus(id, body.Status)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// detectDriftFromPush runs drift detection for a GitHub push event using the
// changed-file lists from its commits.
func (s *Server) detectDriftFromPush(event webhook.PushEvent) {
	var added, modified, removed []string
	for _, c := range event.Commits {
		added = append(added, c.Added...)
		modified = append(modified, c.Modified...)
		removed = append(removed, c.Removed...)
	}
	if len(added)+len(modified)+len(removed) == 0 {
		return
	}

	repoDir := ""
	base := filepath.Clean(webhookCloneDir)
	dir := filepath.Clean(filepath.Join(base, event.Repository.FullName))
	if dir != base && strings.HasPrefix(dir, base+string(filepath.Separator)) {
		repoDir = dir
	}

	s.driftDetector.DetectDrift(event.Repository.FullName, repoDir, added, modified, removed)
}
