package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// AnthropicLLM adalah implementasi LLM menggunakan Anthropic Claude API.
// Prompt dan parsing dibagi dengan implementasi lain via llm_prompts.go
// (ADR-006 Step B).
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
	return a.chat(ctx, promptAnalyzeCodebase(path))
}

// GenerateTestPlan membuat rencana pengujian berdasarkan analisis kode
func (a *AnthropicLLM) GenerateTestPlan(ctx context.Context, analysis, requirements string) (*TestPlan, error) {
	resp, err := a.chat(ctx, promptGenerateTestPlan(analysis, requirements))
	if err != nil {
		return nil, err
	}
	return parseTestPlan(resp)
}

// GenerateTestScripts membuat file test Playwright dari test plan
func (a *AnthropicLLM) GenerateTestScripts(ctx context.Context, plan *TestPlan, analysis string) ([]TestFile, error) {
	resp, err := a.chat(ctx, promptGenerateTestScripts(plan, analysis))
	if err != nil {
		return nil, err
	}
	return parseTestFiles(resp)
}

// SuggestFixes meminta LLM untuk memperbaiki test yang gagal
func (a *AnthropicLLM) SuggestFixes(ctx context.Context, failures []Failure, files []TestFile) ([]TestFile, error) {
	resp, err := a.chat(ctx, promptSuggestFixes(failures, files))
	if err != nil {
		return nil, err
	}
	return parseFixedFiles(resp)
}

// HealAction meminta LLM untuk memperbaiki aksi tunggal Playwright (Self-Healing)
func (a *AnthropicLLM) HealAction(ctx context.Context, action string, domSnapshot string, errorMsg string) (string, error) {
	resp, err := a.chat(ctx, promptHealAction(action, domSnapshot, errorMsg))
	if err != nil {
		return "", err
	}
	return stripJSONMarkers(resp), nil
}

// HealActionWithVision meminta LLM untuk memperbaiki aksi dengan menyertakan screenshot dari browser
func (a *AnthropicLLM) HealActionWithVision(ctx context.Context, action string, domSnapshot string, errorMsg string, imageBase64 string) (string, error) {
	resp, err := a.chatWithVision(ctx, promptHealActionWithVision(action, domSnapshot, errorMsg), imageBase64)
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
