package agent

import (
	"context"
	"time"
)

type APIRunner struct{}

func NewAPIRunner() *APIRunner {
	return &APIRunner{}
}

func (r *APIRunner) Run(ctx context.Context, testFiles []TestFile, projectURL string) (*RunResult, error) {
	// API runner is not yet implemented — no real HTTP assertions are performed.
	// Return zero passed to signal no meaningful result was produced.
	result := &RunResult{
		Passed:   0,
		Failed:   0,
		Total:    0,
		Failures: []Failure{},
	}

	for range testFiles {
		// Placeholder: real implementation would send net/http requests,
		// assert responses, and record redacted logs.
		time.Sleep(500 * time.Millisecond)
		result.Total++
	}

	return result, nil
}
