package metrics_test

import (
	"testing"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/metrics"
)

func TestComputeHotspots(t *testing.T) {
	runs := []*agent.TestRun{
		{RunResult: &agent.RunResult{Failures: []agent.Failure{{Test: "login"}, {Test: "checkout"}}}},
		{RunResult: &agent.RunResult{Failures: []agent.Failure{{Test: "login"}}}},
		{RunResult: &agent.RunResult{Failures: []agent.Failure{{Test: "login"}, {Test: "signup"}}}},
	}
	hotspots := metrics.ComputeHotspots(runs, 5)
	if len(hotspots) == 0 {
		t.Fatal("expected hotspots")
	}
	if hotspots[0].TestName != "login" {
		t.Fatalf("expected login as top hotspot, got %s", hotspots[0].TestName)
	}
	if hotspots[0].FailCount != 3 {
		t.Fatalf("expected 3 fails, got %d", hotspots[0].FailCount)
	}
}

func TestComputeSummary(t *testing.T) {
	runs := []*agent.TestRun{
		{State: agent.StateDone, RunResult: &agent.RunResult{Total: 5, Passed: 4, Failed: 1}},
		{State: agent.StateFailed, RunResult: &agent.RunResult{Total: 3, Passed: 1, Failed: 2}},
	}
	sum := metrics.ComputeSummary(runs)
	if sum.TotalRuns != 2 {
		t.Fatalf("expected 2 runs, got %d", sum.TotalRuns)
	}
	if sum.TotalTests != 8 {
		t.Fatalf("expected 8 tests, got %d", sum.TotalTests)
	}
	if sum.PassRate != 0.5 {
		t.Fatalf("expected 0.5 pass rate, got %f", sum.PassRate)
	}
}

func TestComputeTrend(t *testing.T) {
	now := time.Now()
	runs := []*agent.TestRun{
		{CreatedAt: now, RunResult: &agent.RunResult{Total: 5, Failed: 1}},
		{CreatedAt: now.Add(-24 * time.Hour), RunResult: &agent.RunResult{Total: 3, Failed: 0}},
	}
	trend := metrics.ComputeTrend(runs)
	if len(trend) == 0 {
		t.Fatal("expected trend points")
	}
}
