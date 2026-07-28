package release

import (
	"testing"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/agent"
)

func TestStore_CreateAppliesDefaults(t *testing.T) {
	s := NewStore()
	r := s.Create(&Release{Name: "v1 rollout", ProjectID: "p1"})
	if r.ID == "" || r.Status != "active" || r.CreatedAt.IsZero() {
		t.Fatalf("create defaults not applied: %+v", r)
	}

	// Explicit status preserved.
	r2 := s.Create(&Release{Name: "hotfix", Status: "completed"})
	if r2.Status != "completed" {
		t.Fatalf("explicit status overwritten: %+v", r2)
	}
}

func TestStore_GetAndList(t *testing.T) {
	s := NewStore()
	r1 := s.Create(&Release{Name: "first"})
	r2 := s.Create(&Release{Name: "second"})

	got, ok := s.Get(r1.ID)
	if !ok || got.Name != "first" {
		t.Fatalf("Get: %v / %+v", ok, got)
	}
	if _, ok := s.Get("missing"); ok {
		t.Fatal("expected miss for unknown ID")
	}

	// Newest first.
	list := s.List()
	if len(list) != 2 || list[0].ID != r2.ID || list[1].ID != r1.ID {
		t.Fatalf("expected [second first], got %+v", list)
	}
}

func TestStore_Update(t *testing.T) {
	s := NewStore()
	r := s.Create(&Release{Name: "rel"})
	before := r.UpdatedAt

	time.Sleep(time.Millisecond) // ensure UpdatedAt visibly advances
	ok := s.Update(r.ID, func(rel *Release) {
		rel.Status = "completed"
		rel.RunIDs = append(rel.RunIDs, "run-1")
	})
	if !ok {
		t.Fatal("Update returned false for existing release")
	}
	got, _ := s.Get(r.ID)
	if got.Status != "completed" || len(got.RunIDs) != 1 {
		t.Fatalf("update not applied: %+v", got)
	}
	if !got.UpdatedAt.After(before) {
		t.Fatalf("UpdatedAt not advanced: %v -> %v", before, got.UpdatedAt)
	}

	if s.Update("missing", func(*Release) {}) {
		t.Fatal("Update should return false for unknown ID")
	}
}

func TestSummarize_Aggregation(t *testing.T) {
	rel := &Release{ID: "rel-1"}
	runs := []*agent.TestRun{
		{State: agent.StateFailed, RunResult: &agent.RunResult{Total: 5, Passed: 3, Failed: 2}},
		{State: agent.StateDone, RunResult: &agent.RunResult{Total: 4, Passed: 4, Failed: 0}},
		{State: agent.StateRunning}, // in-flight: no RunResult, counts toward TotalRuns only
		{State: agent.StateDone, RunResult: &agent.RunResult{Total: 2, Passed: 2, Failed: 0}},
	}

	sum := Summarize(rel, runs)
	if sum.ReleaseID != "rel-1" || sum.TotalRuns != 4 {
		t.Fatalf("basic fields wrong: %+v", sum)
	}
	if sum.PassedRuns != 2 || sum.FailedRuns != 1 {
		t.Fatalf("run counting wrong: %+v", sum)
	}
	if sum.TotalTests != 11 || sum.TotalPassed != 9 || sum.TotalFailed != 2 {
		t.Fatalf("test aggregation wrong: %+v", sum)
	}
	if sum.PassRate != 0.5 { // 2 passed of 4 total runs
		t.Fatalf("expected pass rate 0.5, got %v", sum.PassRate)
	}
	// LatestStatus comes from runs[0] (callers pass newest-first).
	if sum.LatestStatus != string(agent.StateFailed) {
		t.Fatalf("expected latest status failed, got %q", sum.LatestStatus)
	}
}

func TestSummarize_EmptyRuns(t *testing.T) {
	sum := Summarize(&Release{ID: "rel-1"}, nil)
	if sum.TotalRuns != 0 || sum.PassRate != 0 || sum.LatestStatus != "" {
		t.Fatalf("empty runs should yield zero summary, got %+v", sum)
	}
}
