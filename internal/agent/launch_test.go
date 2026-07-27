package agent_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/events"
	"github.com/go-go-golems/gotest-agent/internal/execution"
	"github.com/go-go-golems/gotest-agent/internal/recordings"
	"github.com/go-go-golems/gotest-agent/internal/visual"
)

// recordingStore implements agent.RunPersistence and records every saved state.
type recordingStore struct {
	mu     sync.Mutex
	states []agent.State
	saves  int
	done   chan struct{}
	once   sync.Once
}

func newRecordingStore() *recordingStore {
	return &recordingStore{done: make(chan struct{})}
}

func (s *recordingStore) CreateRun(_ context.Context, run *agent.TestRun) error {
	return s.UpdateRun(context.Background(), run)
}

func (s *recordingStore) UpdateRun(_ context.Context, run *agent.TestRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.saves++
	if len(s.states) == 0 || s.states[len(s.states)-1] != run.State {
		s.states = append(s.states, run.State)
	}
	if run.State == agent.StateDone || run.State == agent.StateFailed {
		s.once.Do(func() { close(s.done) })
	}
	return nil
}

func (s *recordingStore) snapshot() []agent.State {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]agent.State, len(s.states))
	copy(out, s.states)
	return out
}

func (s *recordingStore) waitDone(t *testing.T) {
	t.Helper()
	select {
	case <-s.done:
		// A terminal state was saved; give Launch's final save a moment to land.
		time.Sleep(50 * time.Millisecond)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for run to reach a terminal state")
	}
}

// panickyRunner panics during Run to exercise Launch's recovery path.
type panickyRunner struct{}

func (p *panickyRunner) Run(_ context.Context, _ []agent.TestFile, _ string) (*agent.RunResult, error) {
	panic("boom: simulated runner crash")
}

// TestLaunch_FullLifecyclePersistsTransitions runs the canonical async path
// (Agent.Launch, ADR-001) end-to-end with mocks and verifies:
//  1. every state transition is persisted through RunPersistence
//  2. the run terminates in StateDone with FinishedAt set
//  3. lifecycle events are emitted
func TestLaunch_FullLifecyclePersistsTransitions(t *testing.T) {
	store := newRecordingStore()
	evStore := events.NewStore()
	execCtx := execution.NewContext(evStore, recordings.NewStore(), visual.NewStore())

	a := agent.NewWithConfig(&mockLLM{}, &mockRunner{}, 3, agent.AgentConfig{
		Exec:  execCtx,
		Store: store,
	})

	run := &agent.TestRun{ID: "launch-1", ProjectPath: "/tmp/p", State: agent.StateIdle}
	a.Launch(run)
	store.waitDone(t)

	if run.State != agent.StateDone {
		t.Fatalf("expected done, got %s", run.State)
	}
	if run.FinishedAt == nil {
		t.Fatal("expected FinishedAt to be set")
	}

	// Persisted transitions must include the full happy path, in order.
	want := []agent.State{
		agent.StateIdle,
		agent.StateAnalyzing,
		agent.StatePlanGenerated,
		agent.StateWritingTests,
		agent.StateRunning,
		agent.StateDone,
	}
	got := store.snapshot()
	if len(got) < len(want) {
		t.Fatalf("expected at least %d persisted transitions, got %d: %v", len(want), len(got), got)
	}
	wi := 0
	for _, st := range got {
		if wi < len(want) && st == want[wi] {
			wi++
		}
	}
	if wi != len(want) {
		t.Fatalf("persisted transitions missing expected order %v, got %v", want, got)
	}

	// Lifecycle events present.
	evts := evStore.GetEvents("launch-1")
	if len(evts) == 0 {
		t.Fatal("expected lifecycle events")
	}
	last := evts[len(evts)-1]
	if last.Type != events.RunCompleted {
		t.Fatalf("expected final event run_completed, got %s", last.Type)
	}
}

// TestLaunch_FixLoopPersistsFixingState verifies the fixing state is persisted
// when a test fails on first attempt and succeeds after a fix.
func TestLaunch_FixLoopPersistsFixingState(t *testing.T) {
	store := newRecordingStore()
	a := agent.NewWithConfig(&mockLLM{}, &mockRunner{failFirst: true}, 3, agent.AgentConfig{
		Store: store,
	})

	run := &agent.TestRun{ID: "launch-fix", ProjectPath: "/tmp/p", State: agent.StateIdle}
	a.Launch(run)
	store.waitDone(t)

	if run.State != agent.StateDone {
		t.Fatalf("expected done, got %s", run.State)
	}
	if run.FixAttempts != 1 {
		t.Fatalf("expected 1 fix attempt, got %d", run.FixAttempts)
	}
	sawFixing := false
	for _, st := range store.snapshot() {
		if st == agent.StateFixing {
			sawFixing = true
		}
	}
	if !sawFixing {
		t.Fatalf("expected fixing state to be persisted, got %v", store.snapshot())
	}
}

// TestLaunch_PanicRecoveryMarksRunFailed verifies a panicking runner results in
// StateFailed persisted with FinishedAt and a run_failed event — not a crash.
func TestLaunch_PanicRecoveryMarksRunFailed(t *testing.T) {
	store := newRecordingStore()
	evStore := events.NewStore()
	execCtx := execution.NewContext(evStore, recordings.NewStore(), visual.NewStore())

	a := agent.NewWithConfig(&mockLLM{}, &panickyRunner{}, 3, agent.AgentConfig{
		Exec:  execCtx,
		Store: store,
	})

	run := &agent.TestRun{ID: "launch-panic", ProjectPath: "/tmp/p", State: agent.StateIdle}
	a.Launch(run)
	store.waitDone(t)

	if run.State != agent.StateFailed {
		t.Fatalf("expected failed, got %s", run.State)
	}
	if run.FinishedAt == nil {
		t.Fatal("expected FinishedAt to be set on panic")
	}
	if run.Error == "" {
		t.Fatal("expected error message on panic")
	}

	evts := evStore.GetEvents("launch-panic")
	sawFailed := false
	for _, e := range evts {
		if e.Type == events.RunFailed {
			sawFailed = true
		}
	}
	if !sawFailed {
		t.Fatal("expected run_failed event after panic")
	}
}

// TestLaunch_ErrorPathPersistsFailedState verifies an LLM error marks the run
// failed with the terminal state persisted.
func TestLaunch_ErrorPathPersistsFailedState(t *testing.T) {
	store := newRecordingStore()
	a := agent.NewWithConfig(&failingLLM{}, &mockRunner{}, 3, agent.AgentConfig{
		Store: store,
	})

	run := &agent.TestRun{ID: "launch-err", ProjectPath: "/tmp/p", State: agent.StateIdle}
	a.Launch(run)
	store.waitDone(t)

	if run.State != agent.StateFailed {
		t.Fatalf("expected failed, got %s", run.State)
	}
	if run.Error == "" {
		t.Fatal("expected error message")
	}
	got := store.snapshot()
	if got[len(got)-1] != agent.StateFailed {
		t.Fatalf("expected last persisted state failed, got %v", got)
	}
}

// failingLLM errors on AnalyzeCodebase to exercise the fail() path.
type failingLLM struct{ mockLLM }

func (f *failingLLM) AnalyzeCodebase(_ context.Context, _ string) (string, error) {
	return "", context.DeadlineExceeded
}
