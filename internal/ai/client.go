// Package ai provides provider-agnostic text generation for planning workflows.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
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
	// GenerateWithImage sends a prompt plus a base64-encoded JPEG for
	// vision-capable models (ADR-006 Step A). Implementations without an
	// image simply route through the text path when imageBase64 is empty.
	GenerateWithImage(ctx context.Context, prompt, imageBase64 string) (string, error)
}

func New(cfg Config) Client {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if provider == "" {
		provider = "anthropic"
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 4096
	}
	var inner Client
	switch provider {
	case "anthropic":
		if cfg.APIKey == "" {
			return nil
		}
		inner = NewAnthropicClient(cfg.APIKey, cfg.Model, cfg.MaxTokens)
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
		inner = NewOpenAICompatibleClient(cfg)
	default:
		return nil
	}
	// Wrap with retry + circuit breaker for production resilience
	return NewResilientClient(inner)
}

// NewAnthropicClient constructs the Anthropic transport directly, without
// New's empty-key gating. The execution layer (agent.NewLLM) uses this to
// preserve its historical contract: constructors always succeed and missing
// credentials surface at request time (ADR-006 Step C).
func NewAnthropicClient(apiKey, model string, maxTokens int64) *AnthropicClient {
	if maxTokens == 0 {
		maxTokens = 4096
	}
	return &AnthropicClient{client: anthropic.NewClient(option.WithAPIKey(apiKey)), model: model, maxTokens: maxTokens}
}

// NewOpenAICompatibleClient constructs the OpenAI-compatible transport
// directly, without New's empty-key gating (see NewAnthropicClient).
func NewOpenAICompatibleClient(cfg Config) *OpenAICompatibleClient {
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultOpenAICompatibleBaseURL(cfg.Provider)
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 4096
	}
	// Bounded client: a hung LLM endpoint must not pin callers forever. Code
	// generation calls (multiple full test files) can legitimately take several
	// minutes, so the bound is generous rather than aggressive.
	return &OpenAICompatibleClient{cfg: cfg, http: &http.Client{Timeout: 5 * time.Minute}}
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
	return c.generate(ctx, []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
	})
}

// GenerateWithImage sends prompt + base64 JPEG using Anthropic image blocks
// (lifted from agent.AnthropicLLM.chatWithVision — ADR-006 Step A).
func (c *AnthropicClient) GenerateWithImage(ctx context.Context, prompt, imageBase64 string) (string, error) {
	if imageBase64 == "" {
		return c.GenerateText(ctx, prompt)
	}
	return c.generate(ctx, []anthropic.MessageParam{
		anthropic.NewUserMessage(
			anthropic.NewImageBlockBase64("image/jpeg", imageBase64),
			anthropic.NewTextBlock(prompt),
		),
	})
}

func (c *AnthropicClient) generate(ctx context.Context, messages []anthropic.MessageParam) (string, error) {
	msg, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: c.maxTokens,
		Messages:  messages,
	})
	if err != nil {
		return "", fmt.Errorf("anthropic: %s", cleanAnthropicError(err))
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
	return c.generate(ctx, prompt, "")
}

// GenerateWithImage sends prompt + base64 JPEG using OpenAI image_url content
// blocks (lifted from agent.OpenAILLM.chat — ADR-006 Step A).
func (c *OpenAICompatibleClient) GenerateWithImage(ctx context.Context, prompt, imageBase64 string) (string, error) {
	return c.generate(ctx, prompt, imageBase64)
}

func (c *OpenAICompatibleClient) generate(ctx context.Context, prompt, imageBase64 string) (string, error) {
	url := strings.TrimRight(c.cfg.BaseURL, "/") + "/chat/completions"
	var content interface{} = prompt
	if imageBase64 != "" {
		content = []map[string]interface{}{
			{"type": "text", "text": prompt},
			{"type": "image_url", "image_url": map[string]interface{}{
				"url": fmt.Sprintf("data:image/jpeg;base64,%s", imageBase64),
			}},
		}
	}
	payload := map[string]interface{}{
		"model":       c.cfg.Model,
		"temperature": c.cfg.Temperature,
		"max_tokens":  c.cfg.MaxTokens,
		// Explicit: some OpenAI-compatible gateways default to SSE streaming,
		// which breaks JSON decoding ("invalid character 'd'" from "data:" lines).
		"stream": false,
		"messages": []map[string]interface{}{
			{"role": "user", "content": content},
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
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai-compatible: status %d: %s", resp.StatusCode, extractProviderError(respBody))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("openai-compatible: no choices")
	}
	return parsed.Choices[0].Message.Content, nil
}

// anthropicStatusRe captures the HTTP status code from the anthropic-sdk-go error
// string, which has the form: POST "https://...": 402 Payment Required {...}.
var anthropicStatusRe = regexp.MustCompile(`":\s+(\d{3})\s`)

// cleanAnthropicError converts an anthropic-sdk-go error into a concise message
// that surfaces the provider's actual error text (from the response body's
// error.message), consistent with the OpenAI-compatible path. The SDK's internal
// apierror type is not importable, so we parse its well-defined Error() string.
func cleanAnthropicError(err error) string {
	s := err.Error()
	// The response body (raw JSON) trails the first '{'; reuse the shared
	// extractor since Anthropic's body also uses {"error":{"message":"..."}}.
	msg := ""
	if idx := strings.Index(s, "{"); idx >= 0 {
		msg = extractProviderError([]byte(s[idx:]))
	}
	status := ""
	if m := anthropicStatusRe.FindStringSubmatch(s); m != nil {
		status = m[1]
	}
	switch {
	case status != "" && msg != "":
		return "status " + status + ": " + msg
	case msg != "":
		return msg
	default:
		return s
	}
}

// extractProviderError extracts the most useful error message from an
// OpenAI-compatible provider's error response body. Providers return
// {"error":{"message":"..."}}; some proxies further nest the upstream error as a
// JSON string inside that message, so we make a best-effort pass to surface the
// innermost human-readable message (e.g. "Insufficient balance...") instead of
// dumping the raw nested blob.
func extractProviderError(body []byte) string {
	raw := strings.TrimSpace(string(body))
	if raw == "" {
		return "empty response body"
	}
	// HTML error page (not JSON) — the endpoint is likely down or misconfigured.
	if strings.HasPrefix(raw, "<") {
		return "endpoint returned HTML instead of JSON (likely an error page): " + truncateForError(raw, 200)
	}
	// Standard OpenAI error envelope: {"error": {"message": "..."}}.
	var envelope struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(raw), &envelope); err == nil && envelope.Error.Message != "" {
		if inner := extractNestedErrorMessage(envelope.Error.Message); inner != "" {
			return inner
		}
		return envelope.Error.Message
	}
	// Unrecognized shape — return the truncated raw body.
	return truncateForError(raw, 300)
}

// extractNestedErrorMessage finds the first JSON object embedded in s and returns
// its "message" field, if any. Some proxies embed the upstream provider's error as a
// stringified JSON inside the message; this surfaces the actual provider text.
// Best-effort: returns "" when no nested message is found.
func extractNestedErrorMessage(s string) string {
	start := strings.Index(s, "{")
	if start < 0 {
		return ""
	}
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				var nested struct {
					Message string `json:"message"`
				}
				if err := json.Unmarshal([]byte(s[start:i+1]), &nested); err == nil && nested.Message != "" {
					return nested.Message
				}
				return ""
			}
		}
	}
	return ""
}

func truncateForError(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
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
