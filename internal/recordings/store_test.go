package recordings

import "testing"

func TestStore_AddAssignsSequentialIDs(t *testing.T) {
	s := NewStore()
	r1 := s.Add(Recording{RunID: "run-1", StepName: "a"})
	r2 := s.Add(Recording{RunID: "run-1", StepName: "b"})
	if r1.ID != "run-1-rec-1" || r2.ID != "run-1-rec-2" {
		t.Fatalf("expected sequential IDs, got %q %q", r1.ID, r2.ID)
	}
}

func TestStore_AddPreservesExplicitID(t *testing.T) {
	s := NewStore()
	r := s.Add(Recording{ID: "custom-id", RunID: "run-1"})
	if r.ID != "custom-id" {
		t.Fatalf("explicit ID overwritten: %q", r.ID)
	}
}

func TestStore_ByRunFilters(t *testing.T) {
	s := NewStore()
	s.Add(Recording{RunID: "run-1", StepName: "a"})
	s.Add(Recording{RunID: "run-2", StepName: "b"})
	s.Add(Recording{RunID: "run-1", StepName: "c"})

	got := s.ByRun("run-1")
	if len(got) != 2 {
		t.Fatalf("expected 2 for run-1, got %d", len(got))
	}
	if s.ByRun("missing") != nil {
		t.Fatal("expected nil for unknown run")
	}
}

func TestStore_AllReturnsCopy(t *testing.T) {
	s := NewStore()
	s.Add(Recording{RunID: "run-1", Status: "captured"})
	all := s.All()
	if len(all) != 1 {
		t.Fatalf("expected 1, got %d", len(all))
	}
	// Mutating the returned slice must not affect the store's backing array.
	all[0].Status = "hacked"
	fresh := s.All()
	if fresh[0].Status != "captured" {
		t.Fatalf("All leaked shared backing state: %q", fresh[0].Status)
	}
}

func TestItoa(t *testing.T) {
	cases := map[int64]string{0: "0", 1: "1", 9: "9", 10: "10", 42: "42", 100: "100", 123456789: "123456789"}
	for in, want := range cases {
		if got := itoa(in); got != want {
			t.Fatalf("itoa(%d) = %q, want %q", in, got, want)
		}
	}
}
