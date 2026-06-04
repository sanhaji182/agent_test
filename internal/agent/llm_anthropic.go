package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// AnthropicLLM adalah implementasi LLM menggunakan Anthropic Claude API
type AnthropicLLM struct {
	client anthropic.Client
	model  string
}

// NewAnthropicLLM membuat client Anthropic baru dengan API key dan model
func NewAnthropicLLM(apiKey, model string) *AnthropicLLM {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &AnthropicLLM{client: client, model: model}
}

// AnalyzeCodebase menganalisis kode project dan mengembalikan ringkasan terstruktur
func (a *AnthropicLLM) AnalyzeCodebase(ctx context.Context, path string) (string, error) {
	prompt := fmt.Sprintf(`Analyze the codebase at path: %s
Detect: language, framework, routes, controllers, models, API endpoints.
Return a structured summary.`, path)

	return a.chat(ctx, prompt)
}

// GenerateTestPlan membuat rencana pengujian berdasarkan analisis kode
func (a *AnthropicLLM) GenerateTestPlan(ctx context.Context, analysis, requirements string) (*TestPlan, error) {
	prompt := fmt.Sprintf(`Based on this codebase analysis:
%s

And these requirements: %s

Generate a test plan as JSON with this structure:
{"summary": "...", "scenarios": [{"name": "...", "priority": "high|medium|low", "steps": ["..."]}]}

Return ONLY valid JSON, no markdown.`, analysis, requirements)

	resp, err := a.chat(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Bersihkan response dari markdown code fence jika ada
	resp = stripJSONMarkers(resp)
	var plan TestPlan
	if err := json.Unmarshal([]byte(resp), &plan); err != nil {
		return nil, fmt.Errorf("parse test plan: %w", err)
	}
	return &plan, nil
}

// GenerateTestScripts membuat file test Playwright dari test plan
func (a *AnthropicLLM) GenerateTestScripts(ctx context.Context, plan *TestPlan, analysis string) ([]TestFile, error) {
	planJSON, _ := json.Marshal(plan)
	prompt := fmt.Sprintf(`Generate Playwright automation actions for this test plan:
%s

Codebase context: %s

You must return a JSON array of files. Each file represents a test scenario.
For the 'content' field, provide a JSON array of actions as a string.
Supported actions:
- {"action": "goto", "url": "..."}
- {"action": "fill", "selector": "...", "value": "..."}
- {"action": "click", "selector": "..."}
- {"action": "scroll", "y": 500}
- {"action": "wait", "ms": 2000}

Format Example:
[
  {
    "name": "scenario1.json",
    "content": "[\n  {\"action\": \"goto\", \"url\": \"https://example.com\"},\n  {\"action\": \"fill\", \"selector\": \"input[name='q']\", \"value\": \"test\"},\n  {\"action\": \"click\", \"selector\": \"button[type='submit']\"}\n]"
  }
]

Return ONLY valid JSON, no markdown.`, string(planJSON), analysis)

	resp, err := a.chat(ctx, prompt)
	if err != nil {
		return nil, err
	}

	resp = stripJSONMarkers(resp)
	var files []TestFile
	if err := json.Unmarshal([]byte(resp), &files); err != nil {
		return nil, fmt.Errorf("parse test files: %w", err)
	}
	return files, nil
}

// SuggestFixes meminta LLM untuk memperbaiki test yang gagal
func (a *AnthropicLLM) SuggestFixes(ctx context.Context, failures []Failure, files []TestFile) ([]TestFile, error) {
	failJSON, _ := json.Marshal(failures)
	filesJSON, _ := json.Marshal(files)
	prompt := fmt.Sprintf(`These Playwright tests failed:
Failures: %s
Original files: %s

Fix the test files. Return JSON array: [{"name": "...", "content": "..."}]
Return ONLY valid JSON, no markdown.`, string(failJSON), string(filesJSON))

	resp, err := a.chat(ctx, prompt)
	if err != nil {
		return nil, err
	}

	resp = stripJSONMarkers(resp)
	var fixed []TestFile
	if err := json.Unmarshal([]byte(resp), &fixed); err != nil {
		return nil, fmt.Errorf("parse fixes: %w", err)
	}
	return fixed, nil
}

// HealAction meminta LLM untuk memperbaiki aksi tunggal Playwright (Self-Healing)
func (a *AnthropicLLM) HealAction(ctx context.Context, action string, domSnapshot string, errorMsg string) (string, error) {
	prompt := fmt.Sprintf(`A Playwright browser action failed.
Target Action: %s
Error Encountered: %s

Current Simplified DOM:
%s

Please provide a CORRECTED JSON action to replace the failed one.
Supported actions format:
- {"action": "goto", "url": "..."}
- {"action": "fill", "selector": "...", "value": "..."}
- {"action": "click", "selector": "..."}
- {"action": "scroll", "y": 500}
- {"action": "wait", "ms": 2000}

Analyze the DOM to find the correct selector if the old one failed.
Return ONLY valid JSON for the single corrected action, no markdown, no explanation.`, action, errorMsg, domSnapshot)

	resp, err := a.chat(ctx, prompt)
	if err != nil {
		return "", err
	}
	return stripJSONMarkers(resp), nil
}

// HealActionWithVision meminta LLM untuk memperbaiki aksi dengan menyertakan screenshot dari browser
func (a *AnthropicLLM) HealActionWithVision(ctx context.Context, action string, domSnapshot string, errorMsg string, imageBase64 string) (string, error) {
	prompt := fmt.Sprintf(`A Playwright browser action failed.
Target Action: %s
Error Encountered: %s

Current Simplified DOM:
%s

I have also attached a screenshot of the current page state.
Please analyze BOTH the screenshot and the DOM to find the correct selector if the old one failed, or if the UI has changed.

Provide a CORRECTED JSON action to replace the failed one.
Supported actions format:
- {"action": "goto", "url": "..."}
- {"action": "fill", "selector": "...", "value": "..."}
- {"action": "click", "selector": "..."}
- {"action": "scroll", "y": 500}
- {"action": "wait", "ms": 2000}

Return ONLY valid JSON for the single corrected action, no markdown, no explanation.`, action, errorMsg, domSnapshot)

	resp, err := a.chatWithVision(ctx, prompt, imageBase64)
	if err != nil {
		return "", err
	}
	return stripJSONMarkers(resp), nil
}

// chat mengirim pesan ke Anthropic API dan mengembalikan response teks
func (a *AnthropicLLM) chat(ctx context.Context, prompt string) (string, error) {
	msg, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(a.model),
		MaxTokens: int64(4096),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return "", fmt.Errorf("anthropic: %w", err)
	}

	// Ambil teks dari response content block pertama
	for _, block := range msg.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}
	return "", fmt.Errorf("no text in response")
}

// chatWithVision mengirim pesan teks dan gambar (base64) ke Anthropic API
func (a *AnthropicLLM) chatWithVision(ctx context.Context, prompt string, imageBase64 string) (string, error) {
	msg, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(a.model),
		MaxTokens: int64(4096),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewImageBlockBase64("image/jpeg", imageBase64),
				anthropic.NewTextBlock(prompt),
			),
		},
	})
	if err != nil {
		return "", fmt.Errorf("anthropic vision: %w", err)
	}

	for _, block := range msg.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}
	return "", fmt.Errorf("no text in response")
}

// stripJSONMarkers membersihkan markdown code fence dari response LLM
// Contoh: ```json\n{...}\n``` → {...}
func stripJSONMarkers(s string) string {
	s = strings.TrimPrefix(s, "```json\n")
	s = strings.TrimPrefix(s, "```\n")
	s = strings.TrimSuffix(s, "\n```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
