// Package intelligence menyediakan risk scoring, recommendations, suite selection,
// release confidence, dan alert rules untuk enterprise QA platform.
package intelligence

import (
	"fmt"
	"sort"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/gitdiff"
	"github.com/go-go-golems/gotest-agent/internal/schedule"
)

// --- Risk Scoring ---

type RiskItem struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"` // "test", "schedule", "project"
	RiskScore   float64 `json:"risk_score"` // 0.0-1.0
	Reason      string  `json:"reason"`
	Environment string  `json:"environment,omitempty"`
}

// ComputeRisk scores tests/schedules by risk using heuristics
func ComputeRisk(runs []*agent.TestRun, schedules []*schedule.Schedule) []RiskItem {
	var items []RiskItem

	// Score tests by failure frequency
	failCounts := map[string]int{}
	totalAppear := map[string]int{}
	for _, r := range runs {
		if r.RunResult == nil {
			continue
		}
		for _, f := range r.RunResult.Failures {
			failCounts[f.Test]++
		}
		totalAppear[r.ID] = 1
	}
	for test, fc := range failCounts {
		score := clamp(float64(fc) / float64(max(len(runs), 1)))
		reason := "high failure frequency"
		if score > 0.5 {
			reason = "critical: fails in >50% of runs"
		}
		items = append(items, RiskItem{Name: test, Type: "test", RiskScore: score, Reason: reason})
	}

	// Score schedules by staleness and environment
	for _, sch := range schedules {
		score := 0.0
		reason := "normal"
		if sch.Environment == "production" {
			score += 0.3
			reason = "production environment"
		} else if sch.Environment == "staging" {
			score += 0.1
		}
		if sch.LastRunAt != nil && time.Since(*sch.LastRunAt) > 48*time.Hour {
			score += 0.3
			reason = "stale: no run in 48h"
		}
		if sch.LastRunStatus == "failed" {
			score += 0.3
			reason = "last run failed"
		}
		items = append(items, RiskItem{Name: sch.Name, Type: "schedule", RiskScore: clamp(score), Reason: reason, Environment: sch.Environment})
	}

	sort.Slice(items, func(i, j int) bool { return items[i].RiskScore > items[j].RiskScore })
	return items
}

// --- Recommendations ---

type Recommendation struct {
	Action   string  `json:"action"`   // "run_now", "investigate", "disable", "prioritize"
	Target   string  `json:"target"`   // schedule/test name
	Reason   string  `json:"reason"`
	Priority float64 `json:"priority"` // 0.0-1.0
}

func GenerateRecommendations(risks []RiskItem) []Recommendation {
	var recs []Recommendation
	for _, r := range risks {
		if r.RiskScore >= 0.7 {
			action := "investigate"
			if r.Type == "schedule" {
				action = "run_now"
			}
			recs = append(recs, Recommendation{Action: action, Target: r.Name, Reason: r.Reason, Priority: r.RiskScore})
		} else if r.RiskScore >= 0.4 {
			recs = append(recs, Recommendation{Action: "prioritize", Target: r.Name, Reason: r.Reason, Priority: r.RiskScore})
		}
	}
	if len(recs) > 10 {
		recs = recs[:10]
	}
	return recs
}

// --- Suite Selection ---

type SelectionMode string

const (
	SelectAll      SelectionMode = "all"
	SelectImpacted SelectionMode = "impacted"
	SelectHighRisk SelectionMode = "high_risk"
	SelectFlaky    SelectionMode = "flaky"
)

type SuiteSelection struct {
	Mode     SelectionMode `json:"mode"`
	Selected []string      `json:"selected"` // test names to run
	Reason   string        `json:"reason"`
}

// SelectSuite chooses which tests to run based on mode and risk data
func SelectSuite(mode SelectionMode, allTests []string, risks []RiskItem, flakyTests []string) *SuiteSelection {
	return SelectSuiteWithPath(mode, allTests, risks, flakyTests, "")
}

// SelectSuiteWithPath is like SelectSuite but accepts a project path for git-based impacted selection
func SelectSuiteWithPath(mode SelectionMode, allTests []string, risks []RiskItem, flakyTests []string, projectPath string) *SuiteSelection {
	switch mode {
	case SelectHighRisk:
		var selected []string
		for _, r := range risks {
			if r.Type == "test" && r.RiskScore >= 0.4 {
				selected = append(selected, r.Name)
			}
		}
		if len(selected) == 0 {
			return &SuiteSelection{Mode: mode, Selected: allTests, Reason: "no high-risk tests found, running all"}
		}
		return &SuiteSelection{Mode: mode, Selected: selected, Reason: "running high-risk tests only"}
	case SelectFlaky:
		if len(flakyTests) == 0 {
			return &SuiteSelection{Mode: mode, Selected: allTests, Reason: "no flaky tests detected, running all"}
		}
		return &SuiteSelection{Mode: mode, Selected: flakyTests, Reason: "running flaky tests for stability check"}
	case SelectImpacted:
		if projectPath != "" {
			changed, err := gitdiff.ChangedFiles(projectPath)
			if err == nil && len(changed) > 0 {
				impacted := gitdiff.MapToTests(changed, allTests)
				return &SuiteSelection{Mode: mode, Selected: impacted, Reason: fmt.Sprintf("impacted by %d changed files", len(changed))}
			}
		}
		// Fallback to high-risk if no git data
		return SelectSuiteWithPath(SelectHighRisk, allTests, risks, flakyTests, "")
	default:
		return &SuiteSelection{Mode: SelectAll, Selected: allTests, Reason: "running full suite"}
	}
}

// --- Release Confidence ---

type Confidence struct {
	Score       float64 `json:"score"`       // 0.0-1.0
	Grade       string  `json:"grade"`       // A, B, C, D, F
	PassRate    float64 `json:"pass_rate"`
	FlakeRate   float64 `json:"flake_rate"`
	RiskScore   float64 `json:"risk_score"`
	Freshness   float64 `json:"freshness"`   // 1.0 = recent, 0.0 = stale
	Explanation string  `json:"explanation"`
}

func ComputeConfidence(runs []*agent.TestRun, risks []RiskItem) *Confidence {
	c := &Confidence{}
	if len(runs) == 0 {
		c.Score = 0
		c.Grade = "F"
		c.Explanation = "No runs available"
		return c
	}

	// Pass rate
	passed := 0
	for _, r := range runs {
		if r.State == agent.StateDone {
			passed++
		}
	}
	c.PassRate = float64(passed) / float64(len(runs))

	// Risk score (avg of top risks)
	if len(risks) > 0 {
		sum := 0.0
		n := min(len(risks), 5)
		for i := 0; i < n; i++ {
			sum += risks[i].RiskScore
		}
		c.RiskScore = sum / float64(n)
	}

	// Freshness (how recent is the latest run)
	latest := runs[0].CreatedAt
	hoursSince := time.Since(latest).Hours()
	c.Freshness = clamp(1.0 - (hoursSince / 72.0)) // Decays over 72h

	// Composite score
	c.Score = clamp(c.PassRate*0.5 + (1.0-c.RiskScore)*0.3 + c.Freshness*0.2)

	// Grade
	switch {
	case c.Score >= 0.9:
		c.Grade = "A"
		c.Explanation = "High confidence: strong pass rate, low risk"
	case c.Score >= 0.75:
		c.Grade = "B"
		c.Explanation = "Good confidence: mostly passing, moderate risk"
	case c.Score >= 0.6:
		c.Grade = "C"
		c.Explanation = "Fair confidence: some failures or elevated risk"
	case c.Score >= 0.4:
		c.Grade = "D"
		c.Explanation = "Low confidence: significant failures or high risk"
	default:
		c.Grade = "F"
		c.Explanation = "Critical: major failures, high risk, or stale data"
	}

	return c
}

// --- Alert Rules ---

type AlertRule struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Condition string  `json:"condition"` // "pass_rate_drop", "failure_spike", "flake_spike", "risk_increase"
	Threshold float64 `json:"threshold"`
	Enabled   bool    `json:"enabled"`
}

type AlertTrigger struct {
	RuleID  string  `json:"rule_id"`
	Rule    string  `json:"rule_name"`
	Value   float64 `json:"value"`
	Message string  `json:"message"`
}

// EvaluateAlertRules checks if any alert conditions are met
func EvaluateAlertRules(rules []AlertRule, passRate float64, failCount int, riskScore float64) []AlertTrigger {
	var triggers []AlertTrigger
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		switch rule.Condition {
		case "pass_rate_drop":
			if passRate < rule.Threshold {
				triggers = append(triggers, AlertTrigger{RuleID: rule.ID, Rule: rule.Name, Value: passRate, Message: "Pass rate dropped below threshold"})
			}
		case "failure_spike":
			if float64(failCount) > rule.Threshold {
				triggers = append(triggers, AlertTrigger{RuleID: rule.ID, Rule: rule.Name, Value: float64(failCount), Message: "Failure count exceeded threshold"})
			}
		case "risk_increase":
			if riskScore > rule.Threshold {
				triggers = append(triggers, AlertTrigger{RuleID: rule.ID, Rule: rule.Name, Value: riskScore, Message: "Risk score exceeded threshold"})
			}
		}
	}
	return triggers
}

// --- Helpers ---

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
