package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OpenAILLM implements the LLM interface using the OpenAI-compatible REST API
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
		client:  &http.Client{},
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
	prompt := fmt.Sprintf(`Analyze the codebase at path: %s
Detect: language, framework, routes, controllers, models, API endpoints.
Return a structured summary.`, path)

	return a.chat(ctx, prompt, "")
}

// GenerateTestPlan membuat rencana pengujian berdasarkan analisis kode
func (a *OpenAILLM) GenerateTestPlan(ctx context.Context, analysis, requirements string) (*TestPlan, error) {
	prompt := fmt.Sprintf(`Based on this codebase analysis:
%s

And these requirements: %s

Generate a test plan as JSON with this structure:
{"summary": "...", "scenarios": [{"name": "...", "priority": "high|medium|low", "steps": ["..."]}]}

Return ONLY valid JSON, no markdown.`, analysis, requirements)

	resp, err := a.chat(ctx, prompt, "")
	if err != nil {
		return nil, err
	}

	resp = stripJSONMarkers(resp)
	var plan TestPlan
	if err := json.Unmarshal([]byte(resp), &plan); err != nil {
		return nil, fmt.Errorf("parse test plan: %w", err)
	}
	return &plan, nil
}

// GenerateTestScripts membuat file test Playwright dari test plan
func (a *OpenAILLM) GenerateTestScripts(ctx context.Context, plan *TestPlan, analysis string) ([]TestFile, error) {
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

Return ONLY valid JSON, no markdown.`, string(planJSON), analysis)

	resp, err := a.chat(ctx, prompt, "")
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
func (a *OpenAILLM) SuggestFixes(ctx context.Context, failures []Failure, files []TestFile) ([]TestFile, error) {
	failJSON, _ := json.Marshal(failures)
	filesJSON, _ := json.Marshal(files)
	prompt := fmt.Sprintf(`These Playwright tests failed:
Failures: %s
Original files: %s

Fix the test files. Return JSON array: [{"name": "...", "content": "..."}]
Return ONLY valid JSON, no markdown.`, string(failJSON), string(filesJSON))

	resp, err := a.chat(ctx, prompt, "")
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
func (a *OpenAILLM) HealAction(ctx context.Context, action string, domSnapshot string, errorMsg string) (string, error) {
	return a.HealActionWithVision(ctx, action, domSnapshot, errorMsg, "")
}

// HealActionWithVision meminta LLM untuk memperbaiki aksi dengan menyertakan screenshot dari browser
func (a *OpenAILLM) HealActionWithVision(ctx context.Context, action string, domSnapshot string, errorMsg string, imageBase64 string) (string, error) {
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

	resp, err := a.chat(ctx, prompt, imageBase64)
	if err != nil {
		return "", err
	}
	return stripJSONMarkers(resp), nil
}
