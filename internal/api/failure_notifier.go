package api

import (
	"context"
	"log/slog"

	"github.com/go-go-golems/gotest-agent/internal/events"
)

// StartFailureNotifier subscribes to the global event stream and creates a
// failure notification (plus optional webhook delivery) whenever a run fails.
// If the failed run was started by a schedule with NotifyOnFail enabled, the
// schedule's webhook URL is used (AUDIT/TECHNICAL_DEBT UW-6).
//
// Runs until ctx is cancelled. Call from cmd/server, not from tests that
// don't need it, so goroutine lifecycle stays explicit.
func (s *Server) StartFailureNotifier(ctx context.Context) {
	ch, unsubscribe := s.events.SubscribeAll()
	defer unsubscribe()
	slog.Info("failure notifier started")
	for {
		select {
		case <-ctx.Done():
			slog.Info("failure notifier stopped")
			return
		case evt, ok := <-ch:
			if !ok {
				return
			}
			if evt.Type != events.RunFailed {
				continue
			}
			s.notifyRunFailure(evt)
		}
	}
}

// notifyRunFailure records a failure notification for the run, resolving the
// originating schedule (if any) for webhook delivery.
func (s *Server) notifyRunFailure(evt events.Event) {
	scheduleID := ""
	webhookURL := ""
	for _, sch := range s.schedules.List() {
		if sch.LastRunID == evt.RunID {
			scheduleID = sch.ID
			if sch.NotifyOnFail {
				webhookURL = sch.WebhookURL
			}
			break
		}
	}
	s.notifs.TriggerFailure(evt.RunID, scheduleID, webhookURL, evt.Message)
}
