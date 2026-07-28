package reporter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/events"
	"github.com/go-go-golems/gotest-agent/internal/execution"
)

const sampleReport = `{
  "suites": [{
    "title": "login.spec.ts",
    "specs": [{
      "title": "user can log in",
      "tests": [{
        "status": "expected",
        "results": [{
          "status": "passed",
          "steps": [
            {"title": "goto /login", "duration": 120},
            {"title": "fill credentials", "duration": 80}
          ]
        }]
      }]
    }, {
      "title": "wrong password rejected",
      "tests": [{
        "status": "unexpected",
        "results": [{
          "status": "failed",
          "error": {"message": "expected 401"},
          "steps": [
            {"title": "submit", "duration": 50, "error": {"message": "timeout"}}
          ]
        }]
      }]
    }]
  }]
}`

func writeReport(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "report.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write report: %v", err)
	}
	return path
}

func countByType(evs []events.Event, typ events.EventType) int {
	n := 0
	for _, e := range evs {
		if e.Type == typ {
			n++
		}
	}
	return n
}

func TestParseAndEmit_EmitsStepAndAssertionEvents(t *testing.T) {
	ev := events.NewStore()
	ctx := execution.NewContext(ev, nil, nil)

	if err := ParseAndEmit(ctx, "r1", writeReport(t, sampleReport)); err != nil {
		t.Fatalf("ParseAndEmit: %v", err)
	}

	evs := ev.GetEvents("r1")
	if got := countByType(evs, events.TestStarted); got != 2 {
		t.Fatalf("expected 2 test_started, got %d", got)
	}
	if got := countByType(evs, events.StepStarted); got != 3 {
		t.Fatalf("expected 3 step_started, got %d", got)
	}
	if got := countByType(evs, events.StepCompleted); got != 3 {
		t.Fatalf("expected 3 step_completed, got %d", got)
	}
	if got := countByType(evs, events.AssertionPassed); got != 1 {
		t.Fatalf("expected 1 assertion_passed, got %d", got)
	}
	if got := countByType(evs, events.AssertionFailed); got != 1 {
		t.Fatalf("expected 1 assertion_failed, got %d", got)
	}

	// Failed step carries status + error message.
	var failedStep *events.Event
	for i := range evs {
		if evs[i].Type == events.StepCompleted && evs[i].Metadata["status"] == "failed" {
			failedStep = &evs[i]
		}
	}
	if failedStep == nil || failedStep.Message != "FAILED: timeout" {
		t.Fatalf("failed step event wrong: %+v", failedStep)
	}

	// Cumulative timestamps advance: last step of first spec ends at 200ms.
	var lastPassed *events.Event
	for i := range evs {
		if evs[i].Type == events.StepCompleted && evs[i].Metadata["step"] == "fill credentials" {
			lastPassed = &evs[i]
		}
	}
	if lastPassed == nil || lastPassed.Metadata["timestamp_ms"] != "200" {
		t.Fatalf("cumulative timestamp wrong: %+v", lastPassed)
	}
}

func TestParseAndEmit_GracefulOnMissingOrInvalid(t *testing.T) {
	ev := events.NewStore()
	ctx := execution.NewContext(ev, nil, nil)

	// Missing file: nil error, no events (absence of a report is not fatal).
	if err := ParseAndEmit(ctx, "r1", filepath.Join(t.TempDir(), "nope.json")); err != nil {
		t.Fatalf("missing file should be nil error, got %v", err)
	}
	// Invalid JSON: same.
	if err := ParseAndEmit(ctx, "r1", writeReport(t, "not json")); err != nil {
		t.Fatalf("invalid JSON should be nil error, got %v", err)
	}
	if got := ev.GetEvents("r1"); len(got) != 0 {
		t.Fatalf("expected no events, got %d", len(got))
	}
	// Nil context: no panic.
	if err := ParseAndEmit(nil, "r1", "x"); err != nil {
		t.Fatalf("nil ctx should be nil error, got %v", err)
	}
}

func TestItoa(t *testing.T) {
	cases := map[int]string{0: "0", 5: "5", 10: "10", 200: "200", 987654: "987654"}
	for in, want := range cases {
		if got := itoa(in); got != want {
			t.Fatalf("itoa(%d) = %q, want %q", in, got, want)
		}
	}
}
