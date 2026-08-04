package ai

import (
	"fmt"
	"strings"
	"testing"
)

// TestExtractProviderError_NestedProxyMessage uses the real shape returned by the
// OpenAI-compatible proxy: the upstream provider error is stringified and embedded
// inside error.message. We must surface the innermost human-readable message.
func TestExtractProviderError_NestedProxyMessage(t *testing.T) {
	body := []byte(`{"error":{"message":"[openai-compatible-chat-x/claude-sonnet-4.6] [402]: {\"error\":\"request_failed\",\"message\":\"Insufficient balance. Please top up your credits or upgrade your plan.\"}\n (reset after 2m)"}}`)
	want := "Insufficient balance. Please top up your credits or upgrade your plan."
	if got := extractProviderError(body); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestExtractProviderError_StandardEnvelope(t *testing.T) {
	body := []byte(`{"error":{"message":"Invalid API key provided."}}`)
	if got := extractProviderError(body); got != "Invalid API key provided." {
		t.Fatalf("got %q", got)
	}
}

func TestExtractProviderError_HTML(t *testing.T) {
	body := []byte(`<html><body>502 Bad Gateway</body></html>`)
	got := extractProviderError(body)
	if !strings.Contains(got, "HTML") {
		t.Fatalf("expected an HTML hint, got %q", got)
	}
}

func TestExtractProviderError_Empty(t *testing.T) {
	if got := extractProviderError([]byte("   ")); got != "empty response body" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractProviderError_Unrecognized(t *testing.T) {
	body := []byte(`{"foo":"bar"}`)
	if got := extractProviderError(body); got != `{"foo":"bar"}` {
		t.Fatalf("got %q", got)
	}
}

// TestCleanAnthropicError simulates the anthropic-sdk-go error string format and
// verifies we surface the provider's actual message with the status code.
func TestCleanAnthropicError(t *testing.T) {
	err := fmt.Errorf(`POST "https://api.anthropic.com/v1/messages": 402 Payment Required {"type":"error","error":{"type":"invalid_request_error","message":"Your credit balance is too low to access the Anthropic API."}}`)
	want := "status 402: Your credit balance is too low to access the Anthropic API."
	if got := cleanAnthropicError(err); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestCleanAnthropicError_NoBody(t *testing.T) {
	err := fmt.Errorf("context canceled")
	if got := cleanAnthropicError(err); got != "context canceled" {
		t.Fatalf("got %q", got)
	}
}

func TestCleanAnthropicError_InvalidKey(t *testing.T) {
	err := fmt.Errorf(`POST "https://api.anthropic.com/v1/messages": 401 Unauthorized {"type":"error","error":{"type":"authentication_error","message":"invalid x-api-key"}}`)
	want := "status 401: invalid x-api-key"
	if got := cleanAnthropicError(err); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
