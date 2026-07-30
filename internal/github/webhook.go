package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// WebhookEvent represents a GitHub webhook event
type WebhookEvent struct {
	Type    string          `json:"-"` // Set from X-GitHub-Event header
	Payload json.RawMessage `json:"-"` // Raw JSON payload
}

// PushEvent represents a push webhook event
type PushEvent struct {
	Ref        string     `json:"ref"`
	Before     string     `json:"before"`
	After      string     `json:"after"`
	Repository Repository `json:"repository"`
	Commits    []Commit   `json:"commits"`
	Sender     User       `json:"sender"`
}

// Commit represents a commit in a push event
type Commit struct {
	ID        string   `json:"id"`
	Message   string   `json:"message"`
	Timestamp string   `json:"timestamp"`
	Author    User     `json:"author"`
	Added     []string `json:"added"`
	Removed   []string `json:"removed"`
	Modified  []string `json:"modified"`
}

// PullRequestEvent represents a pull request webhook event
type PullRequestEvent struct {
	Action      string      `json:"action"`
	Number      int         `json:"number"`
	PullRequest PullRequest `json:"pull_request"`
	Repository  Repository  `json:"repository"`
	Sender      User        `json:"sender"`
}

// PullRequest represents a pull request
type PullRequest struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	State     string `json:"state"`
	HTMLURL   string `json:"html_url"`
	Head      Ref    `json:"head"`
	Base      Ref    `json:"base"`
	Merged    bool   `json:"merged"`
	MergedBy  *User  `json:"merged_by"`
	User      User   `json:"user"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Ref represents a git reference (branch)
type Ref struct {
	Ref  string     `json:"ref"`
	SHA  string     `json:"sha"`
	Repo Repository `json:"repo"`
}

// WebhookHandler processes GitHub webhook events
type WebhookHandler struct {
	secret string
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(secret string) *WebhookHandler {
	return &WebhookHandler{secret: secret}
}

// VerifySignature verifies the GitHub webhook signature
func (h *WebhookHandler) VerifySignature(payload []byte, signature string) error {
	if h.secret == "" {
		return fmt.Errorf("webhook secret not configured")
	}

	if signature == "" {
		return fmt.Errorf("missing signature")
	}

	// GitHub uses sha256 with HMAC
	parts := strings.SplitN(signature, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid signature format")
	}

	algorithm := parts[0]
	sig := parts[1]

	if algorithm != "sha256" {
		return fmt.Errorf("unsupported signature algorithm: %s", algorithm)
	}

	// Compute expected signature
	mac := hmac.New(sha256.New, []byte(h.secret))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)
	expectedSig := hex.EncodeToString(expectedMAC)

	// Compare signatures
	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// ParseEvent parses a webhook request into a WebhookEvent
func (h *WebhookHandler) ParseEvent(r *http.Request) (*WebhookEvent, error) {
	// Get event type from header
	eventType := r.Header.Get("X-GitHub-Event")
	if eventType == "" {
		return nil, fmt.Errorf("missing X-GitHub-Event header")
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	defer r.Body.Close()

	// Verify signature if secret is configured
	if h.secret != "" {
		signature := r.Header.Get("X-Hub-Signature-256")
		if err := h.VerifySignature(body, signature); err != nil {
			return nil, fmt.Errorf("signature verification failed: %w", err)
		}
	}

	return &WebhookEvent{
		Type:    eventType,
		Payload: json.RawMessage(body),
	}, nil
}

// ParsePushEvent parses a push event payload
func (h *WebhookHandler) ParsePushEvent(payload json.RawMessage) (*PushEvent, error) {
	var event PushEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("failed to parse push event: %w", err)
	}
	return &event, nil
}

// ParsePullRequestEvent parses a pull request event payload
func (h *WebhookHandler) ParsePullRequestEvent(payload json.RawMessage) (*PullRequestEvent, error) {
	var event PullRequestEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("failed to parse pull request event: %w", err)
	}
	return &event, nil
}

// HandleWebhook is an HTTP handler that processes GitHub webhooks
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	event, err := h.ParseEvent(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Process based on event type
	switch event.Type {
	case "push":
		pushEvent, err := h.ParsePushEvent(event.Payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.handlePush(w, pushEvent)

	case "pull_request":
		prEvent, err := h.ParsePullRequestEvent(event.Payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.handlePullRequest(w, prEvent)

	case "ping":
		// GitHub sends ping events to verify webhook URL
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))

	default:
		// Ignore unsupported events
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ignored"}`))
	}
}

// handlePush processes push events
func (h *WebhookHandler) handlePush(w http.ResponseWriter, event *PushEvent) {
	// TODO: Implement push event processing
	// - Clone/pull repository
	// - Analyze changed files
	// - Trigger test generation/execution

	response := map[string]interface{}{
		"status":  "received",
		"event":   "push",
		"repo":    event.Repository.FullName,
		"ref":     event.Ref,
		"commits": len(event.Commits),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handlePullRequest processes pull request events
func (h *WebhookHandler) handlePullRequest(w http.ResponseWriter, event *PullRequestEvent) {
	// TODO: Implement PR event processing
	// - Clone/pull repository
	// - Checkout PR branch
	// - Run tests on PR changes
	// - Post results as PR comment

	response := map[string]interface{}{
		"status":  "received",
		"event":   "pull_request",
		"action":  event.Action,
		"repo":    event.Repository.FullName,
		"pr":      event.PullRequest.Number,
		"title":   event.PullRequest.Title,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
