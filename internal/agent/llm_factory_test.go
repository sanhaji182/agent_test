package agent_test

import (
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/agent"
)

func TestNewLLM_ProviderRouting(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		nilOK    bool
	}{
		// Anthropic SDK path
		{name: "empty defaults to anthropic", provider: "", nilOK: false},
		{name: "anthropic", provider: "anthropic", nilOK: false},
		// OpenAI-compatible REST path
		{name: "openai", provider: "openai", nilOK: false},
		{name: "google", provider: "google", nilOK: false},
		{name: "deepseek", provider: "deepseek", nilOK: false},
		{name: "mistral", provider: "mistral", nilOK: false},
		{name: "groq", provider: "groq", nilOK: false},
		{name: "openrouter", provider: "openrouter", nilOK: false},
		{name: "custom", provider: "custom", nilOK: false},
		{name: "local", provider: "local", nilOK: false},
		{name: "ollama", provider: "ollama", nilOK: false},
		{name: "openai-compatible", provider: "openai-compatible", nilOK: false},
		{name: "huggingface", provider: "huggingface", nilOK: false},
		// Unknown — should return nil
		{name: "unknown returns nil", provider: "some-unknown-xyz", nilOK: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := agent.NewLLM(tt.provider, "test-model", "test-key", "https://test.example.com/v1")
			isNil := llm == nil
			if isNil != tt.nilOK {
				t.Fatalf("NewLLM(%q) returned nil=%v, want nil=%v", tt.provider, isNil, tt.nilOK)
			}
		})
	}
}
