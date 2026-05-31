package intelligence_test

import (
	"testing"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/intelligence"
	"github.com/go-go-golems/gotest-agent/internal/schedule"
)

func TestComputeRisk(t *testing.T) {
	runs := []*agent.TestRun{
		{ID: "1", RunResult: &agent.RunResult{Failures: []agent.Failure{{Test: "login"}}}},
		{ID: "2", RunResult: &agent.RunResult{Failures: []agent.Failure{{Test: "login"}, {Test: "checkout"}}}},
	}
	now := time.Now()
	lastRun := now.Add(-72 * time.Hour)
	scheds := []*schedule.Schedule{
		{Name: "prod-nightly", Environment: "production", LastRunAt: &lastRun, LastRunStatus: "failed"},
	}
	risks := intelligence.ComputeRisk(runs, scheds)
	if len(risks) == 0 {
		t.Fatal("expected risk items")
	}
	// login should be highest risk test
	found := false
	for _, r := range risks {
		if r.Name == "login" && r.Type == "test" {
			found = true
			if r.RiskScore == 0 {
				t.Fatal("expected non-zero risk for login")
			}
		}
	}
	if !found {
		t.Fatal("expected login in risk items")
	}
}

func TestGenerateRecommendations(t *testing.T) {
	risks := []intelligence.RiskItem{
		{Name: "critical-test", Type: "test", RiskScore: 0.8, Reason: "high failure"},
		{Name: "stale-schedule", Type: "schedule", RiskScore: 0.75, Reason: "stale"},
		{Name: "low-risk", Type: "test", RiskScore: 0.2, Reason: "ok"},
	}
	recs := intelligence.GenerateRecommendations(risks)
	if len(recs) < 2 {
		t.Fatalf("expected at least 2 recommendations, got %d", len(recs))
	}
	if recs[0].Action != "investigate" {
		t.Fatalf("expected investigate for high-risk test, got %s", recs[0].Action)
	}
	if recs[1].Action != "run_now" {
		t.Fatalf("expected run_now for stale schedule, got %s", recs[1].Action)
	}
}

func TestSelectSuite(t *testing.T) {
	risks := []intelligence.RiskItem{
		{Name: "login", Type: "test", RiskScore: 0.8},
		{Name: "checkout", Type: "test", RiskScore: 0.5},
		{Name: "about", Type: "test", RiskScore: 0.1},
	}
	all := []string{"login", "checkout", "about", "home"}

	sel := intelligence.SelectSuite(intelligence.SelectHighRisk, all, risks, nil)
	if sel.Mode != intelligence.SelectHighRisk {
		t.Fatalf("expected high_risk mode, got %s", sel.Mode)
	}
	if len(sel.Selected) != 2 {
		t.Fatalf("expected 2 high-risk tests, got %d: %v", len(sel.Selected), sel.Selected)
	}

	sel = intelligence.SelectSuite(intelligence.SelectFlaky, all, risks, []string{"checkout"})
	if len(sel.Selected) != 1 || sel.Selected[0] != "checkout" {
		t.Fatalf("expected flaky selection [checkout], got %v", sel.Selected)
	}

	sel = intelligence.SelectSuite(intelligence.SelectAll, all, risks, nil)
	if len(sel.Selected) != 4 {
		t.Fatalf("expected all 4 tests, got %d", len(sel.Selected))
	}
}

func TestComputeConfidence(t *testing.T) {
	runs := []*agent.TestRun{
		{State: agent.StateDone, CreatedAt: time.Now()},
		{State: agent.StateDone, CreatedAt: time.Now()},
		{State: agent.StateFailed, CreatedAt: time.Now()},
	}
	risks := []intelligence.RiskItem{{RiskScore: 0.3}}
	conf := intelligence.ComputeConfidence(runs, risks)
	if conf.Score == 0 {
		t.Fatal("expected non-zero confidence")
	}
	if conf.Grade == "" {
		t.Fatal("expected grade")
	}
	if conf.PassRate == 0 {
		t.Fatal("expected non-zero pass rate")
	}
}

func TestEvaluateAlertRules(t *testing.T) {
	rules := []intelligence.AlertRule{
		{ID: "1", Name: "Low pass rate", Condition: "pass_rate_drop", Threshold: 0.8, Enabled: true},
		{ID: "2", Name: "High failures", Condition: "failure_spike", Threshold: 5, Enabled: true},
		{ID: "3", Name: "Disabled", Condition: "pass_rate_drop", Threshold: 0.5, Enabled: false},
	}
	triggers := intelligence.EvaluateAlertRules(rules, 0.6, 10, 0.5)
	if len(triggers) != 2 {
		t.Fatalf("expected 2 triggers, got %d", len(triggers))
	}
}
