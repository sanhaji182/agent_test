// Package reporter menyediakan parser untuk Playwright JSON report
// yang menghasilkan step-level events dari hasil eksekusi test.
package reporter

import (
	"encoding/json"
	"os"

	"github.com/go-go-golems/gotest-agent/internal/events"
	"github.com/go-go-golems/gotest-agent/internal/execution"
)

// PlaywrightReport adalah struktur JSON report dari Playwright
type PlaywrightReport struct {
	Suites []Suite `json:"suites"`
}

type Suite struct {
	Title string `json:"title"`
	Specs []Spec `json:"specs"`
}

type Spec struct {
	Title string `json:"title"`
	Tests []Test `json:"tests"`
}

type Test struct {
	Status  string   `json:"status"`
	Results []Result `json:"results"`
}

type Result struct {
	Status string `json:"status"`
	Steps  []Step `json:"steps"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type Step struct {
	Title    string `json:"title"`
	Duration int    `json:"duration"`
	Error    *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// ParseAndEmit membaca Playwright JSON report dan emit step-level events
// Ini dipanggil setelah test selesai dijalankan untuk menghasilkan granular events.
func ParseAndEmit(ctx *execution.Context, runID, reportPath string) error {
	if ctx == nil {
		return nil
	}

	data, err := os.ReadFile(reportPath)
	if err != nil {
		return nil // File tidak ada = tidak ada report, bukan error fatal
	}

	var report PlaywrightReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil
	}

	cumulativeMs := 0 // Track cumulative time across all steps

	for _, suite := range report.Suites {
		for _, spec := range suite.Specs {
			// Emit test_started
			ctx.EmitEvent(runID, events.TestStarted, "running", "Test: "+spec.Title, map[string]string{"test": spec.Title, "suite": suite.Title, "timestamp_ms": itoa(cumulativeMs)})

			for _, test := range spec.Tests {
				for _, result := range test.Results {
					// Emit per-step events with precise timestamps
					for _, step := range result.Steps {
						startMs := cumulativeMs
						ctx.EmitEvent(runID, events.StepStarted, "running", step.Title, map[string]string{
							"test": spec.Title, "step": step.Title,
							"timestamp_ms": itoa(startMs),
						})

						cumulativeMs += step.Duration

						status := "passed"
						msg := "OK"
						if step.Error != nil {
							status = "failed"
							msg = "FAILED: " + step.Error.Message
						}
						ctx.EmitEvent(runID, events.StepCompleted, "running", msg, map[string]string{
							"test": spec.Title, "step": step.Title,
							"status": status, "duration_ms": itoa(step.Duration),
							"timestamp_ms": itoa(cumulativeMs),
						})
					}

					// Emit assertion result
					if result.Status == "passed" || result.Status == "expected" {
						ctx.EmitEvent(runID, events.AssertionPassed, "running", spec.Title+" passed", map[string]string{"test": spec.Title, "timestamp_ms": itoa(cumulativeMs)})
					} else {
						msg := ""
						if result.Error != nil {
							msg = result.Error.Message
						}
						ctx.EmitEvent(runID, events.AssertionFailed, "running", spec.Title+" failed: "+msg, map[string]string{"test": spec.Title, "timestamp_ms": itoa(cumulativeMs)})
					}
				}
			}
		}
	}

	return nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
