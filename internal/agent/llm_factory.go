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
	c := newAIClient(ProviderConfig{Provider: provider, Model: model, APIKey: apiKey, BaseURL: baseURL})
	if c == nil {
		return nil
	}
	return &promptLLM{client: c}
}

// ProviderConfig describes a single LLM provider endpoint for fallback chaining.
// MaxTokens and Temperature are optional; zero values fall back to sane defaults
// (4096 tokens, 0.2) in newAIClient.
type ProviderConfig struct {
	Provider    string
	Model       string
	APIKey      string
	BaseURL     string
	MaxTokens   int64
	Temperature float64
}

// newAIClient builds the raw ai transport for a provider (no prompt adapter).
// Returns nil for an unknown provider. Mirrors NewLLM's routing exactly so the
// single-provider and fallback paths cannot drift.
func newAIClient(p ProviderConfig) ai.Client {
	normalized := strings.ToLower(strings.TrimSpace(p.Provider))
	maxTokens := p.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}
	temperature := p.Temperature
	if temperature <= 0 {
		temperature = 0.2
	}
	switch normalized {
	case "anthropic", "":
		return ai.NewAnthropicClient(p.APIKey, p.Model, maxTokens)
	case "openai", "google", "deepseek", "mistral", "groq", "openrouter",
		"custom", "local", "ollama", "openai-compatible", "huggingface":
		baseURL := p.BaseURL
		if baseURL == "" {
			baseURL = ai.DefaultOpenAICompatibleBaseURL(normalized)
		}
		return ai.NewOpenAICompatibleClient(ai.Config{
			Provider:    "openai-compatible",
			Model:       p.Model,
			APIKey:      p.APIKey,
			BaseURL:     baseURL,
			MaxTokens:   maxTokens,
			Temperature: temperature,
		})
	default:
		return nil
	}
}

// NewFallbackLLM builds an LLM that fails over across the given providers in
// order: the first provider is tried first and each subsequent provider is used
// only when the previous one errors. Providers with an unknown name are skipped.
// With a single usable provider this behaves exactly like NewLLM. Returns nil
// when no usable provider is configured.
func NewFallbackLLM(providers ...ProviderConfig) LLM {
	clients := make([]ai.Client, 0, len(providers))
	for _, p := range providers {
		if c := newAIClient(p); c != nil {
			clients = append(clients, c)
		}
	}
	c := ai.NewFallbackClient(clients...)
	if c == nil {
		return nil
	}
	return &promptLLM{client: c}
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
