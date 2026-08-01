package ai

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-go-golems/gotest-agent/internal/recordings"
)

// RecordingGenerator converts recorded session events into Playwright test code.
type RecordingGenerator struct {
	client Client
}

// NewRecordingGenerator creates a generator backed by an LLM client.
func NewRecordingGenerator(client Client) *RecordingGenerator {
	return &RecordingGenerator{client: client}
}

// GeneratePlaywrightTest produces Playwright TypeScript code from a recorded session.
// Falls back to a skeleton generator if no LLM client is available.
func (g *RecordingGenerator) GeneratePlaywrightTest(ctx context.Context, session *recordings.Session, events []recordings.Event) (string, error) {
	if g.client != nil {
		return g.generateWithLLM(ctx, session, events)
	}
	return generatePlaywrightSkeleton(session, events), nil
}

func (g *RecordingGenerator) generateWithLLM(ctx context.Context, session *recordings.Session, events []recordings.Event) (string, error) {
	prompt := buildRecordingPrompt(session, events)
	out, err := g.client.GenerateText(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("llm generation failed: %w", err)
	}
	code := extractCodeBlock(out)
	if strings.TrimSpace(code) == "" {
		return out, nil // return raw output if no code block found
	}
	return code, nil
}

func buildRecordingPrompt(session *recordings.Session, events []recordings.Event) string {
	var b strings.Builder
	b.WriteString("You are a senior test automation engineer. Convert the following recorded user interactions into a Playwright test.\n\n")
	fmt.Fprintf(&b, "Test name: %s\n", session.Name)
	fmt.Fprintf(&b, "Base URL: %s\n\n", session.BaseURL)
	b.WriteString("Recorded events (in order):\n")
	for i, ev := range events {
		fmt.Fprintf(&b, "%d. %s", i+1, ev.EventType)
		if ev.Selector != "" {
			fmt.Fprintf(&b, " on '%s'", ev.Selector)
		}
		if ev.Value != "" {
			fmt.Fprintf(&b, " with value '%s'", ev.Value)
		}
		if ev.URL != "" && ev.EventType == recordings.EventNavigate {
			fmt.Fprintf(&b, " to '%s'", ev.URL)
		}
		b.WriteString("\n")
	}
	b.WriteString("\nGenerate a complete Playwright TypeScript test file that reproduces these interactions. ")
	b.WriteString("Use data-testid selectors where possible. Include proper assertions at key steps. ")
	b.WriteString("Return ONLY the test code inside one fenced code block.\n")
	return b.String()
}

func generatePlaywrightSkeleton(session *recordings.Session, events []recordings.Event) string {
	var b strings.Builder
	b.WriteString("import { test, expect } from '@playwright/test';\n\n")
	fmt.Fprintf(&b, "test('%s', async ({ page }) => {\n", session.Name)
	fmt.Fprintf(&b, "  await page.goto('%s');\n\n", session.BaseURL)
	for _, ev := range events {
		switch ev.EventType {
		case recordings.EventClick:
			if ev.Selector != "" {
				fmt.Fprintf(&b, "  await page.click('%s');\n", sanitizeForTS(ev.Selector))
			}
		case recordings.EventFill:
			if ev.Selector != "" {
				fmt.Fprintf(&b, "  await page.fill('%s', '%s');\n", sanitizeForTS(ev.Selector), sanitizeForTS(ev.Value))
			}
		case recordings.EventNavigate:
			if ev.URL != "" {
				fmt.Fprintf(&b, "  await page.goto('%s');\n", sanitizeForTS(ev.URL))
			}
		case recordings.EventSelect:
			if ev.Selector != "" && ev.Value != "" {
				fmt.Fprintf(&b, "  await page.selectOption('%s', '%s');\n", sanitizeForTS(ev.Selector), sanitizeForTS(ev.Value))
			}
		case recordings.EventHover:
			if ev.Selector != "" {
				fmt.Fprintf(&b, "  await page.hover('%s');\n", sanitizeForTS(ev.Selector))
			}
		case recordings.EventPress:
			if ev.Value != "" {
				fmt.Fprintf(&b, "  await page.keyboard.press('%s');\n", sanitizeForTS(ev.Value))
			}
		case recordings.EventScroll:
			b.WriteString("  await page.evaluate(() => window.scrollBy(0, 300));\n")
		case recordings.EventAssertText:
			if ev.Selector != "" && ev.Value != "" {
				fmt.Fprintf(&b, "  await expect(page.locator('%s')).toContainText('%s');\n", sanitizeForTS(ev.Selector), sanitizeForTS(ev.Value))
			}
		case recordings.EventAssertVisible:
			if ev.Selector != "" {
				fmt.Fprintf(&b, "  await expect(page.locator('%s')).toBeVisible();\n", sanitizeForTS(ev.Selector))
			}
		}
	}
	b.WriteString("});\n")
	return b.String()
}

var codeBlockRe = regexp.MustCompile("(?s)```[a-zA-Z0-9_+-]*\\n(.*?)```")

func extractCodeBlock(s string) string {
	if m := codeBlockRe.FindStringSubmatch(s); m != nil {
		return strings.TrimSpace(m[1])
	}
	return strings.TrimSpace(s)
}

func sanitizeForTS(s string) string {
	return strings.ReplaceAll(s, "'", "\\'")
}
