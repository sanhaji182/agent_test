package intelligence

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/gotest-agent/internal/agent"
)

// --- Code Review Assistant ---

// CodeReview represents an automated review of a generated test's code.
type CodeReview struct {
	TestName    string             `json:"test_name"`
	Score       int                `json:"score"` // 0-100
	Passed      bool               `json:"passed"`
	Findings    []ReviewFinding    `json:"findings"`
	Suggestions []ReviewSuggestion `json:"suggestions"`
}

// ReviewFinding is a specific issue found in test code.
type ReviewFinding struct {
	Category    string `json:"category"` // "selector", "assertion", "structure", "naming", "performance"
	Severity    string `json:"severity"` // "error", "warning", "info"
	Description string `json:"description"`
	Line        int    `json:"line,omitempty"`
}

// ReviewSuggestion is a concrete improvement recommendation.
type ReviewSuggestion struct {
	Before string `json:"before,omitempty"`
	After  string `json:"after"`
	Reason string `json:"reason"`
}

// ReviewGeneratedTest analyzes test code for issues and suggests improvements.
func ReviewGeneratedTest(testCode string) *CodeReview {
	review := &CodeReview{
		TestName: extractTestName(testCode),
		Score:    100,
		Passed:   true,
	}

	// Check for fragile selectors
	fragileSelectors := []string{".css-", "class*=", "div > div > div", "nth-child"}
	for _, fragile := range fragileSelectors {
		if strings.Contains(testCode, fragile) {
			review.Findings = append(review.Findings, ReviewFinding{
				Category:    "selector",
				Severity:    "warning",
				Description: fmt.Sprintf("Fragile selector detected: '%s'. Prefer data-testid or accessible selectors.", fragile),
			})
			review.Suggestions = append(review.Suggestions, ReviewSuggestion{
				After:  "Use page.getByTestId('my-element') or page.getByRole('button', {name: 'Submit'})",
				Reason: "Fragile selectors break easily when UI changes",
			})
			review.Score -= 10
		}
	}

	// Check for missing assertions
	if strings.Contains(testCode, "test(") && !strings.Contains(testCode, "expect(") && !strings.Contains(testCode, "assert") {
		review.Findings = append(review.Findings, ReviewFinding{
			Category:    "assertion",
			Severity:    "error",
			Description: "Test has no assertions. Add expect() or assert() calls to verify behavior.",
		})
		review.Score -= 25
	}

	// Check for missing await
	if strings.Contains(testCode, "page.click(") && !strings.Contains(testCode, "await page.click(") {
		review.Findings = append(review.Findings, ReviewFinding{
			Category:    "structure",
			Severity:    "error",
			Description: "Missing 'await' before page action. Playwright operations must be awaited.",
		})
		review.Suggestions = append(review.Suggestions, ReviewSuggestion{
			Before: "page.click(",
			After:  "await page.click(",
			Reason: "Without await, the action may not complete before the next statement",
		})
		review.Score -= 20
	}

	// Check for hardcoded credentials
	credentials := []string{"password", "secret", "token", "apiKey", "api_key"}
	for _, cred := range credentials {
		if strings.Contains(strings.ToLower(testCode), cred+" =") || strings.Contains(strings.ToLower(testCode), cred+":") {
			review.Findings = append(review.Findings, ReviewFinding{
				Category:    "structure",
				Severity:    "error",
				Description: fmt.Sprintf("Potential hardcoded credential: '%s'. Use environment variables or test fixtures.", cred),
			})
			review.Score -= 15
			break // only report once
		}
	}

	// Check for test naming conventions
	if strings.Contains(testCode, "test('") {
		// Extract the name
		idx := strings.Index(testCode, "test('")
		end := strings.Index(testCode[idx+6:], "'")
		if end == -1 {
			end = strings.Index(testCode[idx+6:], `"`)
		}
		if end > 0 {
			testName := testCode[idx+6 : idx+6+end]
			if len(testName) < 10 {
				review.Findings = append(review.Findings, ReviewFinding{
					Category:    "naming",
					Severity:    "info",
					Description: fmt.Sprintf("Test name '%s' is short. Use descriptive names explaining expected behavior.", testName),
				})
				review.Score -= 5
			}
		}
	}

	// Check for unused imports (simple heuristic)
	if strings.Count(testCode, "import {") > 3 {
		review.Findings = append(review.Findings, ReviewFinding{
			Category:    "structure",
			Severity:    "info",
			Description: "Consider reducing imports to only what's used in this test file.",
		})
	}

	// Cap score
	if review.Score < 0 {
		review.Score = 0
	}
	if review.Score < 60 {
		review.Passed = false
	}

	return review
}

// ReviewTestRun reviews all generated test files in a run.
func ReviewTestRun(run *agent.TestRun) []CodeReview {
	var reviews []CodeReview
	for _, tf := range run.TestFiles {
		r := ReviewGeneratedTest(tf.Content)
		r.TestName = tf.Name
		reviews = append(reviews, *r)
	}
	return reviews
}

func extractTestName(code string) string {
	if idx := strings.Index(code, "test('"); idx >= 0 {
		start := idx + 6
		end := strings.IndexAny(code[start:], `'"`)
		if end > 0 {
			return code[start : start+end]
		}
	}
	if idx := strings.Index(code, `test("`); idx >= 0 {
		start := idx + 6
		end := strings.IndexAny(code[start:], `"`)
		if end > 0 {
			return code[start : start+end]
		}
	}
	return "unknown"
}
