package ai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/ai"
)

// newStubServer returns an OpenAI-compatible /chat/completions stub that
// captures the request payload and responds with the given content.
func newStubServer(t *testing.T, captured *map[string]interface{}, replyContent string, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		*captured = payload
		w.WriteHeader(status)
		if status >= 200 && status < 300 {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"choices": []map[string]interface{}{
					{"message": map[string]string{"content": replyContent}},
				},
			})
		} else {
			w.Write([]byte(`{"error":{"message":"boom"}}`))
		}
	}))
}

func newClient(t *testing.T, baseURL string) ai.Client {
	t.Helper()
	c := ai.New(ai.Config{Provider: "openai-compatible", Model: "test-model", APIKey: "k", BaseURL: baseURL})
	if c == nil {
		t.Fatal("ai.New returned nil")
	}
	return c
}

func TestOpenAICompatible_GenerateText(t *testing.T) {
	var captured map[string]interface{}
	srv := newStubServer(t, &captured, "hello back", http.StatusOK)
	defer srv.Close()

	got, err := newClient(t, srv.URL).GenerateText(context.Background(), "hello")
	if err != nil {
		t.Fatalf("GenerateText: %v", err)
	}
	if got != "hello back" {
		t.Fatalf("got %q, want %q", got, "hello back")
	}
	// Text-only: content must be a plain string, and max_tokens must be set.
	msgs := captured["messages"].([]interface{})
	content := msgs[0].(map[string]interface{})["content"]
	if _, ok := content.(string); !ok {
		t.Fatalf("text-only content should be string, got %T", content)
	}
	if captured["max_tokens"] == nil {
		t.Fatal("max_tokens missing from payload")
	}
	// Explicit stream:false — some OpenAI-compatible gateways default to SSE
	// streaming, which breaks JSON decoding (found in 2026-07-28 E2E smoke).
	if v, ok := captured["stream"].(bool); !ok || v {
		t.Fatalf("payload must carry stream:false, got %v", captured["stream"])
	}
}

func TestOpenAICompatible_GenerateWithImage(t *testing.T) {
	var captured map[string]interface{}
	srv := newStubServer(t, &captured, "vision reply", http.StatusOK)
	defer srv.Close()

	got, err := newClient(t, srv.URL).GenerateWithImage(context.Background(), "what is this", "aGVsbG8=")
	if err != nil {
		t.Fatalf("GenerateWithImage: %v", err)
	}
	if got != "vision reply" {
		t.Fatalf("got %q, want %q", got, "vision reply")
	}
	// Vision: content must be a block array with text + image_url parts.
	msgs := captured["messages"].([]interface{})
	blocks, ok := msgs[0].(map[string]interface{})["content"].([]interface{})
	if !ok || len(blocks) != 2 {
		t.Fatalf("vision content should be 2-block array, got %#v", msgs[0])
	}
	img := blocks[1].(map[string]interface{})
	if img["type"] != "image_url" {
		t.Fatalf("second block type = %v, want image_url", img["type"])
	}
	url := img["image_url"].(map[string]interface{})["url"].(string)
	if !strings.HasPrefix(url, "data:image/jpeg;base64,aGVsbG8=") {
		t.Fatalf("unexpected data URL %q", url)
	}
}

func TestOpenAICompatible_GenerateWithImage_EmptyImageFallsBackToText(t *testing.T) {
	var captured map[string]interface{}
	srv := newStubServer(t, &captured, "plain", http.StatusOK)
	defer srv.Close()

	if _, err := newClient(t, srv.URL).GenerateWithImage(context.Background(), "prompt", ""); err != nil {
		t.Fatalf("GenerateWithImage(empty image): %v", err)
	}
	msgs := captured["messages"].([]interface{})
	if _, ok := msgs[0].(map[string]interface{})["content"].(string); !ok {
		t.Fatal("empty-image call should send plain string content")
	}
}

func TestOpenAICompatible_ErrorIncludesBody(t *testing.T) {
	var captured map[string]interface{}
	srv := newStubServer(t, &captured, "", http.StatusUnauthorized)
	defer srv.Close()

	_, err := newClient(t, srv.URL).GenerateText(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error on 401")
	}
	if !strings.Contains(err.Error(), "status 401") || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("error should include status and body, got %q", err.Error())
	}
}

func TestStripJSONMarkers(t *testing.T) {
	in := "```json\n{\"a\":1}\n```"
	if got := ai.StripJSONMarkers(in); got != `{"a":1}` {
		t.Fatalf("StripJSONMarkers = %q", got)
	}
}
