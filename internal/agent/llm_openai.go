package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAILLM implements the LLM interface using the OpenAI-compatible REST API.
// Prompt dan parsing dibagi dengan implementasi lain via llm_prompts.go
// (ADR-006 Step B).
type OpenAILLM struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// NewOpenAILLM creates a new OpenAI-compatible client
func NewOpenAILLM(apiKey, model, baseURL string) *OpenAILLM {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	// Pastikan baseURL tidak diakhiri dengan slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &OpenAILLM{
		apiKey:  apiKey,
		model:   model,
		baseURL: baseURL,
		// LLM completions can be slow (large prompts, reasoning models), but a
		// hung connection must not pin a run goroutine forever.
		client: &http.Client{Timeout: 2 * time.Minute},
	}
}

func (a *OpenAILLM) chat(ctx context.Context, prompt string, imageBase64 string) (string, error) {
	url := a.baseURL + "/chat/completions"

	var content interface{}
	if imageBase64 != "" {
		content = []map[string]interface{}{
			{
				"type": "text",
				"text": prompt,
			},
			{
				"type": "image_url",
				"image_url": map[string]interface{}{
					"url": fmt.Sprintf("data:image/jpeg;base64,%s", imageBase64),
				},
			},
		}
	} else {
		content = prompt
	}

	payload := map[string]interface{}{
		"model": a.model,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": content,
			},
		},
		"temperature": 0.2,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if a.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+a.apiKey)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse openai response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("empty choices in openai response")
	}

	return result.Choices[0].Message.Content, nil
}

// AnalyzeCodebase menganalisis kode project dan mengembalikan ringkasan terstruktur
func (a *OpenAILLM) AnalyzeCodebase(ctx context.Context, path string) (string, error) {
	return a.chat(ctx, promptAnalyzeCodebase(path), "")
}

// GenerateTestPlan membuat rencana pengujian berdasarkan analisis kode
func (a *OpenAILLM) GenerateTestPlan(ctx context.Context, analysis, requirements string) (*TestPlan, error) {
	resp, err := a.chat(ctx, promptGenerateTestPlan(analysis, requirements), "")
	if err != nil {
		return nil, err
	}
	return parseTestPlan(resp)
}

// GenerateTestScripts membuat file test Playwright dari test plan
func (a *OpenAILLM) GenerateTestScripts(ctx context.Context, plan *TestPlan, analysis string) ([]TestFile, error) {
	resp, err := a.chat(ctx, promptGenerateTestScripts(plan, analysis), "")
	if err != nil {
		return nil, err
	}
	return parseTestFiles(resp)
}

// SuggestFixes meminta LLM untuk memperbaiki test yang gagal
func (a *OpenAILLM) SuggestFixes(ctx context.Context, failures []Failure, files []TestFile) ([]TestFile, error) {
	resp, err := a.chat(ctx, promptSuggestFixes(failures, files), "")
	if err != nil {
		return nil, err
	}
	return parseFixedFiles(resp)
}

// HealAction meminta LLM untuk memperbaiki aksi tunggal Playwright (Self-Healing).
// Menggunakan prompt non-vision khusus (sebelumnya memakai prompt vision yang
// keliru mengklaim ada screenshot terlampir — diperbaiki di ADR-006 Step B).
func (a *OpenAILLM) HealAction(ctx context.Context, action string, domSnapshot string, errorMsg string) (string, error) {
	resp, err := a.chat(ctx, promptHealAction(action, domSnapshot, errorMsg), "")
	if err != nil {
		return "", err
	}
	return stripJSONMarkers(resp), nil
}

// HealActionWithVision meminta LLM untuk memperbaiki aksi dengan menyertakan screenshot dari browser
func (a *OpenAILLM) HealActionWithVision(ctx context.Context, action string, domSnapshot string, errorMsg string, imageBase64 string) (string, error) {
	if imageBase64 == "" {
		return a.HealAction(ctx, action, domSnapshot, errorMsg)
	}
	resp, err := a.chat(ctx, promptHealActionWithVision(action, domSnapshot, errorMsg), imageBase64)
	if err != nil {
		return "", err
	}
	return stripJSONMarkers(resp), nil
}
