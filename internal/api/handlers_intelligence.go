package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/intelligence"
)

func (s *Server) handleReleaseConfidence(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	rel, ok := s.releases.Get(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "not found")
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
		writeJSONError(w, http.StatusNotFound, "not found")
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

func (s *Server) handleReleaseExplanation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	rel, ok := s.releases.Get(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "not found")
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

	// Build explanation factors
	factors := []map[string]interface{}{}
	factors = append(factors, map[string]interface{}{"factor": "pass_rate", "value": conf.PassRate, "impact": "positive", "detail": fmt.Sprintf("%.0f%% of runs passed", conf.PassRate*100)})
	if conf.RiskScore > 0.5 {
		factors = append(factors, map[string]interface{}{"factor": "risk_score", "value": conf.RiskScore, "impact": "negative", "detail": "High risk tests detected"})
	}
	if conf.Freshness < 0.5 {
		factors = append(factors, map[string]interface{}{"factor": "freshness", "value": conf.Freshness, "impact": "negative", "detail": "Test data is stale (>36h old)"})
	} else {
		factors = append(factors, map[string]interface{}{"factor": "freshness", "value": conf.Freshness, "impact": "positive", "detail": "Recent test data available"})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"confidence": conf,
		"factors":    factors,
	})
}

func (s *Server) handleSuiteSelection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Mode       string   `json:"mode"`
		AllTests   []string `json:"all_tests"`
		FlakyTests []string `json:"flaky_tests"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	runs := s.getAllRuns(r)
	risks := intelligence.ComputeRisk(runs, nil)
	sel := intelligence.SelectSuite(intelligence.SelectionMode(req.Mode), req.AllTests, risks, req.FlakyTests)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sel)
}
