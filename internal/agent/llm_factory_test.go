package agent_test

import (
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/ai"
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
		// Normalization (DL-2): DB-stored settings may carry arbitrary casing
		{name: "uppercase anthropic", provider: "ANTHROPIC", nilOK: false},
		{name: "mixed case with spaces", provider: "  OpenAI  ", nilOK: false},
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

// TestProviderRoutingParity guards DL-2: every provider accepted by the
// execution layer (agent.NewLLM) must also be accepted by the planning layer
// (ai.New), and vice versa. If this fails, the two factories have drifted.
func TestProviderRoutingParity(t *testing.T) {
	providers := []string{
		"", "anthropic",
		"openai", "google", "deepseek", "mistral", "groq", "openrouter",
		"custom", "local", "ollama", "openai-compatible", "huggingface",
		"some-unknown-xyz",
	}
	for _, p := range providers {
		t.Run("provider="+p, func(t *testing.T) {
			agentLLM := agent.NewLLM(p, "test-model", "test-key", "https://test.example.com/v1")
			aiClient := ai.New(ai.Config{Provider: p, Model: "test-model", APIKey: "test-key", BaseURL: "https://test.example.com/v1"})
			if (agentLLM == nil) != (aiClient == nil) {
				t.Fatalf("routing drift for %q: agent.NewLLM nil=%v, ai.New nil=%v", p, agentLLM == nil, aiClient == nil)
			}
		})
	}
}

// TestDefaultBaseURLParity ensures hosted providers resolve to per-provider
// endpoints (not silently api.openai.com) in both layers.
func TestDefaultBaseURLParity(t *testing.T) {
	cases := map[string]string{
		"google":     "https://generativelanguage.googleapis.com/v1beta/openai",
		"deepseek":   "https://api.deepseek.com/v1",
		"mistral":    "https://api.mistral.ai/v1",
		"groq":       "https://api.groq.com/openai/v1",
		"openrouter": "https://openrouter.ai/api/v1",
		"openai":     "https://api.openai.com/v1",
		"custom":     "https://api.openai.com/v1",
	}
	for provider, want := range cases {
		if got := ai.DefaultOpenAICompatibleBaseURL(provider); got != want {
			t.Fatalf("DefaultOpenAICompatibleBaseURL(%q) = %q, want %q", provider, got, want)
		}
	}
}
