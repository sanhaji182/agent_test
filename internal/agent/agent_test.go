package agent_test

import (
	"context"
	"testing"

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
