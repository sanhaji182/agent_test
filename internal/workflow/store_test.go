package workflow

import (
	"testing"
	"time"
)

// --- ReviewStore ---

func TestReviewStore_CreateForcesPending(t *testing.T) {
	s := NewReviewStore()
	r := s.Create(&Review{RunID: "run-1", Type: "test_plan", Status: Approved})
	if r.ID == "" || r.CreatedAt.IsZero() {
		t.Fatalf("create defaults not applied: %+v", r)
	}
	// Create always resets status to Pending — approval must go through Approve.
	if r.Status != Pending {
		t.Fatalf("expected status forced to pending, got %s", r.Status)
	}
}

func TestReviewStore_GetAndByRun(t *testing.T) {
	s := NewReviewStore()
	r1 := s.Create(&Review{RunID: "run-1", Type: "test_plan"})
	s.Create(&Review{RunID: "run-2", Type: "test_scripts"})
	s.Create(&Review{RunID: "run-1", Type: "fix_suggestion"})

	got, ok := s.Get(r1.ID)
	if !ok || got.Type != "test_plan" {
		t.Fatalf("Get: %v / %+v", ok, got)
	}
	if _, ok := s.Get("missing"); ok {
		t.Fatal("expected miss for unknown ID")
	}

	byRun := s.ByRun("run-1")
	if len(byRun) != 2 {
		t.Fatalf("expected 2 reviews for run-1, got %d", len(byRun))
	}
	if s.ByRun("missing") != nil {
		t.Fatal("expected nil for unknown run")
	}
}

func TestReviewStore_ApproveAndReject(t *testing.T) {
	s := NewReviewStore()
	r1 := s.Create(&Review{RunID: "run-1"})
	r2 := s.Create(&Review{RunID: "run-1"})
	before := r1.UpdatedAt

	time.Sleep(time.Millisecond)
	if !s.Approve(r1.ID, "alice", "lgtm") {
		t.Fatal("Approve returned false for existing review")
	}
	got, _ := s.Get(r1.ID)
	if got.Status != Approved || got.Reviewer != "alice" || got.Comment != "lgtm" {
		t.Fatalf("approve not applied: %+v", got)
	}
	if !got.UpdatedAt.After(before) {
		t.Fatalf("UpdatedAt not advanced: %v -> %v", before, got.UpdatedAt)
	}

	if !s.Reject(r2.ID, "bob", "needs work") {
		t.Fatal("Reject returned false for existing review")
	}
	got2, _ := s.Get(r2.ID)
	if got2.Status != Rejected || got2.Reviewer != "bob" {
		t.Fatalf("reject not applied: %+v", got2)
	}

	if s.Approve("missing", "x", "") || s.Reject("missing", "x", "") {
		t.Fatal("Approve/Reject should return false for unknown ID")
	}
}

// --- SuiteStore ---

func TestSuiteStore_CreateGetListOrder(t *testing.T) {
	s := NewSuiteStore()
	s1 := s.Create(&Suite{Name: "first"})
	s2 := s.Create(&Suite{Name: "second"})
	if s1.ID == "" || s1.CreatedAt.IsZero() {
		t.Fatalf("create defaults not applied: %+v", s1)
	}

	got, ok := s.Get(s1.ID)
	if !ok || got.Name != "first" {
		t.Fatalf("Get: %v / %+v", ok, got)
	}

	// Newest first.
	list := s.List()
	if len(list) != 2 || list[0].ID != s2.ID || list[1].ID != s1.ID {
		t.Fatalf("expected [second first], got %+v", list)
	}
}

func TestSuiteStore_ByTag(t *testing.T) {
	s := NewSuiteStore()
	s.Create(&Suite{Name: "smoke", Tags: []string{"smoke", "ci"}})
	s.Create(&Suite{Name: "full", Tags: []string{"nightly"}})
	s.Create(&Suite{Name: "quick", Tags: []string{"ci"}})

	ci := s.ByTag("ci")
	if len(ci) != 2 {
		t.Fatalf("expected 2 suites tagged ci, got %d", len(ci))
	}
	if s.ByTag("missing") != nil {
		t.Fatal("expected nil for unknown tag")
	}
}

func TestSuiteStore_Delete(t *testing.T) {
	s := NewSuiteStore()
	s1 := s.Create(&Suite{Name: "doomed"})
	s2 := s.Create(&Suite{Name: "keeper"})

	if !s.Delete(s1.ID) {
		t.Fatal("Delete returned false for existing suite")
	}
	if _, ok := s.Get(s1.ID); ok {
		t.Fatal("suite still present after delete")
	}
	// Order slice must be pruned too.
	list := s.List()
	if len(list) != 1 || list[0].ID != s2.ID {
		t.Fatalf("expected only keeper after delete, got %+v", list)
	}

	if s.Delete("missing") {
		t.Fatal("Delete should return false for unknown ID")
	}
}
