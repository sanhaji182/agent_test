package db

import (
	"context"
	"testing"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/agent"
)

// These tests guard the snapshot semantics introduced by the 2026-07-27 race
// fixes: MemoryStore must never share mutable state (slices, RunResult,
// FinishedAt) between callers and the stored copy. A regression here silently
// reintroduces data races between HTTP handlers and the Agent.Launch goroutine.

func sampleRun(id string) *agent.TestRun {
	fin := time.Date(2026, 7, 28, 10, 0, 0, 0, time.UTC)
	return &agent.TestRun{
		ID:          id,
		ProjectPath: "/tmp/p",
		State:       agent.StateIdle,
		Screenshots: []string{"a.png"},
		TestFiles:   []agent.TestFile{{Name: "t.json", Content: "[]"}},
		RunResult: &agent.RunResult{
			Passed:   1,
			Failures: []agent.Failure{{Test: "t", Message: "m"}},
		},
		FinishedAt: &fin,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func TestMemoryStore_CRUDLifecycle(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	if err := s.CreateRun(ctx, sampleRun("r1")); err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	got, err := s.GetRun(ctx, "r1")
	if err != nil {
		t.Fatalf("GetRun: %v", err)
	}
	if got.ID != "r1" || got.ProjectPath != "/tmp/p" {
		t.Fatalf("unexpected run: %+v", got)
	}

	got.State = agent.StateDone
	if err := s.UpdateRun(ctx, got); err != nil {
		t.Fatalf("UpdateRun: %v", err)
	}
	got2, _ := s.GetRun(ctx, "r1")
	if got2.State != agent.StateDone {
		t.Fatalf("update not persisted: %s", got2.State)
	}

	if err := s.DeleteRun(ctx, "r1"); err != nil {
		t.Fatalf("DeleteRun: %v", err)
	}
	if _, err := s.GetRun(ctx, "r1"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestMemoryStore_NotFoundPaths(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	if _, err := s.GetRun(ctx, "missing"); err != ErrNotFound {
		t.Fatalf("GetRun missing: expected ErrNotFound, got %v", err)
	}
	if err := s.DeleteRun(ctx, "missing"); err != ErrNotFound {
		t.Fatalf("DeleteRun missing: expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_ListRuns_PaginationAndOrder(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	for _, id := range []string{"r1", "r2", "r3"} {
		if err := s.CreateRun(ctx, sampleRun(id)); err != nil {
			t.Fatalf("CreateRun(%s): %v", id, err)
		}
	}

	// Newest first
	all, err := s.ListRuns(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(all) != 3 || all[0].ID != "r3" || all[2].ID != "r1" {
		t.Fatalf("expected [r3 r2 r1], got %v", []string{all[0].ID, all[1].ID, all[2].ID})
	}

	// Pagination window
	page, err := s.ListRuns(ctx, 1, 1)
	if err != nil {
		t.Fatalf("ListRuns page: %v", err)
	}
	if len(page) != 1 || page[0].ID != "r2" {
		t.Fatalf("expected [r2], got %+v", page)
	}

	// Offset past end
	empty, err := s.ListRuns(ctx, 10, 99)
	if err != nil || empty != nil {
		t.Fatalf("expected nil past end, got %v / %v", empty, err)
	}
}

func TestMemoryStore_GetReturnsIsolatedSnapshot(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	if err := s.CreateRun(ctx, sampleRun("r1")); err != nil {
		t.Fatalf("CreateRun: %v", err)
	}

	got, _ := s.GetRun(ctx, "r1")
	// Mutate everything mutable on the returned copy.
	got.Screenshots[0] = "hacked.png"
	got.Screenshots = append(got.Screenshots, "extra.png")
	got.TestFiles[0].Content = "hacked"
	got.RunResult.Passed = 999
	got.RunResult.Failures[0].Message = "hacked"
	*got.FinishedAt = time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC)

	fresh, _ := s.GetRun(ctx, "r1")
	if fresh.Screenshots[0] != "a.png" || len(fresh.Screenshots) != 1 {
		t.Fatalf("stored Screenshots mutated via returned copy: %v", fresh.Screenshots)
	}
	if fresh.TestFiles[0].Content != "[]" {
		t.Fatalf("stored TestFiles mutated: %+v", fresh.TestFiles)
	}
	if fresh.RunResult.Passed != 1 || fresh.RunResult.Failures[0].Message != "m" {
		t.Fatalf("stored RunResult mutated: %+v", fresh.RunResult)
	}
	if fresh.FinishedAt.Year() != 2026 {
		t.Fatalf("stored FinishedAt mutated: %v", fresh.FinishedAt)
	}
}

func TestMemoryStore_StoreIsolatedFromCallerAfterWrite(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	in := sampleRun("r1")
	if err := s.CreateRun(ctx, in); err != nil {
		t.Fatalf("CreateRun: %v", err)
	}

	// Mutate the caller's object AFTER Create — simulates the Agent.Launch
	// goroutine continuing to write to its run pointer.
	in.Screenshots[0] = "post-create.png"
	in.RunResult.Failed = 42
	in.State = agent.StateFailed

	got, _ := s.GetRun(ctx, "r1")
	if got.Screenshots[0] != "a.png" {
		t.Fatalf("store shares Screenshots with caller: %v", got.Screenshots)
	}
	if got.RunResult.Failed != 0 {
		t.Fatalf("store shares RunResult with caller: %+v", got.RunResult)
	}
	if got.State != agent.StateIdle {
		t.Fatalf("store shares scalar state? got %s", got.State)
	}
}

func TestMemoryStore_ListReturnsIsolatedSnapshots(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	if err := s.CreateRun(ctx, sampleRun("r1")); err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	list, _ := s.ListRuns(ctx, 10, 0)
	list[0].RunResult.Passed = 777
	list[0].Screenshots[0] = "hacked.png"

	fresh, _ := s.GetRun(ctx, "r1")
	if fresh.RunResult.Passed != 1 || fresh.Screenshots[0] != "a.png" {
		t.Fatalf("ListRuns leaked shared state: %+v", fresh)
	}
}

func TestCloneRun_NilSafety(t *testing.T) {
	if cloneRun(nil) != nil {
		t.Fatal("cloneRun(nil) must be nil")
	}
	// Run with all optional fields nil must not panic and must round-trip.
	minimal := &agent.TestRun{ID: "min", State: agent.StateIdle}
	c := cloneRun(minimal)
	if c.ID != "min" || c.Screenshots != nil || c.RunResult != nil || c.FinishedAt != nil {
		t.Fatalf("unexpected clone of minimal run: %+v", c)
	}
}
