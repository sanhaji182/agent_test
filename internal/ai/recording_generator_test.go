package ai

import (
	"context"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/recordings"
)

func TestGeneratePlaywrightSkeleton(t *testing.T) {
	sess := &recordings.Session{
		Name:    "Login Flow",
		BaseURL: "https://example.com",
	}
	events := []recordings.Event{
		{EventType: recordings.EventNavigate, URL: "https://example.com/login"},
		{EventType: recordings.EventFill, Selector: "[name='email']", Value: "user@test.com"},
		{EventType: recordings.EventFill, Selector: "[name='password']", Value: "secret"},
		{EventType: recordings.EventClick, Selector: "[type='submit']"},
		{EventType: recordings.EventAssertText, Selector: ".welcome", Value: "Welcome"},
	}

	code := generatePlaywrightSkeleton(sess, events)
	if code == "" {
		t.Fatal("expected non-empty code")
	}
	if !contains(code, "import { test, expect }") {
		t.Error("missing import statement")
	}
	if !contains(code, "test('Login Flow'") {
		t.Error("missing test name")
	}
	if !contains(code, "page.goto('https://example.com/login'") {
		t.Error("missing navigate action")
	}
	if !contains(code, "page.fill") {
		t.Error("missing fill action")
	}
	if !contains(code, "page.click") {
		t.Error("missing click action")
	}
}

func TestGeneratePlaywrightSkeleton_EmptyEvents(t *testing.T) {
	sess := &recordings.Session{Name: "Empty", BaseURL: "https://test.com"}
	code := generatePlaywrightSkeleton(sess, nil)
	if code == "" {
		t.Fatal("expected code even with empty events")
	}
	if !contains(code, "page.goto('https://test.com'") {
		t.Error("expected base URL navigation")
	}
}

func TestRecordingGenerator_WithoutLLM(t *testing.T) {
	gen := NewRecordingGenerator(nil) // No LLM client → should use skeleton
	sess := &recordings.Session{Name: "Test", BaseURL: "https://test.com"}
	events := []recordings.Event{
		{EventType: recordings.EventClick, Selector: "#btn"},
	}
	code, err := gen.GeneratePlaywrightTest(context.Background(), sess, events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code == "" {
		t.Fatal("expected generated code")
	}
}

type mockAIClient struct {
	text string
	err  error
}

func (m *mockAIClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	return m.text, m.err
}

func (m *mockAIClient) GenerateWithImage(ctx context.Context, prompt, imageBase64 string) (string, error) {
	return m.text, m.err
}

func TestRecordingGenerator_WithLLM(t *testing.T) {
	llm := &mockAIClient{
		text: "```typescript\nimport { test } from '@playwright/test';\ntest('generated', async ({ page }) => {});\n```",
	}
	gen := NewRecordingGenerator(llm)
	sess := &recordings.Session{Name: "Test", BaseURL: "https://test.com"}
	events := []recordings.Event{{EventType: recordings.EventClick, Selector: "#btn"}}
	code, err := gen.GeneratePlaywrightTest(context.Background(), sess, events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(code, "import { test }") {
		t.Errorf("expected LLM-generated code, got: %s", code)
	}
}
