package agent_test

import (
	"context"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/events"
	"github.com/go-go-golems/gotest-agent/internal/execution"
	"github.com/go-go-golems/gotest-agent/internal/recordings"
	"github.com/go-go-golems/gotest-agent/internal/visual"
)

// TestFullPipeline_ProducesEventsRecordingsArtifacts verifies that a run with
// failures produces events, recordings, and visual artifacts via execution context.
func TestFullPipeline_ProducesEventsRecordingsArtifacts(t *testing.T) {
	evStore := events.NewStore()
	recStore := recordings.NewStore()
	visStore := visual.NewStore()
	execCtx := execution.NewContext(evStore, recStore, visStore)

	ss := &mockScreenshotter{}
	a := agent.NewWithConfig(
		&mockLLM{},
		&mockRunner{failFirst: true},
		3,
		agent.AgentConfig{
			Screenshotter: ss,
			Exec:          execCtx,
		},
	)

	run := &agent.TestRun{ID: "pipeline-test", ProjectPath: "/tmp/p", State: agent.StateIdle}
	err := a.Execute(context.Background(), run)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify events were emitted
	evts := evStore.GetEvents("pipeline-test")
	if len(evts) == 0 {
		t.Fatal("expected events to be emitted")
	}

	// Check specific event types exist
	hasType := func(et events.EventType) bool {
		for _, e := range evts {
			if e.Type == et {
				return true
			}
		}
		return false
	}

	if !hasType(events.RunStarted) {
		t.Error("missing run_started event")
	}
	if !hasType(events.AnalysisStarted) {
		t.Error("missing analysis_started event")
	}
	if !hasType(events.AnalysisCompleted) {
		t.Error("missing analysis_completed event")
	}
	if !hasType(events.PlanGenerated) {
		t.Error("missing plan_generated event")
	}
	if !hasType(events.ScriptGenerated) {
		t.Error("missing script_generated event")
	}
	if !hasType(events.TestStarted) {
		t.Error("missing test_started event")
	}
	if !hasType(events.ScreenshotCaptured) {
		t.Error("missing screenshot_captured event")
	}
	if !hasType(events.FixAttemptStarted) {
		t.Error("missing fix_attempt_started event")
	}
	if !hasType(events.FixAttemptCompleted) {
		t.Error("missing fix_attempt_completed event")
	}
	if !hasType(events.RunCompleted) {
		t.Error("missing run_completed event")
	}

	// Verify recordings were created
	recs := recStore.ByRun("pipeline-test")
	if len(recs) == 0 {
		t.Fatal("expected recordings to be created from screenshots")
	}
	if recs[0].Status != "captured" {
		t.Fatalf("expected status=captured, got %s", recs[0].Status)
	}
	if recs[0].ScreenshotURL == "" {
		t.Fatal("expected screenshot URL in recording")
	}

	// Verify visual artifacts were created
	arts := visStore.ByRun("pipeline-test")
	if len(arts) == 0 {
		t.Fatal("expected visual artifacts to be created from screenshots")
	}
	if arts[0].CurrentURL == "" {
		t.Fatal("expected current URL in visual artifact")
	}
}

// TestFullPipeline_NoExecContext verifies agent works without exec context (backward compat)
func TestFullPipeline_NoExecContext(t *testing.T) {
	a := agent.New(&mockLLM{}, &mockRunner{}, 3)
	run := &agent.TestRun{ID: "no-ctx", ProjectPath: "/tmp/p", State: agent.StateIdle}
	err := a.Execute(context.Background(), run)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if run.State != agent.StateDone {
		t.Fatalf("expected done, got %s", run.State)
	}
}
