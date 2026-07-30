package github

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhookHandler_VerifySignature(t *testing.T) {
	secret := "test-secret"
	handler := NewWebhookHandler(secret)

	tests := []struct {
		name      string
		payload   []byte
		signature string
		wantErr   bool
	}{
		{
			name:      "valid signature",
			payload:   []byte(`{"test":"data"}`),
			signature: computeSignature([]byte(`{"test":"data"}`), secret),
			wantErr:   false,
		},
		{
			name:      "invalid signature",
			payload:   []byte(`{"test":"data"}`),
			signature: "sha256=invalid",
			wantErr:   true,
		},
		{
			name:      "missing signature",
			payload:   []byte(`{"test":"data"}`),
			signature: "",
			wantErr:   true,
		},
		{
			name:      "wrong algorithm",
			payload:   []byte(`{"test":"data"}`),
			signature: "sha1=abc123",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.VerifySignature(tt.payload, tt.signature)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifySignature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWebhookHandler_VerifySignature_NoSecret(t *testing.T) {
	handler := NewWebhookHandler("")
	err := handler.VerifySignature([]byte(`{"test":"data"}`), "sha256=abc")
	if err == nil {
		t.Error("expected error when secret not configured")
	}
}

func TestWebhookHandler_ParsePushEvent(t *testing.T) {
	handler := NewWebhookHandler("test-secret")

	payload := `{
		"ref": "refs/heads/main",
		"before": "abc123",
		"after": "def456",
		"repository": {
			"full_name": "owner/repo",
			"clone_url": "https://github.com/owner/repo.git"
		},
		"commits": [
			{
				"id": "def456",
				"message": "test commit",
				"timestamp": "2024-01-01T00:00:00Z",
				"author": {"login": "testuser"},
				"added": ["file1.txt"],
				"removed": [],
				"modified": ["file2.txt"]
			}
		],
		"sender": {"login": "testuser"}
	}`

	event, err := handler.ParsePushEvent(json.RawMessage(payload))
	if err != nil {
		t.Fatalf("ParsePushEvent() error = %v", err)
	}

	if event.Ref != "refs/heads/main" {
		t.Errorf("expected ref 'refs/heads/main', got %s", event.Ref)
	}

	if event.Repository.FullName != "owner/repo" {
		t.Errorf("expected repo 'owner/repo', got %s", event.Repository.FullName)
	}

	if len(event.Commits) != 1 {
		t.Errorf("expected 1 commit, got %d", len(event.Commits))
	}

	if event.Commits[0].ID != "def456" {
		t.Errorf("expected commit ID 'def456', got %s", event.Commits[0].ID)
	}
}

func TestWebhookHandler_ParsePullRequestEvent(t *testing.T) {
	handler := NewWebhookHandler("test-secret")

	payload := `{
		"action": "opened",
		"number": 42,
		"pull_request": {
			"number": 42,
			"title": "Test PR",
			"state": "open",
			"html_url": "https://github.com/owner/repo/pull/42",
			"head": {
				"ref": "feature-branch",
				"sha": "abc123"
			},
			"base": {
				"ref": "main",
				"sha": "def456"
			},
			"user": {"login": "testuser"}
		},
		"repository": {
			"full_name": "owner/repo"
		},
		"sender": {"login": "testuser"}
	}`

	event, err := handler.ParsePullRequestEvent(json.RawMessage(payload))
	if err != nil {
		t.Fatalf("ParsePullRequestEvent() error = %v", err)
	}

	if event.Action != "opened" {
		t.Errorf("expected action 'opened', got %s", event.Action)
	}

	if event.Number != 42 {
		t.Errorf("expected PR number 42, got %d", event.Number)
	}

	if event.PullRequest.Title != "Test PR" {
		t.Errorf("expected title 'Test PR', got %s", event.PullRequest.Title)
	}

	if event.PullRequest.Head.Ref != "feature-branch" {
		t.Errorf("expected head ref 'feature-branch', got %s", event.PullRequest.Head.Ref)
	}
}

func TestWebhookHandler_HandleWebhook_Push(t *testing.T) {
	secret := "test-secret"
	handler := NewWebhookHandler(secret)

	payload := `{
		"ref": "refs/heads/main",
		"repository": {"full_name": "owner/repo"},
		"commits": [],
		"sender": {"login": "testuser"}
	}`

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString(payload))
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-Hub-Signature-256", computeSignature([]byte(payload), secret))

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["event"] != "push" {
		t.Errorf("expected event 'push', got %v", response["event"])
	}
}

func TestWebhookHandler_HandleWebhook_PullRequest(t *testing.T) {
	secret := "test-secret"
	handler := NewWebhookHandler(secret)

	payload := `{
		"action": "opened",
		"number": 42,
		"pull_request": {
			"number": 42,
			"title": "Test PR",
			"head": {"ref": "feature", "sha": "abc"},
			"base": {"ref": "main", "sha": "def"},
			"user": {"login": "testuser"}
		},
		"repository": {"full_name": "owner/repo"},
		"sender": {"login": "testuser"}
	}`

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString(payload))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-Hub-Signature-256", computeSignature([]byte(payload), secret))

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["event"] != "pull_request" {
		t.Errorf("expected event 'pull_request', got %v", response["event"])
	}
}

func TestWebhookHandler_HandleWebhook_Ping(t *testing.T) {
	handler := NewWebhookHandler("test-secret")

	payload := `{}`
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString(payload))
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-Hub-Signature-256", computeSignature([]byte(payload), "test-secret"))

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestWebhookHandler_HandleWebhook_MethodNotAllowed(t *testing.T) {
	handler := NewWebhookHandler("test-secret")

	req := httptest.NewRequest(http.MethodGet, "/webhook", nil)
	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestWebhookHandler_HandleWebhook_MissingEventHeader(t *testing.T) {
	handler := NewWebhookHandler("test-secret")

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString(`{}`))
	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestWebhookHandler_HandleWebhook_InvalidSignature(t *testing.T) {
	handler := NewWebhookHandler("test-secret")

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString(`{}`))
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-Hub-Signature-256", "sha256=invalid")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestWebhookHandler_HandleWebhook_UnsupportedEvent(t *testing.T) {
	handler := NewWebhookHandler("test-secret")

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString(`{}`))
	req.Header.Set("X-GitHub-Event", "unsupported")
	req.Header.Set("X-Hub-Signature-256", computeSignature([]byte(`{}`), "test-secret"))

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for unsupported event, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["status"] != "ignored" {
		t.Errorf("expected status 'ignored', got %v", response["status"])
	}
}

// computeSignature computes a valid HMAC-SHA256 signature for testing
func computeSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
