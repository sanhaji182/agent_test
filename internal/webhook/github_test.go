package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// Security-critical coverage: HMAC verification is the only barrier between
// the public webhook endpoint and auto-triggered test runs.

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func post(t *testing.T, h *GitHubHandler, event, body, signature string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/webhook/github", strings.NewReader(body))
	req.Header.Set("X-GitHub-Event", event)
	if signature != "" {
		req.Header.Set("X-Hub-Signature-256", signature)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

const pushBody = `{"ref":"refs/heads/main","repository":{"full_name":"o/r","clone_url":"https://x"},"head_commit":{"id":"abc","message":"m"}}`

func TestGitHub_RejectsNonPost(t *testing.T) {
	h := NewGitHubHandler("", nil)
	req := httptest.NewRequest(http.MethodGet, "/webhook/github", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestGitHub_ValidSignatureTriggersCallback(t *testing.T) {
	var mu sync.Mutex
	var got *PushEvent
	h := NewGitHubHandler("s3cret", func(e PushEvent) {
		mu.Lock()
		got = &e
		mu.Unlock()
	})

	rec := post(t, h, "push", pushBody, sign("s3cret", []byte(pushBody)))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	// Callback runs in a goroutine; poll briefly.
	deadline := time.Now().Add(2 * time.Second)
	for {
		mu.Lock()
		done := got != nil
		mu.Unlock()
		if done {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("onPush callback never fired")
		}
		time.Sleep(5 * time.Millisecond)
	}
	mu.Lock()
	defer mu.Unlock()
	if got.Ref != "refs/heads/main" || got.Repository.FullName != "o/r" || got.HeadCommit.ID != "abc" {
		t.Fatalf("unexpected event: %+v", got)
	}
}

func TestGitHub_InvalidSignatureRejected(t *testing.T) {
	called := false
	h := NewGitHubHandler("s3cret", func(PushEvent) { called = true })

	cases := map[string]string{
		"wrong secret":   sign("wrong", []byte(pushBody)),
		"missing prefix": strings.TrimPrefix(sign("s3cret", []byte(pushBody)), "sha256="),
		"empty header":   "",
		"garbage":        "sha256=zzzz",
		"tampered body":  sign("s3cret", []byte(pushBody+" ")),
	}
	for name, sig := range cases {
		rec := post(t, h, "push", pushBody, sig)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s: expected 401, got %d", name, rec.Code)
		}
	}
	time.Sleep(20 * time.Millisecond)
	if called {
		t.Fatal("callback fired despite invalid signature")
	}
}

func TestGitHub_NoSecretSkipsVerification(t *testing.T) {
	// Development mode: empty secret accepts unsigned requests.
	h := NewGitHubHandler("", nil)
	rec := post(t, h, "push", pushBody, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 without secret, got %d", rec.Code)
	}
}

func TestGitHub_InvalidJSONRejected(t *testing.T) {
	h := NewGitHubHandler("", nil)
	rec := post(t, h, "push", "not json", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid payload, got %d", rec.Code)
	}
}

func TestGitHub_PingAndUnknownEventsOK(t *testing.T) {
	h := NewGitHubHandler("", nil)
	if rec := post(t, h, "ping", `{}`, ""); rec.Code != http.StatusOK {
		t.Fatalf("ping: expected 200, got %d", rec.Code)
	}
	if rec := post(t, h, "issues", `{}`, ""); rec.Code != http.StatusOK {
		t.Fatalf("unknown event: expected 200, got %d", rec.Code)
	}
}
