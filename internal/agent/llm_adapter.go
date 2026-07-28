package agent

import (
	"context"

	"github.com/go-go-golems/gotest-agent/internal/ai"
)

// promptLLM implements the agent.LLM interface by composing the shared prompt
// builders (llm_prompts.go) with a single ai.Client transport (ADR-006 Step C).
// It replaces the former AnthropicLLM and OpenAILLM structs, which each carried
// their own duplicated HTTP/SDK transport.
type promptLLM struct {
	client ai.Client
}

func (p *promptLLM) AnalyzeCodebase(ctx context.Context, path string) (string, error) {
	return p.client.GenerateText(ctx, promptAnalyzeCodebase(path))
}

func (p *promptLLM) GenerateTestPlan(ctx context.Context, analysis, requirements string) (*TestPlan, error) {
	resp, err := p.client.GenerateText(ctx, promptGenerateTestPlan(analysis, requirements))
	if err != nil {
		return nil, err
	}
	return parseTestPlan(resp)
}

func (p *promptLLM) GenerateTestScripts(ctx context.Context, plan *TestPlan, analysis string) ([]TestFile, error) {
	resp, err := p.client.GenerateText(ctx, promptGenerateTestScripts(plan, analysis))
	if err != nil {
		return nil, err
	}
	return parseTestFiles(resp)
}

func (p *promptLLM) SuggestFixes(ctx context.Context, failures []Failure, files []TestFile) ([]TestFile, error) {
	resp, err := p.client.GenerateText(ctx, promptSuggestFixes(failures, files))
	if err != nil {
		return nil, err
	}
	return parseFixedFiles(resp)
}

func (p *promptLLM) HealAction(ctx context.Context, action, domSnapshot, errorMsg string) (string, error) {
	resp, err := p.client.GenerateText(ctx, promptHealAction(action, domSnapshot, errorMsg))
	if err != nil {
		return "", err
	}
	return stripJSONMarkers(resp), nil
}

func (p *promptLLM) HealActionWithVision(ctx context.Context, action, domSnapshot, errorMsg, imageBase64 string) (string, error) {
	// No image → non-vision prompt via the text path (matches prior behavior
	// after the Step B fix).
	if imageBase64 == "" {
		return p.HealAction(ctx, action, domSnapshot, errorMsg)
	}
	resp, err := p.client.GenerateWithImage(ctx, promptHealActionWithVision(action, domSnapshot, errorMsg), imageBase64)
	if err != nil {
		return "", err
	}
	return stripJSONMarkers(resp), nil
}
