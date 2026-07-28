// Package ai provides provider-agnostic text generation for planning workflows.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type Config struct {
	Provider    string
	Model       string
	APIKey      string
	BaseURL     string
	MaxTokens   int64
	Temperature float64
}

type Client interface {
	GenerateText(ctx context.Context, prompt string) (string, error)
}

func New(cfg Config) Client {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if provider == "" {
		provider = "anthropic"
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 4096
	}
	switch provider {
	case "anthropic":
		if cfg.APIKey == "" {
			return nil
		}
		return &AnthropicClient{client: anthropic.NewClient(option.WithAPIKey(cfg.APIKey)), model: cfg.Model, maxTokens: cfg.MaxTokens}
	case "openai", "openai-compatible", "ollama", "local", "custom",
		"google", "deepseek", "mistral", "groq", "openrouter", "huggingface":
		// Same OpenAI-compatible provider set as agent.NewLLM (DL-2 routing
		// alignment): hosted providers expose /chat/completions endpoints.
		if cfg.BaseURL == "" {
			cfg.BaseURL = DefaultOpenAICompatibleBaseURL(provider)
		}
		if cfg.APIKey == "" && provider != "ollama" && provider != "local" {
			return nil
		}
		// Bounded client: a hung LLM endpoint must not pin callers forever.
		return &OpenAICompatibleClient{cfg: cfg, http: &http.Client{Timeout: 2 * time.Minute}}
	default:
		return nil
	}
}

// DefaultOpenAICompatibleBaseURL maps a provider name to its OpenAI-compatible
// endpoint when no explicit base URL is configured. Shared by both LLM layers
// (ai.New and agent.NewLLM) so provider routing cannot drift (DL-2). Mirrors
// the approved-origin list in api.isApprovedLLMOrigin (ADR-005 Phase 2).
func DefaultOpenAICompatibleBaseURL(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "google":
		return "https://generativelanguage.googleapis.com/v1beta/openai"
	case "deepseek":
		return "https://api.deepseek.com/v1"
	case "mistral":
		return "https://api.mistral.ai/v1"
	case "groq":
		return "https://api.groq.com/openai/v1"
	case "openrouter":
		return "https://openrouter.ai/api/v1"
	default:
		return "https://api.openai.com/v1"
	}
}

func ConfigFromEnv() Config {
	provider := getenv("LLM_PROVIDER", "anthropic")
	apiKey := getenv("LLM_API_KEY", "")
	if apiKey == "" && provider == "anthropic" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if apiKey == "" && (provider == "openai" || provider == "openai-compatible") {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	return Config{
		Provider:    provider,
		Model:       getenv("LLM_MODEL", "claude-sonnet-4-5"),
		APIKey:      apiKey,
		BaseURL:     os.Getenv("LLM_BASE_URL"),
		MaxTokens:   4096,
		Temperature: 0.2,
	}
}

type AnthropicClient struct {
	client    anthropic.Client
	model     string
	maxTokens int64
}

func (c *AnthropicClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	msg, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: c.maxTokens,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return "", fmt.Errorf("anthropic: %w", err)
	}
	for _, block := range msg.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}
	return "", fmt.Errorf("anthropic: no text content")
}

type OpenAICompatibleClient struct {
	cfg  Config
	http *http.Client
}

func (c *OpenAICompatibleClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	url := strings.TrimRight(c.cfg.BaseURL, "/") + "/chat/completions"
	payload := map[string]interface{}{
		"model":       c.cfg.Model,
		"temperature": c.cfg.Temperature,
		"max_tokens":  c.cfg.MaxTokens,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai-compatible: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai-compatible: status %d", resp.StatusCode)
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("openai-compatible: no choices")
	}
	return parsed.Choices[0].Message.Content, nil
}

func StripJSONMarkers(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json\n")
	s = strings.TrimPrefix(s, "```\n")
	s = strings.TrimSuffix(s, "\n```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
