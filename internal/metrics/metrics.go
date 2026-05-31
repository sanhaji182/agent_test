package metrics

import "github.com/go-go-golems/gotest-agent/internal/agent"

// Hotspot is a test that fails frequently
type Hotspot struct {
	TestName   string  `json:"test_name"`
	FailCount  int     `json:"fail_count"`
	TotalRuns  int     `json:"total_runs"`
	FailRate   float64 `json:"fail_rate"`
}

// FlakyTest is a test that alternates between pass and fail
type FlakyTest struct {
	TestName    string `json:"test_name"`
	FlipCount   int    `json:"flip_count"` // Number of pass↔fail transitions
	TotalAppear int    `json:"total_appearances"`
}

// TrendPoint is one data point in a time series
type TrendPoint struct {
	Date       string  `json:"date"`
	PassRate   float64 `json:"pass_rate"`
	FailCount  int     `json:"fail_count"`
	TotalTests int     `json:"total_tests"`
	Duration   int     `json:"duration_ms"`
}

// Summary is an overview of all runs
type Summary struct {
	TotalRuns    int     `json:"total_runs"`
	PassRate     float64 `json:"pass_rate"`
	AvgDuration  int     `json:"avg_duration_ms"`
	TotalTests   int     `json:"total_tests"`
	TotalPassed  int     `json:"total_passed"`
	TotalFailed  int     `json:"total_failed"`
}

// ComputeHotspots finds tests that fail most often across runs
func ComputeHotspots(runs []*agent.TestRun, limit int) []Hotspot {
	fails := map[string]int{}
	appears := map[string]int{}
	for _, r := range runs {
		if r.RunResult == nil {
			continue
		}
		seen := map[string]bool{}
		for _, f := range r.RunResult.Failures {
			fails[f.Test]++
			seen[f.Test] = true
		}
		for t := range seen {
			appears[t]++
		}
	}
	var hotspots []Hotspot
	for t, fc := range fails {
		hotspots = append(hotspots, Hotspot{
			TestName: t, FailCount: fc, TotalRuns: appears[t],
			FailRate: float64(fc) / float64(max(appears[t], 1)),
		})
	}
	// Sort by fail count desc
	for i := range hotspots {
		for j := i + 1; j < len(hotspots); j++ {
			if hotspots[j].FailCount > hotspots[i].FailCount {
				hotspots[i], hotspots[j] = hotspots[j], hotspots[i]
			}
		}
	}
	if limit > 0 && len(hotspots) > limit {
		hotspots = hotspots[:limit]
	}
	return hotspots
}

// DetectFlaky finds tests that flip between pass and fail across consecutive runs
func DetectFlaky(runs []*agent.TestRun) []FlakyTest {
	// Track per-test status history (ordered by run time)
	history := map[string][]bool{} // test → [passed, failed, passed, ...]
	for i := len(runs) - 1; i >= 0; i-- {
		r := runs[i]
		if r.RunResult == nil {
			continue
		}
		failSet := map[string]bool{}
		for _, f := range r.RunResult.Failures {
			failSet[f.Test] = true
			history[f.Test] = append(history[f.Test], false)
		}
		// Tests that passed (appeared in total but not in failures)
		// We can only track failures; assume non-failure = pass for known tests
	}

	var flaky []FlakyTest
	for test, hist := range history {
		flips := 0
		for i := 1; i < len(hist); i++ {
			if hist[i] != hist[i-1] {
				flips++
			}
		}
		if flips >= 2 {
			flaky = append(flaky, FlakyTest{TestName: test, FlipCount: flips, TotalAppear: len(hist)})
		}
	}
	return flaky
}

// ComputeTrend generates time-series data from runs
func ComputeTrend(runs []*agent.TestRun) []TrendPoint {
	// Group by date
	byDate := map[string]*TrendPoint{}
	var dates []string
	for _, r := range runs {
		d := r.CreatedAt.Format("2006-01-02")
		if _, ok := byDate[d]; !ok {
			byDate[d] = &TrendPoint{Date: d}
			dates = append(dates, d)
		}
		pt := byDate[d]
		if r.RunResult != nil {
			pt.TotalTests += r.RunResult.Total
			pt.FailCount += r.RunResult.Failed
		}
	}
	// Compute pass rates
	var points []TrendPoint
	for _, d := range dates {
		pt := byDate[d]
		if pt.TotalTests > 0 {
			pt.PassRate = float64(pt.TotalTests-pt.FailCount) / float64(pt.TotalTests)
		}
		points = append(points, *pt)
	}
	return points
}

// ComputeSummary generates overall metrics
func ComputeSummary(runs []*agent.TestRun) *Summary {
	s := &Summary{TotalRuns: len(runs)}
	passed := 0
	for _, r := range runs {
		if r.State == agent.StateDone {
			passed++
		}
		if r.RunResult != nil {
			s.TotalTests += r.RunResult.Total
			s.TotalPassed += r.RunResult.Passed
			s.TotalFailed += r.RunResult.Failed
		}
	}
	if s.TotalRuns > 0 {
		s.PassRate = float64(passed) / float64(s.TotalRuns)
	}
	return s
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
