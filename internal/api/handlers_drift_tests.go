package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/gotest-agent/internal/drift"
)

// handleGenerateDriftTest synthesizes a test for a drift via the LLM and stores
// it as a generated test (status: generated). Requires AI planning enabled.
func (s *Server) handleGenerateDriftTest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	d, ok := s.drifts.Get(id)
	if !ok {
		http.Error(w, "drift not found", http.StatusNotFound)
		return
	}

	llm := s.aiClient(r.Context())
	if llm == nil {
		http.Error(w, "AI planning is not enabled (set GOTEST_AI_PLANNING=1)", http.StatusServiceUnavailable)
		return
	}

	gen := drift.NewGenerator(llm, s.driftTests)
	gt, err := gen.GenerateForDrift(context.Background(), *d)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "cannot auto-generate") {
			status = http.StatusUnprocessableEntity
		}
		http.Error(w, err.Error(), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(gt)
}

// handleListDriftTests returns generated tests for a drift.
func (s *Server) handleListDriftTests(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	list := s.driftTests.ByDrift(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// handleUpdateDriftTestStatus updates a generated test's status
// (generated, reviewed, rejected).
func (s *Server) handleUpdateDriftTestStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	updated, err := s.driftTests.UpdateStatus(id, body.Status)
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
