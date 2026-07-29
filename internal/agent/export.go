package agent

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ExportPlaywrightScript converts a TestFile's action JSON into a runnable
// Playwright TypeScript test file. This lets users take AI-generated actions
// and run them independently in their own Playwright setup.
func ExportPlaywrightScript(tf TestFile, options ExportOptions) string {
	var actions []BrowserAction
	if err := json.Unmarshal([]byte(tf.Content), &actions); err != nil {
		return fmt.Sprintf("// Error: could not parse actions: %v\n", err)
	}

	var sb strings.Builder
	testName := strings.TrimSuffix(tf.Name, ".json")

	sb.WriteString(`import { test, expect } from '@playwright/test';

`)
	sb.WriteString(fmt.Sprintf("test('%s', async ({ page }) => {\n", testName))

	for _, a := range actions {
		sb.WriteString(actionToPlaywrightCode(a, options))
	}

	sb.WriteString("});\n")
	return sb.String()
}

// ExportOptions configures code generation style.
type ExportOptions struct {
	Timeout    int    // default timeout in ms (default 5000)
	Language   string // "typescript" (default) or "javascript"
	AddWaits   bool   // add waitForLoadState after navigation
	IndentWith string // "  " (default) or "\t"
}

func actionToPlaywrightCode(a BrowserAction, opts ExportOptions) string {
	indent := opts.IndentWith
	if indent == "" {
		indent = "  "
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 5000
	}

	switch a.Action {
	case "goto":
		code := fmt.Sprintf("%sawait page.goto('%s');\n", indent, escapeJS(a.URL))
		if opts.AddWaits {
			code += fmt.Sprintf("%sawait page.waitForLoadState('networkidle');\n", indent)
		}
		return code
	case "fill":
		return fmt.Sprintf("%sawait page.locator('%s').fill('%s', { timeout: %d });\n",
			indent, escapeJS(a.Selector), escapeJS(a.Value), timeout)
	case "click":
		return fmt.Sprintf("%sawait page.locator('%s').first().click({ timeout: %d });\n",
			indent, escapeJS(a.Selector), timeout)
	case "hover":
		return fmt.Sprintf("%sawait page.locator('%s').first().hover({ timeout: %d });\n",
			indent, escapeJS(a.Selector), timeout)
	case "press":
		return fmt.Sprintf("%sawait page.locator('%s').first().press('%s');\n",
			indent, escapeJS(a.Selector), escapeJS(a.Key))
	case "scroll":
		return fmt.Sprintf("%sawait page.evaluate(() => window.scrollBy(0, %d));\n",
			indent, a.Y)
	case "wait":
		return fmt.Sprintf("%sawait page.waitForTimeout(%d);\n", indent, a.Ms)
	case "screenshot":
		return fmt.Sprintf("%sawait page.screenshot({ path: 'screenshot.png', fullPage: true });\n", indent)
	case "network_wait":
		return fmt.Sprintf("%sawait page.waitForResponse(resp => resp.url().includes('%s'), { timeout: %d });\n",
			indent, escapeJS(a.NetworkURL), max(a.Ms, 10000))
	case "assert":
		return assertToPlaywrightCode(a, indent, timeout)
	default:
		return fmt.Sprintf("%s// Unknown action: %s\n", indent, a.Action)
	}
}

func assertToPlaywrightCode(a BrowserAction, indent string, timeout int) string {
	switch a.Assert {
	case "visible":
		return fmt.Sprintf("%sawait expect(page.locator('%s')).toBeVisible({ timeout: %d });\n",
			indent, escapeJS(a.Selector), timeout)
	case "hidden":
		return fmt.Sprintf("%sawait expect(page.locator('%s')).toBeHidden({ timeout: %d });\n",
			indent, escapeJS(a.Selector), timeout)
	case "text_contains":
		return fmt.Sprintf("%sawait expect(page.locator('%s')).toContainText('%s', { timeout: %d });\n",
			indent, escapeJS(a.Selector), escapeJS(a.Text), timeout)
	case "url_contains":
		return fmt.Sprintf("%sawait expect(page).toHaveURL(/%s/);\n",
			indent, escapeRegex(a.Text))
	case "title_contains":
		return fmt.Sprintf("%sawait expect(page).toHaveTitle(/%s/);\n",
			indent, escapeRegex(a.Text))
	case "count":
		return fmt.Sprintf("%sawait expect(page.locator('%s')).toHaveCount(%s);\n",
			indent, escapeJS(a.Selector), a.Text)
	case "attribute":
		return fmt.Sprintf("%sawait expect(page.locator('%s')).toHaveAttribute('%s', '%s');\n",
			indent, escapeJS(a.Selector), escapeJS(a.Key), escapeJS(a.Text))
	default:
		return fmt.Sprintf("%s// Unknown assert: %s\n", indent, a.Assert)
	}
}

// ExportAllScripts converts all test files to a map of filename → code.
func ExportAllScripts(testFiles []TestFile, opts ExportOptions) map[string]string {
	result := make(map[string]string, len(testFiles))
	for _, tf := range testFiles {
		name := strings.TrimSuffix(tf.Name, ".json") + ".spec.ts"
		result[name] = ExportPlaywrightScript(tf, opts)
	}
	return result
}

func escapeJS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

func escapeRegex(s string) string {
	special := []string{"\\", ".", "*", "+", "?", "(", ")", "[", "]", "{", "}", "^", "$", "|", "/"}
	for _, c := range special {
		s = strings.ReplaceAll(s, c, "\\"+c)
	}
	return s
}
