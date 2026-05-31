package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type GitHubHandler struct {
	secret    string
	onPush    func(event PushEvent)
}

type PushEvent struct {
	Ref        string     `json:"ref"`
	Repository Repository `json:"repository"`
	HeadCommit Commit     `json:"head_commit"`
}

type Repository struct {
	FullName string `json:"full_name"`
	CloneURL string `json:"clone_url"`
}

type Commit struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

func NewGitHubHandler(secret string, onPush func(PushEvent)) *GitHubHandler {
	return &GitHubHandler{secret: secret, onPush: onPush}
}

func (h *GitHubHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}

	if h.secret != "" {
		sig := r.Header.Get("X-Hub-Signature-256")
		if !h.verifySignature(body, sig) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
	}

	event := r.Header.Get("X-GitHub-Event")
	switch event {
	case "push":
		var push PushEvent
		if err := json.Unmarshal(body, &push); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		slog.Info("github push received", "repo", push.Repository.FullName, "ref", push.Ref)
		if h.onPush != nil {
			go h.onPush(push)
		}
	case "ping":
		slog.Info("github ping received")
	default:
		slog.Info("github event ignored", "event", event)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ok"}`)
}

func (h *GitHubHandler) verifySignature(body []byte, signature string) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	sig := strings.TrimPrefix(signature, "sha256=")
	mac := hmac.New(sha256.New, []byte(h.secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(expected))
}
