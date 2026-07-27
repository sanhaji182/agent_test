package agent

// NewLLM is the canonical provider factory for all execution paths.
// It routes provider names through the correct client type, supporting
// all 7 approved origins (ADR-005 Phase 2) plus custom/local endpoints.
//
// Provider routing:
//   - "" or "anthropic" → AnthropicLLM (Anthropic SDK)
//   - "openai", "google", "deepseek", "mistral", "groq", "openrouter" → OpenAILLM (OpenAI-compatible REST)
//   - "custom", "local", "ollama" → OpenAILLM (self-hosted, requires explicit api_key)
//   - unknown → nil (explicit rejection)
func NewLLM(provider, model, apiKey, baseURL string) LLM {
	switch provider {
	case "anthropic", "":
		return NewAnthropicLLM(apiKey, model)
	case "openai", "google", "deepseek", "mistral", "groq", "openrouter",
		"custom", "local", "ollama", "openai-compatible", "huggingface":
		return NewOpenAILLM(apiKey, model, baseURL)
	default:
		return nil
	}
}
