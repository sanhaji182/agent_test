package appmetrics

import (
	"strings"
	"testing"
)

func TestMetrics_CountersIncrement(t *testing.T) {
	m := New()
	m.RunsCreated.Add(3)
	m.RunsCompleted.Add(2)
	m.RunsFailed.Add(1)
	m.ActionsExecuted.Add(10)
	m.ActionsHealed.Add(4)

	out := m.Render()
	for _, want := range []string{
		"gotest_runs_created_total 3",
		"gotest_runs_completed_total 2",
		"gotest_runs_failed_total 1",
		"gotest_actions_executed_total 10",
		"gotest_actions_healed_total 4",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Render() missing %q", want)
		}
	}
}

func TestMetrics_Gauges(t *testing.T) {
	m := New()
	m.SetActiveRuns(5)
	m.RecordRunFinish()

	out := m.Render()
	if !strings.Contains(out, "gotest_active_runs 5") {
		t.Error("missing active_runs gauge")
	}
	if !strings.Contains(out, "gotest_last_run_timestamp_seconds") {
		t.Error("missing last_run gauge")
	}
}

func TestMetrics_HistogramBuckets(t *testing.T) {
	m := New()
	m.ObserveActionDuration(50)   // <= 100
	m.ObserveActionDuration(3000)  // <= 5000
	m.ObserveActionDuration(99999) // only Inf

	out := m.Render()
	if !strings.Contains(out, "# TYPE gotest_action_duration_ms histogram") {
		t.Error("missing histogram type declaration")
	}
	if !strings.Contains(out, `gotest_action_duration_ms_bucket{le="100"} 1`) {
		t.Error("le_100 bucket should have count 1")
	}
	if !strings.Contains(out, `gotest_action_duration_ms_bucket{le="Inf"} 3`) {
		t.Error("le_Inf bucket should have count 3 (all observations)")
	}
}

func TestMetrics_RenderHasHelpAndType(t *testing.T) {
	m := New()
	out := m.Render()
	if !strings.Contains(out, "# HELP gotest_runs_created_total") {
		t.Error("missing HELP line")
	}
	if !strings.Contains(out, "# TYPE gotest_runs_created_total counter") {
		t.Error("missing TYPE counter line")
	}
}
