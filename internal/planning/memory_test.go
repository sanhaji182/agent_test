package planning

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// Coverage for the in-memory Store implementation used in local development
// and the database-unavailable fallback path. Mirrors the deep-copy isolation
// guards in internal/db/memory_test.go: stored values must never share
// mutable slices with callers.

func TestMemoryStore_DraftLifecycle(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	p := &DraftPlan{ProjectID: "proj-1", Cases: []DraftCase{{Title: "case A"}}}
	if err := s.CreateDraft(ctx, p); err != nil {
		t.Fatalf("CreateDraft: %v", err)
	}
	// prepareDraftCreate contract: ID, status default, case IDs, timestamps.
	if p.ID == "" || p.Status != "draft" || p.Cases[0].ID == "" || p.CreatedAt.IsZero() {
		t.Fatalf("create defaults not applied: %+v", p)
	}

	got, err := s.GetDraft(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetDraft: %v", err)
	}
	if got.ProjectID != "proj-1" || len(got.Cases) != 1 {
		t.Fatalf("unexpected draft: %+v", got)
	}

	got.Status = "approved"
	got.Cases = append(got.Cases, DraftCase{Title: "case B"})
	if err := s.UpdateDraft(ctx, got); err != nil {
		t.Fatalf("UpdateDraft: %v", err)
	}
	// UpdateDraft assigns IDs to new cases.
	fresh, _ := s.GetDraft(ctx, p.ID)
	if fresh.Status != "approved" || len(fresh.Cases) != 2 || fresh.Cases[1].ID == "" {
		t.Fatalf("update not applied: %+v", fresh)
	}
}

func TestMemoryStore_DraftNotFound(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	if _, err := s.GetDraft(ctx, "missing"); err != pgx.ErrNoRows {
		t.Fatalf("GetDraft: expected pgx.ErrNoRows, got %v", err)
	}
	if err := s.UpdateDraft(ctx, &DraftPlan{ID: "missing"}); err != pgx.ErrNoRows {
		t.Fatalf("UpdateDraft: expected pgx.ErrNoRows, got %v", err)
	}
}

func TestMemoryStore_TestCaseCRUDAndFiltering(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	early := time.Now().Add(-time.Hour)
	cases := []*TestCase{
		{ProjectID: "p1", Title: "older", CreatedAt: early, UpdatedAt: early},
		{ProjectID: "p1", Title: "newer"},
		{ProjectID: "p2", Title: "other-project"},
	}
	if err := s.CreateTestCases(ctx, cases); err != nil {
		t.Fatalf("CreateTestCases: %v", err)
	}
	// prepareTestCaseCreate contract.
	if cases[0].Type != "ui" || cases[0].Priority != "medium" || cases[0].Version != 1 {
		t.Fatalf("create defaults not applied: %+v", cases[0])
	}

	p1, err := s.ListTestCases(ctx, "p1")
	if err != nil {
		t.Fatalf("ListTestCases: %v", err)
	}
	if len(p1) != 2 || p1[0].Title != "newer" || p1[1].Title != "older" {
		t.Fatalf("expected [newer older] for p1, got %+v", p1)
	}
	all, _ := s.ListTestCases(ctx, "")
	if len(all) != 3 {
		t.Fatalf("empty filter should list all, got %d", len(all))
	}

	tc, err := s.GetTestCase(ctx, cases[1].ID)
	if err != nil {
		t.Fatalf("GetTestCase: %v", err)
	}
	tc.Title = "renamed"
	if err := s.UpdateTestCase(ctx, tc); err != nil {
		t.Fatalf("UpdateTestCase: %v", err)
	}
	fresh, _ := s.GetTestCase(ctx, tc.ID)
	if fresh.Title != "renamed" {
		t.Fatalf("update not persisted: %+v", fresh)
	}

	if _, err := s.GetTestCase(ctx, "missing"); err != pgx.ErrNoRows {
		t.Fatalf("expected pgx.ErrNoRows, got %v", err)
	}
	if err := s.UpdateTestCase(ctx, &TestCase{ID: "missing"}); err != pgx.ErrNoRows {
		t.Fatalf("expected pgx.ErrNoRows, got %v", err)
	}
}

func TestMemoryStore_TestLists(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	l := &TestList{Name: "smoke", ProjectID: "p1"}
	if err := s.CreateTestList(ctx, l); err != nil {
		t.Fatalf("CreateTestList: %v", err)
	}
	// prepareTestListCreate contract: non-nil slices.
	if l.ID == "" || l.Tags == nil || l.TestCaseIDs == nil {
		t.Fatalf("create defaults not applied: %+v", l)
	}

	got, err := s.GetTestList(ctx, l.ID)
	if err != nil || got.Name != "smoke" {
		t.Fatalf("GetTestList: %v / %+v", err, got)
	}
	if _, err := s.GetTestList(ctx, "missing"); err != pgx.ErrNoRows {
		t.Fatalf("expected pgx.ErrNoRows, got %v", err)
	}

	lists, _ := s.ListTestLists(ctx, "p1")
	if len(lists) != 1 {
		t.Fatalf("expected 1 list for p1, got %d", len(lists))
	}
	none, _ := s.ListTestLists(ctx, "p2")
	if len(none) != 0 {
		t.Fatalf("expected 0 lists for p2, got %d", len(none))
	}
}

func TestMemoryStore_ChangeProposalLifecycle(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	p := &ChangeProposal{
		TestCaseID: "tc-1",
		Prompt:     "make it faster",
		Original:   TestCase{ID: "tc-1", Title: "orig", Steps: []string{"s1"}},
		Proposed:   TestCase{ID: "tc-1", Title: "prop", Steps: []string{"s1", "s2"}},
	}
	if err := s.CreateChangeProposal(ctx, p); err != nil {
		t.Fatalf("CreateChangeProposal: %v", err)
	}
	if p.ID == "" || p.Status != "pending" {
		t.Fatalf("create defaults not applied: %+v", p)
	}

	got, err := s.GetChangeProposal(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetChangeProposal: %v", err)
	}
	now := time.Now()
	got.Status = "approved"
	got.ReviewedAt = &now
	if err := s.UpdateChangeProposal(ctx, got); err != nil {
		t.Fatalf("UpdateChangeProposal: %v", err)
	}
	fresh, _ := s.GetChangeProposal(ctx, p.ID)
	if fresh.Status != "approved" || fresh.ReviewedAt == nil {
		t.Fatalf("update not persisted: %+v", fresh)
	}

	byCase, _ := s.ListChangeProposals(ctx, "tc-1")
	if len(byCase) != 1 {
		t.Fatalf("expected 1 proposal for tc-1, got %d", len(byCase))
	}
	none, _ := s.ListChangeProposals(ctx, "tc-other")
	if len(none) != 0 {
		t.Fatalf("expected 0 for other case, got %d", len(none))
	}

	if err := s.UpdateChangeProposal(ctx, &ChangeProposal{ID: "missing"}); err != pgx.ErrNoRows {
		t.Fatalf("expected pgx.ErrNoRows, got %v", err)
	}
}

func TestMemoryStore_CloneIsolation(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	// Draft: mutating returned case slices must not touch the store.
	d := &DraftPlan{ProjectID: "p1", Cases: []DraftCase{{Title: "t", Steps: []string{"s1"}, Tags: []string{"a"}}}}
	s.CreateDraft(ctx, d)
	got, _ := s.GetDraft(ctx, d.ID)
	got.Cases[0].Steps[0] = "hacked"
	got.Cases[0].Tags[0] = "hacked"
	fresh, _ := s.GetDraft(ctx, d.ID)
	if fresh.Cases[0].Steps[0] != "s1" || fresh.Cases[0].Tags[0] != "a" {
		t.Fatalf("draft store shares slices with caller: %+v", fresh.Cases[0])
	}

	// Caller-side mutation after create must not touch the store either.
	d.Cases[0].Steps[0] = "post-create"
	fresh2, _ := s.GetDraft(ctx, d.ID)
	if fresh2.Cases[0].Steps[0] != "s1" {
		t.Fatalf("draft store shares input slices: %+v", fresh2.Cases[0])
	}

	// TestCase isolation.
	tc := &TestCase{ProjectID: "p1", Title: "t", Steps: []string{"s1"}}
	s.CreateTestCases(ctx, []*TestCase{tc})
	gotTC, _ := s.GetTestCase(ctx, tc.ID)
	gotTC.Steps[0] = "hacked"
	freshTC, _ := s.GetTestCase(ctx, tc.ID)
	if freshTC.Steps[0] != "s1" {
		t.Fatalf("test case store shares Steps: %+v", freshTC)
	}

	// TestList isolation.
	l := &TestList{Name: "l", TestCaseIDs: []string{"tc-1"}}
	s.CreateTestList(ctx, l)
	gotL, _ := s.GetTestList(ctx, l.ID)
	gotL.TestCaseIDs[0] = "hacked"
	freshL, _ := s.GetTestList(ctx, l.ID)
	if freshL.TestCaseIDs[0] != "tc-1" {
		t.Fatalf("test list store shares TestCaseIDs: %+v", freshL)
	}

	// Proposal isolation (nested TestCase snapshots + ReviewedAt pointer).
	now := time.Now()
	pr := &ChangeProposal{TestCaseID: "tc-1", Original: TestCase{Steps: []string{"s1"}}, Proposed: TestCase{Steps: []string{"s1"}}, ReviewedAt: &now}
	s.CreateChangeProposal(ctx, pr)
	gotP, _ := s.GetChangeProposal(ctx, pr.ID)
	gotP.Original.Steps[0] = "hacked"
	*gotP.ReviewedAt = time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC)
	freshP, _ := s.GetChangeProposal(ctx, pr.ID)
	if freshP.Original.Steps[0] != "s1" {
		t.Fatalf("proposal store shares Original.Steps: %+v", freshP.Original)
	}
	if freshP.ReviewedAt.Year() == 1999 {
		t.Fatalf("proposal store shares ReviewedAt pointer: %v", freshP.ReviewedAt)
	}
}
