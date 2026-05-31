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
