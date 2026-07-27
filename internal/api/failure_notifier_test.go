package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/api"
	"github.com/go-go-golems/gotest-agent/internal/config"
	"github.com/go-go-golems/gotest-agent/internal/db"
)

// TestFailureNotifier_CreatesNotificationOnRunFailed verifies that emitting a
// run_failed event produces a failure notification.
func TestFailureNotifier_CreatesNotificationOnRunFailed(t *testing.T) {
	srv := api.NewServer(&config.Config{MaxConcurrentRuns: 2}, db.NewMemoryStore(), nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go srv.StartFailureNotifier(ctx)

	// Give the notifier a moment to subscribe before emitting.
	time.Sleep(20 * time.Millisecond)
	srv.Events().Emit("failed-run", "run_failed", "failed", "boom", nil)

	deadline := time.After(2 * time.Second)
	for {
		notifs := srv.Notifications().ByRun("failed-run")
		if len(notifs) == 1 {
			if notifs[0].Type != "failure" {
				t.Fatalf("expected failure notification, got %s", notifs[0].Type)
			}
			if notifs[0].Message != "boom" {
				t.Fatalf("expected message boom, got %q", notifs[0].Message)
			}
			return
		}
		select {
		case <-deadline:
			t.Fatalf("expected 1 notification for failed-run, got %d", len(notifs))
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// TestFailureNotifier_IgnoresNonFailureEvents verifies non-failure events do
// not create notifications.
func TestFailureNotifier_IgnoresNonFailureEvents(t *testing.T) {
	srv := api.NewServer(&config.Config{MaxConcurrentRuns: 2}, db.NewMemoryStore(), nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go srv.StartFailureNotifier(ctx)

	time.Sleep(20 * time.Millisecond)
	srv.Events().Emit("ok-run", "run_completed", "done", "all good", nil)
	time.Sleep(50 * time.Millisecond)

	if n := len(srv.Notifications().ByRun("ok-run")); n != 0 {
		t.Fatalf("expected 0 notifications for completed run, got %d", n)
	}
}
