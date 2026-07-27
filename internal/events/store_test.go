package events_test

import (
	"testing"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/events"
)

func TestEmitAndGet(t *testing.T) {
	s := events.NewStore()
	s.Emit("run-1", events.RunStarted, "idle", "started", nil)
	s.Emit("run-1", events.AnalysisStarted, "analyzing", "analyzing", map[string]string{"path": "/tmp"})
	s.Emit("run-2", events.RunStarted, "idle", "other run", nil)

	evts := s.GetEvents("run-1")
	if len(evts) != 2 {
		t.Fatalf("expected 2 events for run-1, got %d", len(evts))
	}
	if evts[0].Type != events.RunStarted {
		t.Fatalf("expected run_started, got %s", evts[0].Type)
	}
	if evts[1].Metadata["path"] != "/tmp" {
		t.Fatal("expected metadata path=/tmp")
	}

	evts2 := s.GetEvents("run-2")
	if len(evts2) != 1 {
		t.Fatalf("expected 1 event for run-2, got %d", len(evts2))
	}
}

func TestSubscribe(t *testing.T) {
	s := events.NewStore()
	ch, unsub := s.Subscribe("run-1")
	defer unsub()

	s.Emit("run-1", events.TestStarted, "running", "test", nil)

	select {
	case evt := <-ch:
		if evt.Type != events.TestStarted {
			t.Fatalf("expected test_started, got %s", evt.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for event")
	}
}

func TestSubscribe_Unsubscribe(t *testing.T) {
	s := events.NewStore()
	ch, unsub := s.Subscribe("run-1")
	unsub()

	s.Emit("run-1", events.RunCompleted, "done", "done", nil)

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected channel closed")
		}
	case <-time.After(50 * time.Millisecond):
		// OK - channel closed, no read
	}
}

func TestEnableDB_NoPanicOnNilPool(t *testing.T) {
	s := events.NewStore()
	// GetDBEvents should not panic when dbPool is nil
	events, err := s.GetDBEvents(t.Context(), "nonexistent")
	if err != nil {
		t.Fatalf("expected nil error with nil pool, got %v", err)
	}
	if events != nil {
		t.Fatal("expected nil events with nil db pool")
	}
}

func TestEmitWithDBDisabled(t *testing.T) {
	// Default store with no DB enabled should work normally
	s := events.NewStore()
	s.Emit("run-1", events.RunStarted, "idle", "started", nil)

	evts := s.GetEvents("run-1")
	if len(evts) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evts))
	}
}

func TestNewStore_SubscribersInitialized(t *testing.T) {
	s := events.NewStore()
	// Global subscription should not panic
	ch, unsub := s.SubscribeAll()
	defer unsub()

	s.Emit("any-run", events.RunStarted, "idle", "global test", nil)

	select {
	case evt := <-ch:
		if evt.Type != events.RunStarted {
			t.Fatalf("expected run_started, got %s", evt.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for global event")
	}
}

func TestEmit_CapsPerRunEvents(t *testing.T) {
	s := events.NewStore()
	// Emit MaxEventsPerRun + 100 events — should keep only the last MaxEventsPerRun
	n := events.MaxEventsPerRun + 100
	for i := 0; i < n; i++ {
		s.Emit("run-1", events.StepStarted, "running", "event", nil)
	}
	got := s.GetEvents("run-1")
	if len(got) != events.MaxEventsPerRun {
		t.Fatalf("expected %d events after cap, got %d", events.MaxEventsPerRun, len(got))
	}
	// The first events should have been pruned — the earliest surviving event
	// should be the 101st event (index 100 in zero-based).
	if got[0].ID != "run-1-101" {
		t.Fatalf("expected first event to be run-1-101, got %s", got[0].ID)
	}
	if got[len(got)-1].ID != "run-1-10100" {
		t.Fatalf("expected last event to be run-1-10100, got %s", got[len(got)-1].ID)
	}
}
