package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type AnthropicLLM struct {
	client anthropic.Client
	model  string
}

func NewAnthropicLLM(apiKey, model string) *AnthropicLLM {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &AnthropicLLM{client: client, model: model}
}

func (a *AnthropicLLM) AnalyzeCodebase(ctx context.Context, path string) (string, error) {
	prompt := fmt.Sprintf(`Analyze the codebase at path: %s
Detect: language, framework, routes, controllers, models, API endpoints.
Return a structured summary.`, path)

	return a.chat(ctx, prompt)
}

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

	resp = stripJSONMarkers(resp)
	var plan TestPlan
	if err := json.Unmarshal([]byte(resp), &plan); err != nil {
		return nil, fmt.Errorf("parse test plan: %w", err)
	}
	return &plan, nil
}

func (a *AnthropicLLM) GenerateTestScripts(ctx context.Context, plan *TestPlan, analysis string) ([]TestFile, error) {
	planJSON, _ := json.Marshal(plan)
	prompt := fmt.Sprintf(`Generate Playwright TypeScript test files for this test plan:
%s

Codebase context: %s

Return JSON array: [{"name": "test-name.spec.ts", "content": "..."}]
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

	for _, block := range msg.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}
	return "", fmt.Errorf("no text in response")
}

func stripJSONMarkers(s string) string {
	s = strings.TrimPrefix(s, "```json\n")
	s = strings.TrimPrefix(s, "```\n")
	s = strings.TrimSuffix(s, "\n```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
