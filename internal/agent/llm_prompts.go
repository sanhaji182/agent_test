package agent

import (
	"encoding/json"
	"fmt"
)

// llm_prompts.go — shared prompt builders and response parsers for all LLM
// implementations (ADR-006 Step B). Extracted from llm_anthropic.go and
// llm_openai.go, which had drifted: the Anthropic variants are canonical
// (GenerateTestScripts includes the Format Example section; HealAction uses
// a dedicated non-vision prompt). The OpenAI path adopting these is a strict
// improvement — richer instructions, same contract.

func promptAnalyzeCodebase(path string) string {
	return fmt.Sprintf(`Analyze the codebase at path: %s
Detect: language, framework, routes, controllers, models, API endpoints.
Return a structured summary.`, path)
}

func promptGenerateTestPlan(analysis, requirements string) string {
	return fmt.Sprintf(`Based on this codebase analysis:
%s

And these requirements: %s

Generate a test plan as JSON with this structure:
{"summary": "...", "scenarios": [{"name": "...", "priority": "high|medium|low", "steps": ["..."]}]}

Return ONLY valid JSON, no markdown.`, analysis, requirements)
}

func promptGenerateTestScripts(plan *TestPlan, analysis string) string {
	planJSON, _ := json.Marshal(plan)
	return fmt.Sprintf(`Generate Playwright automation actions for this test plan:
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
}

func promptSuggestFixes(failures []Failure, files []TestFile) string {
	failJSON, _ := json.Marshal(failures)
	filesJSON, _ := json.Marshal(files)
	return fmt.Sprintf(`These Playwright tests failed:
Failures: %s
Original files: %s

Fix the test files. Return JSON array: [{"name": "...", "content": "..."}]
Return ONLY valid JSON, no markdown.`, string(failJSON), string(filesJSON))
}

func promptHealAction(action, domSnapshot, errorMsg string) string {
	return fmt.Sprintf(`A Playwright browser action failed.
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
}

func promptHealActionWithVision(action, domSnapshot, errorMsg string) string {
	return fmt.Sprintf(`A Playwright browser action failed.
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
}

// --- Response parsers ---

func parseTestPlan(raw string) (*TestPlan, error) {
	raw = stripJSONMarkers(raw)
	var plan TestPlan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		return nil, fmt.Errorf("parse test plan: %w", err)
	}
	return &plan, nil
}

func parseTestFiles(raw string) ([]TestFile, error) {
	raw = stripJSONMarkers(raw)
	var files []TestFile
	if err := json.Unmarshal([]byte(raw), &files); err != nil {
		return nil, fmt.Errorf("parse test files: %w", err)
	}
	return files, nil
}

func parseFixedFiles(raw string) ([]TestFile, error) {
	raw = stripJSONMarkers(raw)
	var fixed []TestFile
	if err := json.Unmarshal([]byte(raw), &fixed); err != nil {
		return nil, fmt.Errorf("parse fixes: %w", err)
	}
	return fixed, nil
}
