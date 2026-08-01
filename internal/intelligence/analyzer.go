package intelligence

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-go-golems/gotest-agent/internal/agent"
)

// --- Test Quality Analyzer ---

// TestQuality represents a scored analysis of a test case.
type TestQuality struct {
	Name            string   `json:"name"`
	QualityScore    float64  `json:"quality_score"` // 0.0-1.0
	Completeness    float64  `json:"completeness"`  // coverage of failure scenarios
	Reliability     float64  `json:"reliability"`   // inverse of flakiness
	Performance     float64  `json:"performance"`   // based on avg duration vs baseline
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// RedundancyGroup groups tests that appear to test the same thing.
type RedundancyGroup struct {
	Name    string   `json:"name"`    // representative name
	Tests   []string `json:"tests"`   // similar tests
	Reason  string   `json:"reason"`  // why they're redundant
	Savings float64  `json:"savings"` // estimated time saved by deduplication
}

// AnalyzeTestQuality scores all tests in the run history.
func AnalyzeTestQuality(runs []*agent.TestRun) []TestQuality {
	results := map[string]struct {
		passes      int
		fails       int
		runs        int
		durations   []float64
		failReasons []string
	}{}

	for _, run := range runs {
		if run.RunResult == nil {
			continue
		}
		for _, f := range run.RunResult.Failures {
			info := results[f.Test]
			info.fails++
			info.runs++
			info.failReasons = append(info.failReasons, f.Message)
			results[f.Test] = info
		}
		if run.RunResult != nil && run.RunResult.DurationMs > 0 {
			// Assume successful tests contributed to duration
			for _, tf := range run.TestFiles {
				info := results[tf.Name]
				info.passes++
				info.runs++
				info.durations = append(info.durations, float64(run.RunResult.DurationMs)/float64(max(len(run.TestFiles), 1)))
				results[tf.Name] = info
			}
		}
	}

	var qualities []TestQuality
	for name, info := range results {
		q := TestQuality{Name: name}

		// Completeness: ratio of tests that have assertions (heuristic)
		q.Completeness = clamp(float64(info.passes) / float64(max(info.runs, 1)))

		// Reliability: inverse of failure rate
		if info.runs > 0 {
			q.Reliability = 1.0 - float64(info.fails)/float64(info.runs)
		} else {
			q.Reliability = 1.0
		}

		// Performance: compare average duration to a baseline
		if len(info.durations) > 0 {
			avgDuration := 0.0
			for _, d := range info.durations {
				avgDuration += d
			}
			avgDuration /= float64(len(info.durations))
			// Penalize tests slower than 30s
			if avgDuration > 30000 {
				q.Performance = clamp(1.0 - (avgDuration-30000)/60000)
			} else {
				q.Performance = 1.0
			}
		} else {
			q.Performance = 1.0
		}

		// Overall quality score
		q.QualityScore = clamp(q.Completeness*0.3 + q.Reliability*0.4 + q.Performance*0.3)

		// Identify issues
		if q.Reliability < 0.8 {
			q.Issues = append(q.Issues, fmt.Sprintf("Flaky: fails %.0f%% of runs", (1-q.Reliability)*100))
			q.Recommendations = append(q.Recommendations, "Add retry logic or investigate environment dependencies")
		}
		if q.Performance < 0.5 {
			q.Issues = append(q.Issues, "Slow test: exceeds 30s threshold")
			q.Recommendations = append(q.Recommendations, "Optimize test setup or split into smaller tests")
		}
		if q.Completeness < 0.5 {
			q.Issues = append(q.Issues, "Low coverage: missing assertions or edge cases")
			q.Recommendations = append(q.Recommendations, "Add negative test cases and boundary value checks")
		}
		if len(info.failReasons) > 0 {
			uniqueReasons := uniqueStrings(info.failReasons)
			if len(uniqueReasons) > 0 {
				q.Recommendations = append(q.Recommendations, fmt.Sprintf("Fix common failure: %s", truncate(uniqueReasons[0], 80)))
			}
		}

		qualities = append(qualities, q)
	}

	sort.Slice(qualities, func(i, j int) bool {
		return qualities[i].QualityScore > qualities[j].QualityScore
	})

	return qualities
}

// DetectRedundancy finds groups of tests that likely overlap.
func DetectRedundancy(tests []TestQuality) []RedundancyGroup {
	// Simple heuristic: tests with similar names that both have high quality
	// might be redundant.
	groups := []RedundancyGroup{}
	seen := map[string]bool{}

	for i, a := range tests {
		if seen[a.Name] {
			continue
		}
		group := RedundancyGroup{Name: a.Name, Tests: []string{a.Name}}
		for j := i + 1; j < len(tests); j++ {
			b := tests[j]
			if seen[b.Name] {
				continue
			}
			if stringSimilarity(a.Name, b.Name) >= 0.6 && a.QualityScore > 0.5 && b.QualityScore > 0.5 {
				group.Tests = append(group.Tests, b.Name)
				seen[b.Name] = true
			}
		}
		if len(group.Tests) > 1 {
			group.Reason = fmt.Sprintf("%d tests appear to test overlapping functionality", len(group.Tests))
			group.Savings = float64(len(group.Tests)-1) * 5.0 // ~5 min saved per deduped test
			groups = append(groups, group)
		}
		seen[a.Name] = true
	}

	return groups
}

// stringSimilarity is a simple Jaccard-like word overlap measure.
func stringSimilarity(a, b string) float64 {
	wordsA := toWords(a)
	wordsB := toWords(b)
	if len(wordsA) == 0 && len(wordsB) == 0 {
		return 0
	}
	intersection := 0
	set := map[string]bool{}
	for _, w := range wordsA {
		set[w] = true
	}
	for _, w := range wordsB {
		if set[w] {
			intersection++
		}
	}
	union := len(wordsA) + len(wordsB) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

func toWords(s string) []string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' {
			return r
		}
		return ' '
	}, s)
	parts := strings.Fields(s)
	return parts
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func uniqueStrings(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
