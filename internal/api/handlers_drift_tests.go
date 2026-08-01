package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/gotest-agent/internal/drift"
	"github.com/go-go-golems/gotest-agent/internal/events"
)

var (
	errDriftNotFound = errors.New("drift not found")
	errAIDisabled    = errors.New("AI planning is not enabled (set GOTEST_AI_PLANNING=1)")
)

// generateDriftTest generates a test for a drift and stores it as a generated
// test (status: generated). It is used by both the synchronous POST
// /drifts/{id}/generate-test endpoint and the GET /drifts/{id}/auto-generate
// convenience endpoint.
func (s *Server) generateDriftTest(ctx context.Context, id string, emitEvent bool) (*drift.GeneratedTest, error) {
	d, ok := s.drifts.Get(id)
	if !ok {
		return nil, errDriftNotFound
	}

	llm := s.aiClientForDrift(ctx)
	if llm == nil {
		return nil, errAIDisabled
	}

	gen := drift.NewGenerator(llm, s.driftTests)
	gt, err := gen.GenerateForDrift(ctx, *d)
	if err != nil {
		return nil, err
	}

	if emitEvent && s.events != nil {
		s.events.Emit(id, events.EventType("drift_test_generated"), "auto_generate", "Generated test for drift", map[string]string{
			"drift_id":   d.ID,
			"test_id":    gt.ID,
			"test_name":  gt.TestName,
			"repository": d.Repository,
			"file_path":  d.FilePath,
		})
	}

	return gt, nil
}

// handleGenerateDriftTest synthesizes a test for a drift via the LLM and stores
// it as a generated test (status: generated). Requires AI planning enabled.
func (s *Server) handleGenerateDriftTest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	gt, err := s.generateDriftTest(r.Context(), id, false)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case err == errDriftNotFound:
			status = http.StatusNotFound
		case err == errAIDisabled:
			status = http.StatusServiceUnavailable
		case strings.Contains(err.Error(), "cannot auto-generate"):
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
