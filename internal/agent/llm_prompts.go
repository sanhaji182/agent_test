package agent

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
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
	return fmt.Sprintf(`Anda adalah API generasi rencana tes. Anda TIDAK memiliki alat apa pun. JANGAN jalankan perintah, JANGAN gunakan bash atau shell, JANGAN eksplorasi codebase, JANGAN tulis penjelasan, pemikiran, atau prosa. Balas dengan SATU objek JSON SAJA.

Konteks analisis codebase:
%s

Persyaratan: %s

INGAT: SEMUA output teks (summary, nama skenario, langkah-langkah) HARUS dalam BAHASA INDONESIA. Struktur JSON harus persis seperti ini:
{"summary": "...", "scenarios": [{"name": "...", "priority": "high|medium|low", "steps": ["..."]}]}

CRITICAL OUTPUT RULES:
- Your ENTIRE response must be valid JSON only.
- It must start with { and end with }.
- No markdown fences, no prose, no tool calls, no <bash> tags, no thinking, no commentary.
- Do not attempt to inspect or list files; use only the context provided above.
- Return ONLY valid JSON.`, analysis, requirements)
}

func promptGenerateTestScripts(plan *TestPlan, analysis string) string {
	planJSON, _ := json.Marshal(plan)
	return fmt.Sprintf(`You are a Playwright test-generation API. You have NO tools available. Do NOT run commands, do NOT use bash or any shell, do NOT explore any codebase, do NOT write prose, thinking, or explanations. Respond with a JSON array ONLY.

TEST PLAN:
%s

CODEBASE CONTEXT:
%s

INSTRUCTIONS:
Generate comprehensive Playwright test scripts as JSON array. Each test must:

1. **Use robust, semantic selectors** (prefer: data-testid > tag semantik (h1,h2,button,nav,input,form) > aria-label > class):
   - ✅ "h1" atau "h2" untuk heading (situses biasanya pakai tag <h1> standar, BUKAN atribut role/aria-level)
   - ✅ "button[data-testid='submit']" atau "button:has-text('Submit')"
   - ✅ "input[type='email']" atau "input[name='email']"
   - ❌ ".btn.btn-primary.submit" (fragile)
   - ❌ "[role='heading'][aria-level='1']" (JANGAN dipakai untuk heading; atribut role/aria-level hampir tidak pernah ditulis eksplisit di HTML modern)
   - Use robust selectors: semantic tags and data-testid beat fragile class chains.

2. **Add strategic waits** after navigation/actions:
   - {"action": "wait", "ms": 1000} after page loads
   - {"action": "wait", "ms": 500} after form submissions

3. **Verify outcomes** with scroll/checkpoint actions:
   - Scroll to verify success messages appear
   - Wait to ensure async operations complete

4. **Handle common patterns**:
   - Login flows: fill email → fill password → click submit → wait → verify redirect
   - Form submissions: fill fields → submit → wait → verify success message
   - Navigation: goto → wait → verify page loaded

SUPPORTED ACTIONS:
- {"action": "goto", "url": "https://..."}
- {"action": "fill", "selector": "...", "value": "..."}
- {"action": "click", "selector": "..."}
- {"action": "scroll", "y": 500}
- {"action": "wait", "ms": 2000}
- {"action": "hover", "selector": "..."}
- {"action": "press", "selector": "...", "key": "Enter"}
- {"action": "assert", "selector": "...", "assert": "visible"} // element must be visible
- {"action": "assert", "selector": "...", "assert": "hidden"} // element must not be visible
- {"action": "assert", "selector": "...", "assert": "text_contains", "text": "expected text"}
- {"action": "assert", "selector": "...", "assert": "url_contains", "text": "/dashboard"} // URL must contain text
- {"action": "assert", "selector": "...", "assert": "title_contains", "text": "Home"} // page title must contain text
- {"action": "screenshot"} // capture screenshot as evidence

OUTPUT FORMAT (JSON array):
[
  {
    "name": "test_scenario_1.json",
    "content": "[\n  {\"action\": \"goto\", \"url\": \"https://app.example.com/login\"},\n  {\"action\": \"wait\", \"ms\": 1000},\n  {\"action\": \"fill\", \"selector\": \"input[type='email']\", \"value\": \"test@example.com\"},\n  {\"action\": \"fill\", \"selector\": \"input[type='password']\", \"value\": \"password123\"},\n  {\"action\": \"click\", \"selector\": \"button[type='submit']\"},\n  {\"action\": \"wait\", \"ms\": 2000},\n  {\"action\": \"scroll\", \"y\": 0}\n]"
  }
]

Generate tests that are:
- ✅ Resilient to UI changes (use semantic selectors)
- ✅ Fast (minimal unnecessary waits)
- ✅ Verifiable (use assert actions to check outcomes, not just perform actions)
- ✅ Self-documenting (clear action sequences)

HEADING & ASSERTIONS: untuk memverifikasi heading utama, gunakan selector "h1" (atau "h1:has-text('...')" jika tahu teksnya). Jangan gunakan atribut role/aria-level untuk heading.

Return ONLY valid JSON.

IMPORTANT: Every test should end with at least one assert action to verify the expected outcome.

CRITICAL OUTPUT RULES:
- Your ENTIRE response must be a valid JSON array only.
- It must start with [ and end with ].
- No markdown fences, no prose, no tool calls, no <bash> tags, no thinking, no commentary.
- Do not attempt to inspect or list files; use only the context provided above.`, string(planJSON), analysis)
}

func promptSuggestFixes(failures []Failure, files []TestFile) string {
	failJSON, _ := json.Marshal(failures)
	filesJSON, _ := json.Marshal(files)
	return fmt.Sprintf(`You are a test-fixing API. You have NO tools available. Do NOT run commands, do NOT use bash or any shell, do NOT explore any codebase, do NOT write prose, thinking, or explanations. Respond with a JSON array ONLY.

These Playwright tests failed:
Failures: %s
Original files: %s

Fix the test files. Output a JSON array with EXACTLY this structure:
[{"name": "...", "content": "..."}]

CRITICAL OUTPUT RULES:
- Your ENTIRE response must be a valid JSON array only.
- It must start with [ and end with ].
- No markdown fences, no prose, no tool calls, no <bash> tags, no thinking, no commentary.
- Inside "content", escape newlines as \n (never use literal newlines).`, string(failJSON), string(filesJSON))
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

CRITICAL: You have NO tools. Do NOT run commands, do NOT use bash, do NOT explore, do NOT write prose or thinking. Your ENTIRE response must be a single valid JSON object only (start with { and end with }). No markdown, no tool calls, no <bash> tags, no explanation.`, action, errorMsg, domSnapshot)
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

CRITICAL: You have NO tools. Do NOT run commands, do NOT use bash, do NOT explore, do NOT write prose or thinking. Your ENTIRE response must be a single valid JSON object only (start with { and end with }). No markdown, no tool calls, no <bash> tags, no explanation.`, action, errorMsg, domSnapshot)
}

// --- Response parsers ---

// stripJSONMarkers membersihkan markdown code fence dari response LLM
// Contoh: ```json\n{...}\n``` → {...}
func stripJSONMarkers(s string) string {
	s = strings.TrimPrefix(s, "```json\n")
	s = strings.TrimPrefix(s, "```\n")
	s = strings.TrimSuffix(s, "\n```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)
	// Some OpenAI-compatible gateways append an SSE terminator even on
	// non-streamed responses; drop it so JSON decoding succeeds.
	if i := strings.Index(s, "data: [DONE]"); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	return sanitizeJSONControlChars(extractJSONValue(s))
}

// sanitizeJSONControlChars meng-escape karakter kontrol literal (newline, tab,
// carriage return) yang muncul DI DALAM nilai string JSON. Beberapa LLM
// menghasilkan field "content" berisi file test dengan newline asli (bukan \n),
// yang membuat JSON tidak valid; fungsi ini membuatnya bisa di-parse tanpa
// mengubah whitespace struktural di luar string.
func sanitizeJSONControlChars(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inStr := false
	escaped := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inStr {
			if escaped {
				b.WriteByte(c)
				escaped = false
				continue
			}
			switch c {
			case '\\':
				b.WriteByte(c)
				escaped = true
			case '"':
				b.WriteByte(c)
				inStr = false
			case '\n':
				b.WriteString(`\n`)
			case '\r':
				b.WriteString(`\r`)
			case '\t':
				b.WriteString(`\t`)
			default:
				b.WriteByte(c)
			}
			continue
		}
		if c == '"' {
			inStr = true
		}
		b.WriteByte(c)
	}
	return b.String()
}

// extractJSONValue mengisolasi JSON object/array pertama di s, melewati
// prose/HTML di awal dan garbage di akhir yang mungkin dihasilkan model.
func extractJSONValue(s string) string {
	start := strings.IndexAny(s, "{[")
	if start < 0 {
		return s
	}
	open := s[start]
	closing := byte('}')
	if open == '[' {
		closing = ']'
	}
	depth := 0
	inStr := false
	esc := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if esc {
			esc = false
			continue
		}
		if inStr {
			switch c {
			case '\\':
				esc = true
			case '"':
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case open:
			depth++
		case closing:
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return s[start:]
}

func parseTestPlan(raw string) (*TestPlan, error) {
	raw = stripJSONMarkers(raw)
	var plan TestPlan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		preview := raw
		if len(preview) > 500 {
			preview = preview[:500]
		}
		slog.Warn("parse test plan failed", "error", err, "raw_preview", preview)
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
