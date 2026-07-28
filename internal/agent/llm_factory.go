package agent

import (
	"strings"

	"github.com/go-go-golems/gotest-agent/internal/ai"
)

// NewLLM is the canonical provider factory for all execution paths.
// It routes provider names through the single ai transport layer and wraps
// the result in the shared prompt adapter (ADR-006 Step C), supporting all
// 7 approved origins (ADR-005 Phase 2) plus custom/local endpoints.
//
// Provider names are normalized (trimmed, lowercased) and hosted providers
// resolve their default base URL via ai.DefaultOpenAICompatibleBaseURL, so
// routing matches ai.New exactly (DL-2 routing alignment).
//
// Provider routing:
//   - "" or "anthropic" → Anthropic SDK transport
//   - "openai", "google", "deepseek", "mistral", "groq", "openrouter" → OpenAI-compatible REST transport
//   - "custom", "local", "ollama" → OpenAI-compatible REST (self-hosted, requires explicit api_key)
//   - unknown → nil (explicit rejection)
//
// Note: unlike ai.New, this factory does not reject empty API keys — the
// execution layer's historical contract is that constructors always succeed
// and credential errors surface at request time.
func NewLLM(provider, model, apiKey, baseURL string) LLM {
	normalized := strings.ToLower(strings.TrimSpace(provider))
	switch normalized {
	case "anthropic", "":
		return NewAnthropicLLM(apiKey, model)
	case "openai", "google", "deepseek", "mistral", "groq", "openrouter",
		"custom", "local", "ollama", "openai-compatible", "huggingface":
		if baseURL == "" {
			baseURL = ai.DefaultOpenAICompatibleBaseURL(normalized)
		}
		return NewOpenAILLM(apiKey, model, baseURL)
	default:
		return nil
	}
}

// NewAnthropicLLM returns an agent.LLM backed by the Anthropic transport in
// internal/ai. Retained as a named constructor for direct callers.
func NewAnthropicLLM(apiKey, model string) LLM {
	return &promptLLM{client: ai.NewAnthropicClient(apiKey, model, 4096)}
}

// NewOpenAILLM returns an agent.LLM backed by the OpenAI-compatible transport
// in internal/ai. Retained as a named constructor for direct callers.
func NewOpenAILLM(apiKey, model, baseURL string) LLM {
	return &promptLLM{client: ai.NewOpenAICompatibleClient(ai.Config{
		Provider:    "openai-compatible",
		Model:       model,
		APIKey:      apiKey,
		BaseURL:     baseURL,
		MaxTokens:   4096,
		Temperature: 0.2,
	})}
}
