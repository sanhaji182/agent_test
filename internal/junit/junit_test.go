package junit

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/agent"
)

func TestGenerate_AllPassed(t *testing.T) {
	run := &agent.TestRun{
		ID:          "run-1",
		ProjectPath: "/tmp/project",
		Mode:        "simple",
		State:       agent.StateDone,
		RunResult: &agent.RunResult{
			Passed: 5,
			Failed: 0,
			Total:  5,
		},
	}

	ts, err := Generate(run)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if ts.Tests != 5 {
		t.Fatalf("expected 5 tests, got %d", ts.Tests)
	}
	if ts.Failures != 0 {
		t.Fatalf("expected 0 failures, got %d", ts.Failures)
	}
	if len(ts.Suites) != 1 {
		t.Fatalf("expected 1 suite, got %d", len(ts.Suites))
	}
	suite := ts.Suites[0]
	if suite.Tests != 5 {
		t.Fatalf("suite tests = %d, want 5", suite.Tests)
	}
	if len(suite.TestCases) != 5 {
		t.Fatalf("expected 5 test cases, got %d", len(suite.TestCases))
	}
	for _, tc := range suite.TestCases {
		if tc.Failure != nil {
			t.Fatalf("unexpected failure in passed test: %s", tc.Name)
		}
	}
}

func TestGenerate_WithFailures(t *testing.T) {
	run := &agent.TestRun{
		ID:    "run-2",
		State: agent.StateFailed,
		RunResult: &agent.RunResult{
			Passed: 3,
			Failed: 2,
			Total:  5,
			Failures: []agent.Failure{
				{Test: "login test", Message: "expected 200, got 500"},
				{Test: "signup test", Message: "timeout waiting for element"},
			},
		},
	}

	ts, err := Generate(run)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if ts.Failures != 2 {
		t.Fatalf("expected 2 failures, got %d", ts.Failures)
	}

	suite := ts.Suites[0]
	failCount := 0
	for _, tc := range suite.TestCases {
		if tc.Failure != nil {
			failCount++
			if tc.Failure.Message == "" {
				t.Error("failure message should not be empty")
			}
		}
	}
	if failCount != 2 {
		t.Fatalf("expected 2 failure cases, got %d", failCount)
	}
}

func TestGenerate_ExecutionError(t *testing.T) {
	run := &agent.TestRun{
		ID:    "run-3",
		State: agent.StateFailed,
		Error: "playwright not available: could not install driver",
		RunResult: &agent.RunResult{
			Passed: 0,
			Failed: 0,
			Total:  0,
		},
	}

	ts, err := Generate(run)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if ts.Errors != 1 {
		t.Fatalf("expected 1 error, got %d", ts.Errors)
	}
	suite := ts.Suites[0]
	if suite.Errors != 1 {
		t.Fatalf("suite errors = %d, want 1", suite.Errors)
	}
	found := false
	for _, tc := range suite.TestCases {
		if tc.Error != nil {
			found = true
			if !strings.Contains(tc.Error.Message, "playwright") {
				t.Errorf("error message should mention playwright, got %q", tc.Error.Message)
			}
		}
	}
	if !found {
		t.Fatal("expected an error test case")
	}
}

func TestGenerate_NilRunResult(t *testing.T) {
	run := &agent.TestRun{
		ID:    "run-4",
		State: agent.StateIdle,
	}

	ts, err := Generate(run)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(ts.Suites) != 1 {
		t.Fatalf("expected 1 suite, got %d", len(ts.Suites))
	}
	if ts.Suites[0].Tests != 0 {
		t.Fatalf("expected 0 tests, got %d", ts.Suites[0].Tests)
	}
}

func TestGenerate_NilRun(t *testing.T) {
	_, err := Generate(nil)
	if err == nil {
		t.Fatal("expected error for nil run")
	}
}

func TestMarshal_ValidXML(t *testing.T) {
	run := &agent.TestRun{
		ID:    "run-5",
		State: agent.StateDone,
		RunResult: &agent.RunResult{
			Passed: 2,
			Failed: 1,
			Total:  3,
			Failures: []agent.Failure{
				{Test: "checkout", Message: "button not found"},
			},
		},
	}

	data, err := GenerateXML(run)
	if err != nil {
		t.Fatalf("GenerateXML: %v", err)
	}

	// Must start with XML declaration
	if !strings.HasPrefix(string(data), "<?xml") {
		t.Fatal("missing XML declaration")
	}

	// Must be valid XML — unmarshal it back
	var parsed TestSuites
	if err := xml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("output is not valid XML: %v", err)
	}
	if parsed.Tests != 3 {
		t.Fatalf("parsed tests = %d, want 3", parsed.Tests)
	}
	if parsed.Failures != 1 {
		t.Fatalf("parsed failures = %d, want 1", parsed.Failures)
	}
}

func TestMarshal_XMLEscaping(t *testing.T) {
	run := &agent.TestRun{
		ID:    "run-xss",
		State: agent.StateFailed,
		RunResult: &agent.RunResult{
			Failed: 1,
			Total:  1,
			Failures: []agent.Failure{
				{Test: "<script>alert(1)</script>", Message: `expected "foo" & 'bar'`},
			},
		},
	}

	data, err := GenerateXML(run)
	if err != nil {
		t.Fatalf("GenerateXML: %v", err)
	}

	// Raw unescaped content must NOT appear
	if strings.Contains(string(data), "<script>alert(1)</script>") {
		t.Fatal("XSS: unescaped script tag in XML output")
	}
	// Must be parseable
	var parsed TestSuites
	if err := xml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("XSS content broke XML parsing: %v", err)
	}
}

func TestGenerate_Properties(t *testing.T) {
	run := &agent.TestRun{
		ID:          "run-props",
		ProjectPath: "https://example.com",
		Mode:        "advanced",
		State:       agent.StateDone,
		RunResult:   &agent.RunResult{Passed: 1, Total: 1},
	}

	ts, err := Generate(run)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	props := map[string]string{}
	for _, p := range ts.Properties {
		props[p.Name] = p.Value
	}
	if props["run.id"] != "run-props" {
		t.Errorf("run.id = %q", props["run.id"])
	}
	if props["run.project"] != "https://example.com" {
		t.Errorf("run.project = %q", props["run.project"])
	}
	if props["run.mode"] != "advanced" {
		t.Errorf("run.mode = %q", props["run.mode"])
	}
}

func TestGenerateXML_RoundTrip(t *testing.T) {
	run := &agent.TestRun{
		ID:    "roundtrip",
		State: agent.StateDone,
		RunResult: &agent.RunResult{
			Passed: 10,
			Failed: 2,
			Total:  12,
			Failures: []agent.Failure{
				{Test: "test_a", Message: "fail a"},
				{Test: "test_b", Message: "fail b"},
			},
		},
	}

	data, err := GenerateXML(run)
	if err != nil {
		t.Fatalf("GenerateXML: %v", err)
	}

	var parsed TestSuites
	if err := xml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if parsed.Tests != 12 {
		t.Fatalf("tests = %d, want 12", parsed.Tests)
	}
	if parsed.Failures != 2 {
		t.Fatalf("failures = %d, want 2", parsed.Failures)
	}
	if len(parsed.Suites[0].TestCases) != 12 {
		t.Fatalf("test cases = %d, want 12", len(parsed.Suites[0].TestCases))
	}
}
