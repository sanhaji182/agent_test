package agent_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/agent"
)

// mockLLM implements agent.LLM for testing
type mockLLM struct{}

func (m *mockLLM) AnalyzeCodebase(_ context.Context, _ string) (string, error) {
	return `{"language":"go","framework":"chi","routes":["/health"]}`, nil
}

func (m *mockLLM) GenerateTestPlan(_ context.Context, _, _ string) (*agent.TestPlan, error) {
	return &agent.TestPlan{
		Summary:   "Test health endpoint",
		Scenarios: []agent.Scenario{{Name: "Health check", Priority: "high", Steps: []string{"GET /health"}}},
	}, nil
}

func (m *mockLLM) GenerateTestScripts(_ context.Context, _ *agent.TestPlan, _ string) ([]agent.TestFile, error) {
	return []agent.TestFile{{Name: "health.spec.ts", Content: "test('health', async () => {})"}}, nil
}

func (m *mockLLM) SuggestFixes(_ context.Context, _ []agent.Failure, files []agent.TestFile) ([]agent.TestFile, error) {
	return files, nil
}

// blockingRunner blocks in Run until the context is cancelled, simulating a
// long-running Playwright execution so cancellation can be exercised mid-flight.
type blockingRunner struct{}

func (b *blockingRunner) Run(ctx context.Context, _ []agent.TestFile, _ string) (*agent.RunResult, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func TestAgentExecute_CancellationStopsRun(t *testing.T) {
	a := agent.New(&mockLLM{}, &blockingRunner{}, 3)
	run := &agent.TestRun{ID: "test-cancel", ProjectPath: "/tmp/p", State: agent.StateIdle}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- a.Execute(ctx, run)
	}()

	// Let the run reach the blocking runner, then cancel mid-flight.
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected cancellation error, got nil")
		}
		if run.State != agent.StateCancelled {
			t.Fatalf("expected state cancelled, got %s", run.State)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Execute did not return after context cancellation — run not actually stopped")
	}
}

func TestAgentExecute_TimeoutFailsRun(t *testing.T) {
	a := agent.New(&mockLLM{}, &blockingRunner{}, 3)
	run := &agent.TestRun{ID: "test-timeout", ProjectPath: "/tmp/p", State: agent.StateIdle}

	// A deadline (not a manual cancel) simulates the whole-run watchdog timing
	// out. The run must end up FAILED (a timeout), not CANCELLED (user action).
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := a.Execute(ctx, run)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if run.State != agent.StateFailed {
		t.Fatalf("expected state failed on timeout, got %s", run.State)
	}
}

// errFixLLM behaves like mockLLM but its SuggestFixes always fails, simulating
// an unavailable/misconfigured LLM (e.g. out of credits) during the fix step.
type errFixLLM struct{ *mockLLM }

func (errFixLLM) SuggestFixes(_ context.Context, _ []agent.Failure, _ []agent.TestFile) ([]agent.TestFile, error) {
	return nil, errors.New("openai-compatible: status 402: insufficient balance")
}

// alwaysFailRunner always reports one failing action so the fix loop triggers.
type alwaysFailRunner struct{}

func (alwaysFailRunner) Run(_ context.Context, _ []agent.TestFile, _ string) (*agent.RunResult, error) {
	return &agent.RunResult{Passed: 1, Failed: 1, Total: 2, Failures: []agent.Failure{{Test: "flaky", Message: "boom"}}}, nil
}

// TestAgentExecute_FixErrorReportsResultsNotFailure guards the behavior that an
// auto-fix failure (LLM unavailable / out of credits / unparseable response) must
// NOT discard the actual test results. The run should finish as done with its real
// pass/fail counts, not be hard-failed with a misleading "fix: ..." error.
func TestAgentExecute_FixErrorReportsResultsNotFailure(t *testing.T) {
	a := agent.New(errFixLLM{&mockLLM{}}, alwaysFailRunner{}, 3)
	run := &agent.TestRun{ID: "test-fix-err", ProjectPath: "/tmp/p", State: agent.StateIdle}

	err := a.Execute(context.Background(), run)
	if err != nil {
		t.Fatalf("fix error must not fail the run, got: %v", err)
	}
	if run.State != agent.StateDone {
		t.Fatalf("expected state done (results reported), got %s", run.State)
	}
	if run.RunResult == nil || run.RunResult.Passed != 1 || run.RunResult.Failed != 1 {
		t.Fatalf("expected results preserved (1 passed / 1 failed), got %+v", run.RunResult)
	}
}

func TestNewFallbackLLM_ProviderSelection(t *testing.T) {
	if llm := agent.NewFallbackLLM(); llm != nil {
		t.Fatal("no providers should yield nil")
	}
	if llm := agent.NewFallbackLLM(agent.ProviderConfig{Provider: "bogus"}); llm != nil {
		t.Fatal("unknown-only providers should yield nil")
	}
	// Unknown primary is skipped; a valid fallback still produces a usable LLM.
	if llm := agent.NewFallbackLLM(
		agent.ProviderConfig{Provider: "bogus"},
		agent.ProviderConfig{Provider: "anthropic", APIKey: "k", Model: "m"},
	); llm == nil {
		t.Fatal("a valid fallback provider should yield a non-nil LLM")
	}
}

// mockRunner implements agent.Runner for testing
type mockRunner struct {
	failFirst bool
	called    int
}

func (m *mockRunner) Run(_ context.Context, _ []agent.TestFile, _ string) (*agent.RunResult, error) {
	m.called++
	if m.failFirst && m.called == 1 {
		return &agent.RunResult{Passed: 0, Failed: 1, Total: 1, Failures: []agent.Failure{{Test: "health", Message: "timeout"}}}, nil
	}
	return &agent.RunResult{Passed: 1, Failed: 0, Total: 1}, nil
}

func TestAgentExecute_AllPass(t *testing.T) {
	llm := &mockLLM{}
	r := &mockRunner{}
	a := agent.New(llm, r, 3)

	run := &agent.TestRun{ID: "test-1", ProjectPath: "/tmp/project", State: agent.StateIdle}
	err := a.Execute(context.Background(), run)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if run.State != agent.StateDone {
		t.Fatalf("expected state done, got %s", run.State)
	}
	if run.RunResult.Passed != 1 {
		t.Fatalf("expected 1 passed, got %d", run.RunResult.Passed)
	}
}

func TestAgentExecute_FixLoop(t *testing.T) {
	llm := &mockLLM{}
	r := &mockRunner{failFirst: true}
	a := agent.New(llm, r, 3)

	run := &agent.TestRun{ID: "test-2", ProjectPath: "/tmp/project", State: agent.StateIdle}
	err := a.Execute(context.Background(), run)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if run.State != agent.StateDone {
		t.Fatalf("expected state done, got %s", run.State)
	}
	if run.FixAttempts != 1 {
		t.Fatalf("expected 1 fix attempt, got %d", run.FixAttempts)
	}
}

func TestStripJSONMarkers(t *testing.T) {
	// Test via AnthropicLLM indirectly - just verify the types compile
	llm := agent.NewAnthropicLLM("test-key", "claude-sonnet-4-5")
	if llm == nil {
		t.Fatal("expected non-nil LLM")
	}
}

func TestAgentExecute_ModeDefaultsToSimple(t *testing.T) {
	a := agent.New(&mockLLM{}, &mockRunner{}, 3)
	run := &agent.TestRun{ID: "test-mode", ProjectPath: "/tmp/p", State: agent.StateIdle}
	a.Execute(context.Background(), run)
	if run.Mode != "simple" {
		t.Fatalf("expected mode simple, got %s", run.Mode)
	}
}

func TestAgentExecute_AdvancedWithoutSidecar_FallsBackToSimple(t *testing.T) {
	// When mode=advanced but no sidecar configured, should still work (falls back to simple)
	a := agent.New(&mockLLM{}, &mockRunner{}, 3)
	run := &agent.TestRun{ID: "test-adv", ProjectPath: "/tmp/p", Mode: "advanced", State: agent.StateIdle}
	err := a.Execute(context.Background(), run)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if run.State != agent.StateDone {
		t.Fatalf("expected done, got %s", run.State)
	}
}

// mockScreenshotter implements agent.ScreenshotCapturer
type mockScreenshotter struct {
	captured []string
}

func (m *mockScreenshotter) Capture(_ context.Context, runID string, label string) (string, error) {
	url := "/screenshots/" + runID + "/" + label + ".png"
	m.captured = append(m.captured, url)
	return url, nil
}

func TestAgentExecute_ScreenshotOnFailure(t *testing.T) {
	ss := &mockScreenshotter{}
	a := agent.NewWithConfig(&mockLLM{}, &mockRunner{failFirst: true}, 3, agent.AgentConfig{
		Screenshotter: ss,
	})
	run := &agent.TestRun{ID: "test-ss", ProjectPath: "/tmp/p", State: agent.StateIdle}
	err := a.Execute(context.Background(), run)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ss.captured) == 0 {
		t.Fatal("expected screenshots to be captured on failure")
	}
	if len(run.Screenshots) == 0 {
		t.Fatal("expected screenshots in run")
	}
}

func (m *mockLLM) HealAction(ctx context.Context, action string, domSnapshot string, errorMsg string) (string, error) {
	return "", nil
}

func (m *mockLLM) HealActionWithVision(ctx context.Context, action string, domSnapshot string, errorMsg string, imageBase64 string) (string, error) {
	return "", nil
}
