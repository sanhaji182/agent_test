package compare_test

import (
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/compare"
)

func TestCompare_NoResults(t *testing.T) {
	a := &agent.TestRun{ID: "a"}
	b := &agent.TestRun{ID: "b"}
	r := compare.Compare(a, b)
	if r.RunA != "a" || r.RunB != "b" {
		t.Fatal("wrong run IDs")
	}
	if r.Summary == "" {
		t.Fatal("expected summary")
	}
}

func TestCompare_Improved(t *testing.T) {
	a := &agent.TestRun{ID: "a", RunResult: &agent.RunResult{
		Passed: 3, Failed: 2, Total: 5,
		Failures: []agent.Failure{{Test: "login"}, {Test: "checkout"}},
	}}
	b := &agent.TestRun{ID: "b", RunResult: &agent.RunResult{
		Passed: 4, Failed: 1, Total: 5,
		Failures: []agent.Failure{{Test: "checkout"}},
	}}
	r := compare.Compare(a, b)
	if r.FailedDelta != -1 {
		t.Fatalf("expected failed delta -1, got %d", r.FailedDelta)
	}
	if len(r.Recovered) != 1 || r.Recovered[0] != "login" {
		t.Fatalf("expected login recovered, got %v", r.Recovered)
	}
	if len(r.CommonFailures) != 1 || r.CommonFailures[0] != "checkout" {
		t.Fatalf("expected checkout common, got %v", r.CommonFailures)
	}
}

func TestCompare_Regressed(t *testing.T) {
	a := &agent.TestRun{ID: "a", RunResult: &agent.RunResult{Passed: 5, Failed: 0, Total: 5}}
	b := &agent.TestRun{ID: "b", RunResult: &agent.RunResult{
		Passed: 3, Failed: 2, Total: 5,
		Failures: []agent.Failure{{Test: "x"}, {Test: "y"}},
	}}
	r := compare.Compare(a, b)
	if r.FailedDelta != 2 {
		t.Fatalf("expected +2, got %d", r.FailedDelta)
	}
	if len(r.NewFailures) != 2 {
		t.Fatalf("expected 2 new failures, got %d", len(r.NewFailures))
	}
}
