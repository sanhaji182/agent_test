package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/notify"
	"github.com/go-go-golems/gotest-agent/internal/release"
)

// --- Release handlers ---

func (s *Server) handleCreateRelease(w http.ResponseWriter, r *http.Request) {
	var rel release.Release
	if err := json.NewDecoder(r.Body).Decode(&rel); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
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
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rel)
}

func (s *Server) handleUpdateRelease(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var patch map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	ok := s.releases.Update(id, func(rel *release.Release) {
		if v, ok := safeString(patch, "status"); ok {
			rel.Status = v
		}
		if v, ok := safeString(patch, "name"); ok {
			rel.Name = v
		}
	})
	if !ok {
		writeJSONError(w, http.StatusNotFound, "not found")
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
		writeJSONError(w, http.StatusNotFound, "not found")
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
