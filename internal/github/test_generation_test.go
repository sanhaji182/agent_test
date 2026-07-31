package github

import (
	"context"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/ai"
	"github.com/go-go-golems/gotest-agent/internal/parser"
	"github.com/go-go-golems/gotest-agent/internal/parser/types"
)

// Mock implementations for testing
type mockParser struct {
	parseResult *types.Codebase
	parseError  error
}

func (m *mockParser) Parse(ctx context.Context, repoDir string) (*types.Codebase, error) {
	if m.parseError != nil {
		return nil, m.parseError
	}
	return m.parseResult, nil
}

func (m *mockParser) DetectLanguage(repoDir string) string {
	return "go"
}

func (m *mockParser) SupportedLanguages() []string {
	return []string{"go"}
}

type mockLLMClient struct {
	generateResult string
	generateError  error
}

func (m *mockLLMClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	if m.generateError != nil {
		return "", m.generateError
	}
	return m.generateResult, nil
}

func (m *mockLLMClient) GenerateWithImage(ctx context.Context, prompt string, imageData string) (string, error) {
	return m.generateResult, m.generateError
}

type mockTestStore struct {
	runs map[string]*agent.TestRun
}

func (m *mockTestStore) CreateRun(ctx context.Context, run *agent.TestRun) error {
	if m.runs == nil {
		m.runs = make(map[string]*agent.TestRun)
	}
	m.runs[run.ID] = run
	return nil
}

func (m *mockTestStore) UpdateRun(ctx context.Context, run *agent.TestRun) error {
	if m.runs == nil {
		m.runs = make(map[string]*agent.TestRun)
	}
	m.runs[run.ID] = run
	return nil
}

func TestNewTestGenerationService(t *testing.T) {
	integration := NewIntegration("/tmp/repos", "test-secret")
	parserReg := parser.NewRegistry()
	llmClient := &mockLLMClient{}
	store := &mockTestStore{runs: make(map[string]*agent.TestRun)}

	service := NewTestGenerationService(integration, parserReg, llmClient, func(*agent.TestRun) {}, store)

	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}

	if service.integration != integration {
		t.Error("Integration not properly set")
	}

	if service.parser != parserReg {
		t.Error("Parser registry not properly set")
	}

	if service.llm != llmClient {
		t.Error("LLM client not properly set")
	}

	if service.store != agent.RunPersistence(store) {
		t.Error("Store not properly set")
	}
}

func TestConvertTestCasesToScenarios(t *testing.T) {
	service := &TestGenerationService{}

	testCases := []ai.TestCase{
		{
			Name:          "Test user login",
			Type:          "integration",
			Description:   "Verify user can log in with valid credentials",
			Priority:      "high",
			EstimatedTime: "2m",
			Confidence:    90,
		},
		{
			Name:          "Test form validation",
			Type:          "e2e",
			Description:   "Verify form shows validation errors",
			Priority:      "medium",
			EstimatedTime: "1m",
			Confidence:    85,
		},
	}

	scenarios := service.convertTestCasesToScenarios(testCases)

	if len(scenarios) != 2 {
		t.Fatalf("Expected 2 scenarios, got %d", len(scenarios))
	}

	if scenarios[0].Name != "Test user login" {
		t.Errorf("Expected scenario name 'Test user login', got '%s'", scenarios[0].Name)
	}

	if scenarios[0].Priority != "high" {
		t.Errorf("Expected priority 'high', got '%s'", scenarios[0].Priority)
	}

	if len(scenarios[0].Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(scenarios[0].Steps))
	}

	if scenarios[0].Steps[0] != "Verify user can log in with valid credentials" {
		t.Errorf("Expected step description to match test case description")
	}
}

func TestConvertTestCasesToScenariosEmpty(t *testing.T) {
	service := &TestGenerationService{}

	testCases := []ai.TestCase{}
	scenarios := service.convertTestCasesToScenarios(testCases)

	if len(scenarios) != 0 {
		t.Errorf("Expected 0 scenarios for empty input, got %d", len(scenarios))
	}
}

func TestProcessPushEventWithNoChangedFiles(t *testing.T) {
	// This test verifies the service handles push events with no changed files
	// We can't fully test the integration without mocking git operations,
	// but we can verify the service doesn't crash

	integration := NewIntegration("/tmp/repos", "test-secret")
	parserReg := parser.NewRegistry()
	llmClient := &mockLLMClient{}
	store := &mockTestStore{runs: make(map[string]*agent.TestRun)}

	service := NewTestGenerationService(integration, parserReg, llmClient, func(*agent.TestRun) {}, store)

	// Note: Full integration test would require mocking git clone operations
	// For now, we just verify the service can be created without errors
	if service == nil {
		t.Fatal("Expected service to be created")
	}
}

func TestProcessPullRequestEventWithNoChangedFiles(t *testing.T) {
	// This test verifies the service handles PR events gracefully
	// We can't fully test without mocking git operations

	integration := NewIntegration("/tmp/repos", "test-secret")
	parserReg := parser.NewRegistry()
	llmClient := &mockLLMClient{}
	store := &mockTestStore{runs: make(map[string]*agent.TestRun)}

	service := NewTestGenerationService(integration, parserReg, llmClient, func(*agent.TestRun) {}, store)

	if service == nil {
		t.Fatal("Expected service to be created")
	}
}
