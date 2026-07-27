package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/gotest-agent/internal/workflow"
)

// --- Review handlers ---

func (s *Server) handleCreateReview(w http.ResponseWriter, r *http.Request) {
	var rev workflow.Review
	if err := json.NewDecoder(r.Body).Decode(&rev); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
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
		writeJSONError(w, http.StatusNotFound, "not found")
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
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	rev, _ := s.reviews.Get(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rev)
}

func (s *Server) handleRequestChangesReview(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Reviewer string `json:"reviewer"`
		Comment  string `json:"comment"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if !s.reviews.Reject(id, req.Reviewer, "Changes requested: "+req.Comment) {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	rev, _ := s.reviews.Get(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rev)
}

func (s *Server) handleListAllReviews(w http.ResponseWriter, r *http.Request) {
	// Return all reviews across all runs (for review queue)
	allRuns, _ := s.store.ListRuns(r.Context(), 100, 0)
	var all []*workflow.Review
	for _, run := range allRuns {
		revs := s.reviews.ByRun(run.ID)
		all = append(all, revs...)
	}
	if all == nil {
		all = []*workflow.Review{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(all)
}

// --- Suite handlers ---

func (s *Server) handleCreateSuite(w http.ResponseWriter, r *http.Request) {
	var suite workflow.Suite
	if err := json.NewDecoder(r.Body).Decode(&suite); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
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
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suite)
}

func (s *Server) handleDeleteSuite(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !s.suites.Delete(id) {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
