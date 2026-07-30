// Package junit generates JUnit XML reports from test run results.
// JUnit XML is the universal CI/CD format consumed by Jenkins, GitLab CI,
// GitHub Actions, Azure DevOps, CircleCI, and most other CI systems.
package junit

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/agent"
)

// TestSuites is the root element of a JUnit XML report.
type TestSuites struct {
	XMLName    xml.Name    `xml:"testsuites"`
	Tests      int         `xml:"tests,attr"`
	Failures   int         `xml:"failures,attr"`
	Errors     int         `xml:"errors,attr"`
	Time       float64     `xml:"time,attr"`
	Timestamp  string      `xml:"timestamp,attr"`
	Properties []Property  `xml:"properties>property,omitempty"`
	Suites     []TestSuite `xml:"testsuite"`
}

// TestSuite represents a single test suite (maps to one test file).
type TestSuite struct {
	Name      string     `xml:"name,attr"`
	Tests     int        `xml:"tests,attr"`
	Failures  int        `xml:"failures,attr"`
	Errors    int        `xml:"errors,attr"`
	Skipped   int        `xml:"skipped,attr"`
	Time      float64    `xml:"time,attr"`
	Timestamp string     `xml:"timestamp,attr"`
	TestCases []TestCase `xml:"testcase"`
}

// TestCase represents a single test case within a suite.
type TestCase struct {
	Name      string   `xml:"name,attr"`
	ClassName string   `xml:"classname,attr"`
	Time      float64  `xml:"time,attr"`
	Failure   *Failure `xml:"failure,omitempty"`
	Error     *Error   `xml:"error,omitempty"`
	Skipped   *Skipped `xml:"skipped,omitempty"`
	SystemOut string   `xml:"system-out,omitempty"`
}

// Failure represents a test assertion failure.
type Failure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

// Error represents a test execution error.
type Error struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

// Skipped represents a skipped test.
type Skipped struct {
	Message string `xml:"message,attr,omitempty"`
}

// Property is a key-value pair in the report properties section.
type Property struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// Generate creates a JUnit XML report from a test run.
func Generate(run *agent.TestRun) (*TestSuites, error) {
	if run == nil {
		return nil, fmt.Errorf("nil run")
	}

	ts := &TestSuites{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Properties: []Property{
			{Name: "run.id", Value: run.ID},
			{Name: "run.state", Value: string(run.State)},
			{Name: "run.project", Value: run.ProjectPath},
			{Name: "run.mode", Value: run.Mode},
		},
	}

	if run.RunResult == nil {
		// No results — single suite with no test cases
		ts.Suites = []TestSuite{{
			Name:      run.ID,
			Tests:     0,
			Timestamp: ts.Timestamp,
		}}
		return ts, nil
	}

	// Build test cases from failures + passed count
	var cases []TestCase
	failCount := 0

	for _, f := range run.RunResult.Failures {
		failCount++
		cases = append(cases, TestCase{
			Name:      f.Test,
			ClassName: "gotest-agent",
			Time:      0,
			Failure: &Failure{
				Message: f.Message,
				Type:    "AssertionError",
				Content: f.Message,
			},
		})
	}

	// Add passed test cases (synthetic — we don't have individual names)
	for i := 0; i < run.RunResult.Passed; i++ {
		cases = append(cases, TestCase{
			Name:      fmt.Sprintf("test_passed_%d", i+1),
			ClassName: "gotest-agent",
			Time:      0,
		})
	}

	// If run failed but no individual failures recorded, add an error case
	if run.State == agent.StateFailed && failCount == 0 && run.Error != "" {
		cases = append(cases, TestCase{
			Name:      "execution_error",
			ClassName: "gotest-agent",
			Error: &Error{
				Message: run.Error,
				Type:    "ExecutionError",
				Content: run.Error,
			},
		})
	}

	total := run.RunResult.Total
	if total == 0 {
		total = len(cases)
	}

	suite := TestSuite{
		Name:      run.ID,
		Tests:     total,
		Failures:  run.RunResult.Failed,
		Errors:    0,
		Skipped:   0,
		Time:      0,
		Timestamp: ts.Timestamp,
		TestCases: cases,
	}

	if run.State == agent.StateFailed && failCount == 0 && run.Error != "" {
		suite.Errors = 1
	}

	ts.Suites = []TestSuite{suite}
	ts.Tests = total
	ts.Failures = run.RunResult.Failed
	if suite.Errors > 0 {
		ts.Errors = suite.Errors
	}

	return ts, nil
}

// Marshal serializes the TestSuites to JUnit XML bytes.
func Marshal(ts *TestSuites) ([]byte, error) {
	output, err := xml.MarshalIndent(ts, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal junit xml: %w", err)
	}
	return append([]byte(xml.Header), output...), nil
}

// GenerateXML is a convenience function: Generate + Marshal in one call.
func GenerateXML(run *agent.TestRun) ([]byte, error) {
	ts, err := Generate(run)
	if err != nil {
		return nil, err
	}
	return Marshal(ts)
}
