package report

import (
	"strings"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/agent"
)

func TestGenerateHTMLString_FullRun(t *testing.T) {
	run := &agent.TestRun{
		ID:    "run-42",
		State: agent.StateDone,
		RunResult: &agent.RunResult{
			Passed: 3, Failed: 1, Total: 4,
			Failures: []agent.Failure{{Test: "login test", Message: "expected 200, got 500"}},
		},
		TestPlan: &agent.TestPlan{
			Summary:   "Cover the login flow",
			Scenarios: []agent.Scenario{{Name: "happy path", Priority: "high"}},
		},
	}

	html, err := GenerateHTMLString(run)
	if err != nil {
		t.Fatalf("GenerateHTMLString: %v", err)
	}
	for _, want := range []string{
		"run-42", "DONE",
		">3</span>", ">1</span>", ">4</span>", // pass/fail/total stats
		"login test", "expected 200, got 500", // failure table
		"Cover the login flow", "happy path", "high", // test plan
		"75%", // pass rate
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("report missing %q", want)
		}
	}
}

func TestGenerateHTMLString_NoResults(t *testing.T) {
	html, err := GenerateHTMLString(&agent.TestRun{ID: "empty", State: agent.StateIdle})
	if err != nil {
		t.Fatalf("GenerateHTMLString: %v", err)
	}
	if !strings.Contains(html, "No results available.") {
		t.Fatal("empty run should render the no-results message")
	}
	if strings.Contains(html, "Failures") || strings.Contains(html, "Test Plan") {
		t.Fatal("empty run must not render failures or plan sections")
	}
}

func TestGenerateHTML_EscapesUserContent(t *testing.T) {
	// Failure messages come from test output / LLM — must be HTML-escaped.
	run := &agent.TestRun{
		ID:    "xss",
		State: agent.StateFailed,
		RunResult: &agent.RunResult{
			Failed: 1, Total: 1,
			Failures: []agent.Failure{{Test: "<script>alert(1)</script>", Message: `<img src=x onerror="steal()">`}},
		},
	}
	html, err := GenerateHTMLString(run)
	if err != nil {
		t.Fatalf("GenerateHTMLString: %v", err)
	}
	if strings.Contains(html, "<script>alert(1)</script>") || strings.Contains(html, "<img src=x") {
		t.Fatal("user-controlled content rendered unescaped — XSS in report")
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Fatal("expected escaped script tag in output")
	}
}
