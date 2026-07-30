package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/gotest-agent/internal/parser/types"
)

// TestPlan represents a comprehensive test plan generated from codebase analysis
type TestPlan struct {
	Tests []TestCase `json:"test_plan"`
}

// TestCase represents a single test case in the plan
type TestCase struct {
	Name          string `json:"name"`
	Type          string `json:"type"` // unit, integration, e2e
	Description   string `json:"description"`
	Priority      string `json:"priority"` // high, medium, low
	EstimatedTime string `json:"estimated_time"`
	Confidence    int    `json:"confidence"` // 0-100
}

// SynthesisService combines parser output and calls LLM to generate test plans
type SynthesisService struct {
	client Client
}

// NewSynthesisService creates a new synthesis service
func NewSynthesisService(client Client) *SynthesisService {
	return &SynthesisService{client: client}
}

// GenerateTestPlan generates a comprehensive test plan from parsed codebase
func (s *SynthesisService) GenerateTestPlan(ctx context.Context, codebase *types.Codebase) (*TestPlan, error) {
	// Build prompt with codebase information
	prompt := s.buildPrompt(codebase)

	// Call LLM API
	response, err := s.client.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM API: %w", err)
	}

	// Parse JSON response
	testPlan, err := s.parseResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Score confidence for each test case
	scorer := &ConfidenceScorer{}
	for i := range testPlan.Tests {
		testPlan.Tests[i].Confidence = scorer.ScoreTestCase(&testPlan.Tests[i], codebase)
	}

	return testPlan, nil
}

// buildPrompt constructs the LLM prompt with codebase context
func (s *SynthesisService) buildPrompt(codebase *types.Codebase) string {
	routesJSON, _ := json.MarshalIndent(codebase.Routes, "", "  ")
	modelsJSON, _ := json.MarshalIndent(codebase.Models, "", "  ")
	handlersJSON, _ := json.MarshalIndent(codebase.Handlers, "", "  ")

	return fmt.Sprintf(`You are an expert test automation engineer. Analyze this codebase and generate a comprehensive test plan.

CODEBASE INFORMATION:
Language: %s
Framework: %s

ROUTES (%d):
%s

MODELS (%d):
%s

HANDLERS (%d):
%s

Generate a test plan that covers:
1. Happy path scenarios for each route
2. Edge cases (invalid input, missing data, authorization errors)
3. CRUD operations for each model
4. Integration tests for complex workflows
5. Security tests (authentication, authorization, input validation)

Output format (JSON):
{
  "test_plan": [
    {
      "name": "Test name",
      "type": "unit|integration|e2e",
      "description": "What this test verifies",
      "priority": "high|medium|low",
      "estimated_time": "5m"
    }
  ]
}

Be specific and actionable. Each test should be implementable by a developer.
`, codebase.Language, codebase.Framework, len(codebase.Routes), routesJSON, len(codebase.Models), modelsJSON, len(codebase.Handlers), handlersJSON)
}

// parseResponse extracts TestPlan from LLM response
func (s *SynthesisService) parseResponse(response string) (*TestPlan, error) {
	var testPlan TestPlan
	if err := json.Unmarshal([]byte(response), &testPlan); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}

	if len(testPlan.Tests) == 0 {
		return nil, fmt.Errorf("no test cases generated")
	}

	return &testPlan, nil
}
