// Package execution menyediakan ExecutionContext yang menggabungkan semua store
// yang dibutuhkan selama eksekusi test run (events, recordings, visual artifacts).
package execution

import (
	"time"

	"github.com/go-go-golems/gotest-agent/internal/events"
	"github.com/go-go-golems/gotest-agent/internal/recordings"
	"github.com/go-go-golems/gotest-agent/internal/visual"
)

// Context menyatukan semua store yang dibutuhkan selama eksekusi run.
// Digunakan oleh agent dan runner untuk memproduksi data runtime.
type Context struct {
	Events     *events.Store
	Recordings *recordings.Store
	Visuals    *visual.Store
}

// NewContext membuat execution context baru
func NewContext(ev *events.Store, rec *recordings.Store, vis *visual.Store) *Context {
	return &Context{Events: ev, Recordings: rec, Visuals: vis}
}

// EmitEvent mengirim event ke store
func (c *Context) EmitEvent(runID string, eventType events.EventType, phase, message string, metadata map[string]string) {
	if c == nil || c.Events == nil {
		return
	}
	c.Events.Emit(runID, eventType, phase, message, metadata)
}

// RecordScreenshot membuat recording + visual artifact dari screenshot yang diambil
func (c *Context) RecordScreenshot(runID, testName, stepName, screenshotURL string) {
	if c == nil {
		return
	}
	now := time.Now()

	// Buat recording
	if c.Recordings != nil {
		c.Recordings.Add(recordings.Recording{
			RunID:         runID,
			TestName:      testName,
			StepName:      stepName,
			ScreenshotURL: screenshotURL,
			StartTime:     now,
			EndTime:       now,
			Status:        "captured",
		})
	}

	// Buat visual artifact (current = screenshot baru, baseline = URL sebelumnya jika ada)
	if c.Visuals != nil {
		// Cari baseline dari artifact sebelumnya dengan step yang sama
		baseline := ""
		for _, a := range c.Visuals.ByRun(runID) {
			if a.StepName == stepName {
				baseline = a.CurrentURL // Baseline = current terakhir
			}
		}
		c.Visuals.Add(visual.Artifact{
			RunID:       runID,
			StepName:    stepName,
			BaselineURL: baseline,
			CurrentURL:  screenshotURL,
		})
	}

	// Emit event
	if c.Events != nil {
		c.Events.Emit(runID, events.ScreenshotCaptured, "running", "Screenshot: "+stepName, map[string]string{
			"url":       screenshotURL,
			"test_name": testName,
			"step_name": stepName,
		})
	}
}
