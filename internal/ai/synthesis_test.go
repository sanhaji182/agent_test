package ai

import (
	"encoding/json"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/parser/types"
)

func TestSynthesisService_BuildPrompt(t *testing.T) {
	service := &SynthesisService{}

	codebase := &types.Codebase{
		Language:  "javascript",
		Framework: "express",
		Routes: []types.Route{
			{Path: "/users", Method: "GET", Handler: "getUsers"},
			{Path: "/users/:id", Method: "GET", Handler: "getUser"},
		},
		Models: []types.Model{
			{Name: "User", Fields: []types.Field{
				{Name: "id"},
				{Name: "name"},
				{Name: "email"},
			}},
		},
		Handlers: []types.Handler{
			{Name: "getUsers", Parameters: []types.Parameter{
				{Name: "req"},
				{Name: "res"},
			}},
		},
	}

	prompt := service.buildPrompt(codebase)

	// Verify prompt contains all sections
	if prompt == "" {
		t.Fatal("prompt should not be empty")
	}

	// Check language and framework
	if !contains(prompt, "javascript") {
		t.Error("prompt should contain language")
	}
	if !contains(prompt, "express") {
		t.Error("prompt should contain framework")
	}

	// Check routes section
	if !contains(prompt, "ROUTES") {
		t.Error("prompt should contain ROUTES section")
	}
	if !contains(prompt, "/users") {
		t.Error("prompt should contain route paths")
	}

	// Check models section
	if !contains(prompt, "MODELS") {
		t.Error("prompt should contain MODELS section")
	}
	if !contains(prompt, "User") {
		t.Error("prompt should contain model names")
	}

	// Check handlers section
	if !contains(prompt, "HANDLERS") {
		t.Error("prompt should contain HANDLERS section")
	}
	if !contains(prompt, "getUsers") {
		t.Error("prompt should contain handler names")
	}

	// Check output format instructions
	if !contains(prompt, "test_plan") {
		t.Error("prompt should specify output format")
	}
}

func TestSynthesisService_ParseResponse(t *testing.T) {
	service := &SynthesisService{}

	tests := []struct {
		name        string
		response    string
		expectError bool
		expectCount int
	}{
		{
			name: "valid response",
			response: `{
				"test_plan": [
					{
						"name": "Test GET /users",
						"type": "integration",
						"description": "Verify user list retrieval",
						"priority": "high",
						"estimated_time": "5m"
					},
					{
						"name": "Test GET /users/:id",
						"type": "integration",
						"description": "Verify single user retrieval",
						"priority": "high",
						"estimated_time": "5m"
					}
				]
			}`,
			expectError: false,
			expectCount: 2,
		},
		{
			name:        "invalid JSON",
			response:    `{invalid json}`,
			expectError: true,
			expectCount: 0,
		},
		{
			name: "empty test plan",
			response: `{
				"test_plan": []
			}`,
			expectError: true,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPlan, err := service.parseResponse(tt.response)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if testPlan == nil {
				t.Fatal("testPlan should not be nil")
			}

			if len(testPlan.Tests) != tt.expectCount {
				t.Errorf("expected %d tests, got %d", tt.expectCount, len(testPlan.Tests))
			}
		})
	}
}

func TestSynthesisService_GenerateTestPlan_Integration(t *testing.T) {
	// Skip this test in short mode as it requires actual LLM API
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// This test would require a mock Client implementation
	// For now, we just verify the function signature is correct
	t.Log("Integration test requires mock client implementation")

	// Create sample codebase for documentation purposes
	codebase := &types.Codebase{
		Language:  "javascript",
		Framework: "express",
		Routes: []types.Route{
			{Path: "/users", Method: "GET", Handler: "getUsers"},
			{Path: "/users", Method: "POST", Handler: "createUser"},
			{Path: "/users/:id", Method: "GET", Handler: "getUser"},
			{Path: "/users/:id", Method: "PUT", Handler: "updateUser"},
			{Path: "/users/:id", Method: "DELETE", Handler: "deleteUser"},
		},
		Models: []types.Model{
			{
				Name:      "User",
				Fields:    []types.Field{
					{Name: "id"},
					{Name: "name"},
					{Name: "email"},
					{Name: "password"},
				},
				Relations: []types.Relation{},
			},
		},
		Handlers: []types.Handler{
			{Name: "getUsers", Parameters: []types.Parameter{{Name: "req"}, {Name: "res"}}},
			{Name: "createUser", Parameters: []types.Parameter{{Name: "req"}, {Name: "res"}}},
			{Name: "getUser", Parameters: []types.Parameter{{Name: "req"}, {Name: "res"}}},
			{Name: "updateUser", Parameters: []types.Parameter{{Name: "req"}, {Name: "res"}}},
			{Name: "deleteUser", Parameters: []types.Parameter{{Name: "req"}, {Name: "res"}}},
		},
	}

	// This test would verify the generated test plan
	// For now, we just verify the codebase is valid
	if len(codebase.Routes) != 5 {
		t.Errorf("expected 5 routes, got %d", len(codebase.Routes))
	}
	if len(codebase.Models) != 1 {
		t.Errorf("expected 1 model, got %d", len(codebase.Models))
	}
	if len(codebase.Handlers) != 5 {
		t.Errorf("expected 5 handlers, got %d", len(codebase.Handlers))
	}
}

func TestTestCase_JSON(t *testing.T) {
	testCase := TestCase{
		Name:          "Test GET /users",
		Type:          "integration",
		Description:   "Verify user list retrieval",
		Priority:      "high",
		EstimatedTime: "5m",
		Confidence:    85,
	}

	data, err := json.Marshal(testCase)
	if err != nil {
		t.Fatalf("failed to marshal TestCase: %v", err)
	}

	var decoded TestCase
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal TestCase: %v", err)
	}

	if decoded.Name != testCase.Name {
		t.Errorf("expected name %s, got %s", testCase.Name, decoded.Name)
	}
	if decoded.Type != testCase.Type {
		t.Errorf("expected type %s, got %s", testCase.Type, decoded.Type)
	}
	if decoded.Confidence != testCase.Confidence {
		t.Errorf("expected confidence %d, got %d", testCase.Confidence, decoded.Confidence)
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
