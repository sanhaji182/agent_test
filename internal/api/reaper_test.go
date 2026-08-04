package api_test

import (
	"context"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/api"
	"github.com/go-go-golems/gotest-agent/internal/config"
	"github.com/go-go-golems/gotest-agent/internal/db"
)

// TestReapStaleRuns_MarksInFlightAsFailed verifies that the startup reaper marks
// runs left in an in-flight state (running/analyzing/etc.) as failed, while leaving
// already-terminal runs (done/failed) untouched.
func TestReapStaleRuns_MarksInFlightAsFailed(t *testing.T) {
	store := db.NewMemoryStore()
	srv := api.NewServer(&config.Config{MaxConcurrentRuns: 2}, store, nil)
	ctx := context.Background()

	seed := func(id string, state agent.State) {
		if err := store.CreateRun(ctx, &agent.TestRun{ID: id, ProjectPath: "/p", State: state}); err != nil {
			t.Fatalf("seed %s: %v", id, err)
		}
	}
	seed("run-running", agent.StateRunning)
	seed("run-analyzing", agent.StateAnalyzing)
	seed("run-done", agent.StateDone)
	seed("run-failed", agent.StateFailed)

	n, err := srv.ReapStaleRuns(ctx)
	if err != nil {
		t.Fatalf("ReapStaleRuns: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 reaped runs, got %d", n)
	}

	assertState := func(id string, want agent.State) {
		t.Helper()
		got, err := store.GetRun(ctx, id)
		if err != nil {
			t.Fatalf("get %s: %v", id, err)
		}
		if got.State != want {
			t.Fatalf("%s: expected state %s, got %s", id, want, got.State)
		}
	}
	assertState("run-running", agent.StateFailed)
	assertState("run-analyzing", agent.StateFailed)
	assertState("run-done", agent.StateDone)     // terminal: untouched
	assertState("run-failed", agent.StateFailed) // terminal: untouched

	// A reaped run must carry a finished timestamp and an explanatory error.
	r, err := store.GetRun(ctx, "run-running")
	if err != nil {
		t.Fatalf("get run-running: %v", err)
	}
	if r.FinishedAt == nil {
		t.Fatal("reaped run should have finished_at set")
	}
	if r.Error == "" {
		t.Fatal("reaped run should have an error message")
	}
}
