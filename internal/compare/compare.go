// Package compare menyediakan layanan perbandingan antara dua test run.
package compare

import "github.com/go-go-golems/gotest-agent/internal/agent"

// Result adalah hasil perbandingan dua run
type Result struct {
	RunA           string   `json:"run_a"`
	RunB           string   `json:"run_b"`
	Summary        string   `json:"summary"`
	TotalDelta     int      `json:"total_delta"`     // B.total - A.total
	PassedDelta    int      `json:"passed_delta"`    // B.passed - A.passed
	FailedDelta    int      `json:"failed_delta"`    // B.failed - A.failed
	NewFailures    []string `json:"new_failures"`    // Test yang gagal di B tapi pass di A
	Recovered      []string `json:"recovered"`       // Test yang pass di B tapi gagal di A
	CommonFailures []string `json:"common_failures"` // Test yang gagal di keduanya
	ScreenshotDiff int      `json:"screenshot_diff"` // Perbedaan jumlah screenshot
}

// Compare membandingkan dua run yang sudah selesai
func Compare(a, b *agent.TestRun) *Result {
	res := &Result{
		RunA: a.ID,
		RunB: b.ID,
	}

	// Hitung delta hasil
	ra := a.RunResult
	rb := b.RunResult
	if ra != nil && rb != nil {
		res.TotalDelta = rb.Total - ra.Total
		res.PassedDelta = rb.Passed - ra.Passed
		res.FailedDelta = rb.Failed - ra.Failed

		// Kumpulkan set failure per run
		failsA := failureSet(ra.Failures)
		failsB := failureSet(rb.Failures)

		for test := range failsB {
			if _, inA := failsA[test]; !inA {
				res.NewFailures = append(res.NewFailures, test)
			} else {
				res.CommonFailures = append(res.CommonFailures, test)
			}
		}
		for test := range failsA {
			if _, inB := failsB[test]; !inB {
				res.Recovered = append(res.Recovered, test)
			}
		}
	} else if ra == nil && rb != nil {
		res.TotalDelta = rb.Total
		res.PassedDelta = rb.Passed
		res.FailedDelta = rb.Failed
	}

	// Screenshot diff
	ssA := len(a.Screenshots)
	ssB := len(b.Screenshots)
	res.ScreenshotDiff = ssB - ssA

	// Summary
	if res.FailedDelta < 0 {
		res.Summary = "Improved: fewer failures in the newer run"
	} else if res.FailedDelta > 0 {
		res.Summary = "Regressed: more failures in the newer run"
	} else if len(res.NewFailures) > 0 || len(res.Recovered) > 0 {
		res.Summary = "Mixed: some tests recovered, some new failures"
	} else {
		res.Summary = "No significant change between runs"
	}

	return res
}

func failureSet(failures []agent.Failure) map[string]struct{} {
	m := make(map[string]struct{}, len(failures))
	for _, f := range failures {
		m[f.Test] = struct{}{}
	}
	return m
}
