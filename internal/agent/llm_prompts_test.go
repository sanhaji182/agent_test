package agent

import (
	"strings"
	"testing"
)

// Step B (ADR-006) guard: prompt builders are pure functions shared by every
// LLM implementation. These tests pin their contracts so transport refactors
// (Step C) cannot silently alter what models receive.

func TestPromptBuilders_ContainContract(t *testing.T) {
	plan := &TestPlan{Summary: "s", Scenarios: []Scenario{{Name: "n", Priority: "high", Steps: []string{"step"}}}}

	cases := []struct {
		name    string
		prompt  string
		musts   []string
		mustNot []string
	}{
		{
			name:   "analyze codebase",
			prompt: promptAnalyzeCodebase("/tmp/x"),
			musts:  []string{"/tmp/x", "language, framework"},
		},
		{
			name:   "generate test plan",
			prompt: promptGenerateTestPlan("ANALYSIS", "REQS"),
			musts:  []string{"ANALYSIS", "REQS", `"scenarios"`, "Return ONLY valid JSON"},
		},
		{
			name:   "generate test scripts includes format example",
			prompt: promptGenerateTestScripts(plan, "CTX"),
			musts:  []string{"CTX", "Format Example", `"action": "goto"`, "Return ONLY valid JSON"},
		},
		{
			name:    "heal action is non-vision",
			prompt:  promptHealAction("ACT", "DOM", "ERR"),
			musts:   []string{"ACT", "DOM", "ERR", "Analyze the DOM"},
			mustNot: []string{"screenshot"},
		},
		{
			name:   "heal action with vision mentions screenshot",
			prompt: promptHealActionWithVision("ACT", "DOM", "ERR"),
			musts:  []string{"ACT", "DOM", "ERR", "screenshot"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, m := range tc.musts {
				if !strings.Contains(tc.prompt, m) {
					t.Fatalf("prompt missing %q:\n%s", m, tc.prompt)
				}
			}
			for _, m := range tc.mustNot {
				if strings.Contains(tc.prompt, m) {
					t.Fatalf("prompt must not contain %q:\n%s", m, tc.prompt)
				}
			}
		})
	}
}

func TestParseTestPlan_StripsMarkersAndParses(t *testing.T) {
	raw := "```json\n{\"summary\":\"ok\",\"scenarios\":[{\"name\":\"a\",\"priority\":\"high\",\"steps\":[\"s1\"]}]}\n```"
	plan, err := parseTestPlan(raw)
	if err != nil {
		t.Fatalf("parseTestPlan: %v", err)
	}
	if plan.Summary != "ok" || len(plan.Scenarios) != 1 {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}

func TestParseTestFiles_InvalidJSONErrors(t *testing.T) {
	if _, err := parseTestFiles("not json"); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	files, err := parseTestFiles(`[{"name":"f.json","content":"[]"}]`)
	if err != nil {
		t.Fatalf("parseTestFiles: %v", err)
	}
	if len(files) != 1 || files[0].Name != "f.json" {
		t.Fatalf("unexpected files: %+v", files)
	}
}

func TestParseFixedFiles_RoundTrip(t *testing.T) {
	fixed, err := parseFixedFiles("```\n[{\"name\":\"x\",\"content\":\"y\"}]\n```")
	if err != nil {
		t.Fatalf("parseFixedFiles: %v", err)
	}
	if len(fixed) != 1 || fixed[0].Content != "y" {
		t.Fatalf("unexpected fixed: %+v", fixed)
	}
}
