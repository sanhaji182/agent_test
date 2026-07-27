package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/intelligence"
	"github.com/go-go-golems/gotest-agent/internal/metrics"
)

// --- Metrics handlers ---

func (s *Server) handleMetricsSummary(w http.ResponseWriter, r *http.Request) {
	runs, _ := s.store.ListRuns(r.Context(), 1000, 0)
	sum := metrics.ComputeSummary(runs)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sum)
}

func (s *Server) handleMetricsHotspots(w http.ResponseWriter, r *http.Request) {
	runs, _ := s.store.ListRuns(r.Context(), 1000, 0)
	// Need full run data for hotspots
	var fullRuns []*agent.TestRun
	for _, run := range runs {
		if full, err := s.store.GetRun(r.Context(), run.ID); err == nil {
			fullRuns = append(fullRuns, full)
		}
	}
	hotspots := metrics.ComputeHotspots(fullRuns, 10)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hotspots)
}

func (s *Server) handleMetricsFlaky(w http.ResponseWriter, r *http.Request) {
	runs, _ := s.store.ListRuns(r.Context(), 1000, 0)
	var fullRuns []*agent.TestRun
	for _, run := range runs {
		if full, err := s.store.GetRun(r.Context(), run.ID); err == nil {
			fullRuns = append(fullRuns, full)
		}
	}
	flaky := metrics.DetectFlaky(fullRuns)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flaky)
}

func (s *Server) handleMetricsTrend(w http.ResponseWriter, r *http.Request) {
	runs, _ := s.store.ListRuns(r.Context(), 1000, 0)
	var fullRuns []*agent.TestRun
	for _, run := range runs {
		if full, err := s.store.GetRun(r.Context(), run.ID); err == nil {
			fullRuns = append(fullRuns, full)
		}
	}
	trend := metrics.ComputeTrend(fullRuns)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trend)
}

// --- Intelligence handlers ---

func (s *Server) handleMetricsRisk(w http.ResponseWriter, r *http.Request) {
	runs := s.getAllRuns(r)
	scheds := s.schedules.List()
	risks := intelligence.ComputeRisk(runs, scheds)
	if risks == nil {
		risks = []intelligence.RiskItem{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(risks)
}

func (s *Server) handleMetricsRecommendations(w http.ResponseWriter, r *http.Request) {
	runs := s.getAllRuns(r)
	scheds := s.schedules.List()
	risks := intelligence.ComputeRisk(runs, scheds)
	recs := intelligence.GenerateRecommendations(risks)
	if recs == nil {
		recs = []intelligence.Recommendation{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recs)
}

// --- Alert rules handler ---

func (s *Server) handleEvaluateAlertRules(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Rules []intelligence.AlertRule `json:"rules"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	runs := s.getAllRuns(r)
	sum := metrics.ComputeSummary(runs)
	allRuns := s.getAllRuns(r)
	risks := intelligence.ComputeRisk(allRuns, s.schedules.List())
	avgRisk := 0.0
	if len(risks) > 0 {
		for _, ri := range risks[:min(len(risks), 5)] {
			avgRisk += ri.RiskScore
		}
		avgRisk /= float64(min(len(risks), 5))
	}
	triggers := intelligence.EvaluateAlertRules(req.Rules, sum.PassRate, sum.TotalFailed, avgRisk)
	if triggers == nil {
		triggers = []intelligence.AlertTrigger{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(triggers)
}
