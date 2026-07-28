package execution

import (
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/events"
	"github.com/go-go-golems/gotest-agent/internal/recordings"
	"github.com/go-go-golems/gotest-agent/internal/visual"
)

func TestEmitEvent_NilSafety(t *testing.T) {
	// Both nil receiver and nil Events store must be no-ops, not panics —
	// runners call EmitEvent unconditionally.
	var nilCtx *Context
	nilCtx.EmitEvent("r1", events.TestStarted, "running", "m", nil)

	c := NewContext(nil, nil, nil)
	c.EmitEvent("r1", events.TestStarted, "running", "m", nil)
	c.RecordScreenshot("r1", "t", "s", "url")
}

func TestEmitEvent_ForwardsToStore(t *testing.T) {
	ev := events.NewStore()
	c := NewContext(ev, nil, nil)
	c.EmitEvent("r1", events.TestStarted, "running", "hello", map[string]string{"k": "v"})

	got := ev.GetEvents("r1")
	if len(got) != 1 || got[0].Message != "hello" {
		t.Fatalf("event not forwarded: %+v", got)
	}
}

func TestRecordScreenshot_CreatesAllArtifacts(t *testing.T) {
	ev := events.NewStore()
	rec := recordings.NewStore()
	vis := visual.NewStore()
	c := NewContext(ev, rec, vis)

	c.RecordScreenshot("r1", "login test", "step-1", "http://x/1.png")

	recs := rec.ByRun("r1")
	if len(recs) != 1 || recs[0].Status != "captured" || recs[0].ScreenshotURL != "http://x/1.png" {
		t.Fatalf("recording not created: %+v", recs)
	}
	arts := vis.ByRun("r1")
	if len(arts) != 1 || arts[0].CurrentURL != "http://x/1.png" || arts[0].BaselineURL != "" {
		t.Fatalf("first artifact should have empty baseline: %+v", arts)
	}
	evs := ev.GetEvents("r1")
	if len(evs) != 1 || evs[0].Type != events.ScreenshotCaptured {
		t.Fatalf("screenshot event not emitted: %+v", evs)
	}
}

func TestRecordScreenshot_BaselineChaining(t *testing.T) {
	vis := visual.NewStore()
	c := NewContext(nil, nil, vis)

	// Same step captured twice: second artifact's baseline = first's current.
	c.RecordScreenshot("r1", "t", "step-1", "http://x/1.png")
	c.RecordScreenshot("r1", "t", "step-1", "http://x/2.png")
	// Different step: independent baseline.
	c.RecordScreenshot("r1", "t", "step-2", "http://x/3.png")

	arts := vis.ByRun("r1")
	if len(arts) != 3 {
		t.Fatalf("expected 3 artifacts, got %d", len(arts))
	}
	if arts[1].BaselineURL != "http://x/1.png" || arts[1].CurrentURL != "http://x/2.png" {
		t.Fatalf("baseline chaining broken: %+v", arts[1])
	}
	if arts[2].BaselineURL != "" {
		t.Fatalf("different step must start with empty baseline: %+v", arts[2])
	}
}
