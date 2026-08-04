package report

import (
	"strings"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/agent"
)

func TestGenerateHTMLString_FullRun_Indonesian(t *testing.T) {
	run := &agent.TestRun{
		ID:    "run-42",
		State: agent.StateDone,
		RunResult: &agent.RunResult{
			Passed: 3, Failed: 1, Total: 4,
			Failures: []agent.Failure{{Test: "login test", Message: "expected 200, got 500"}},
		},
		TestPlan: &agent.TestPlan{
			Summary:   "Cover the login flow",
			Scenarios: []agent.Scenario{{Name: "happy path", Priority: "high", Steps: []string{"step 1", "step 2"}}},
		},
	}

	html, err := GenerateHTMLString(run) // default = Indonesian
	if err != nil {
		t.Fatalf("GenerateHTMLString: %v", err)
	}
	for _, want := range []string{
		"run-42", "SELESAI", // Indonesian state badge
		">3</span>", ">1</span>", ">4</span>", // pass/fail/total stats
		"login test", "expected 200, got 500", // failure details
		"Cover the login flow", "happy path", // test plan
		"75%",                 // pass rate
		"Cara Kerja Test Ini", // how it works section
		"Ringkasan Singkat",   // executive summary
		"Saran untuk Anda",    // recommendations
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("Indonesian report missing %q", want)
		}
	}
}

func TestGenerateHTMLString_FullRun_English(t *testing.T) {
	run := &agent.TestRun{
		ID:    "run-42",
		State: agent.StateDone,
		RunResult: &agent.RunResult{
			Passed: 3, Failed: 1, Total: 4,
			Failures: []agent.Failure{{Test: "login test", Message: "expected 200, got 500"}},
		},
		TestPlan: &agent.TestPlan{
			Summary:   "Cover the login flow",
			Scenarios: []agent.Scenario{{Name: "happy path", Priority: "high", Steps: []string{"step 1", "step 2"}}},
		},
	}

	html, err := GenerateHTMLStringLang(run, "en")
	if err != nil {
		t.Fatalf("GenerateHTMLStringLang(en): %v", err)
	}
	for _, want := range []string{
		"run-42", "COMPLETED", // English state badge
		">3</span>", ">1</span>", ">4</span>", // pass/fail/total stats
		"login test", "expected 200, got 500", // failure details
		"Cover the login flow", "happy path", // test plan
		"75%",                 // pass rate
		"How This Test Works", // how it works section
		"Quick Summary",       // executive summary
		"Suggestions for You", // recommendations
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("English report missing %q", want)
		}
	}
}

func TestGenerateHTMLString_NoResults_Indonesian(t *testing.T) {
	html, err := GenerateHTMLString(&agent.TestRun{ID: "empty", State: agent.StateIdle})
	if err != nil {
		t.Fatalf("GenerateHTMLString: %v", err)
	}
	if !strings.Contains(html, "Belum ada hasil") {
		t.Fatal("empty run should render the Indonesian no-results message")
	}
	if strings.Contains(html, "Yang Perlu Diperbaiki") || strings.Contains(html, "Daftar yang Dicek") {
		t.Fatal("empty run must not render failures or plan sections")
	}
}

func TestGenerateHTMLString_NoResults_English(t *testing.T) {
	html, err := GenerateHTMLStringLang(&agent.TestRun{ID: "empty", State: agent.StateIdle}, "en")
	if err != nil {
		t.Fatalf("GenerateHTMLStringLang(en): %v", err)
	}
	if !strings.Contains(html, "No results to show yet") {
		t.Fatal("empty run should render the English no-results message")
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
