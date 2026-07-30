# Phase 3 Implementation Guide: Continuous Sync & Drift Detection

**Status**: Ready for Implementation  
**Date**: July 31, 2026  
**Goal**: Implement continuous synchronization between code changes and test updates

---

## Overview

Phase 3 focuses on making the testing system truly continuous and self-maintaining. The system will automatically detect code changes, identify drift between code and tests, and regenerate/update tests as needed.

### Key Features
1. **GitHub Webhook Integration** - Receive real-time notifications when code changes
2. **Webhook Registration Service** - Allow users to register repositories for monitoring
3. **Drift Detection Service** - Detect drift between code and tests
4. **Auto-Generation Service** - Automatically generate tests for drifts
5. **Alert System** - Notify teams when tests are outdated or failing

---

## Architecture

### Components

```
┌─────────────────────┐
│  GitHub Webhook     │
│  (Push Events)      │
└──────────┬──────────┘
           │ POST /api/v1/webhooks/github
           ▼
┌─────────────────────┐
│  Webhook Handler    │
│  - Verify HMAC      │
│  - Parse payload    │
│  - Extract changes  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Drift Detector     │
│  - Filter files     │
│  - Parse codebase   │
│  - Analyze drift    │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Auto-Generator     │
│  - Generate tests   │
│  - Store results    │
│  - Update status    │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Alert System       │
│  - Notify team      │
│  - Send reports     │
└─────────────────────┘
```

### Database Schema

```sql
-- Webhook registrations
CREATE TABLE webhook_registrations (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    repository_id UUID NOT NULL,
    repository_url VARCHAR(500) NOT NULL,
    webhook_id VARCHAR(255) NOT NULL, -- GitHub webhook ID
    status VARCHAR(50) DEFAULT 'active', -- active, inactive, error
    last_sync_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_webhook_registrations_user ON webhook_registrations(user_id);
CREATE INDEX idx_webhook_registrations_repo ON webhook_registrations(repository_url);

-- Drift records
CREATE TABLE drifts (
    id UUID PRIMARY KEY,
    repository_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL, -- missing_test, outdated_test, removed_test
    file_path VARCHAR(500) NOT NULL,
    description TEXT NOT NULL,
    severity VARCHAR(50) NOT NULL, -- high, medium, low
    status VARCHAR(50) DEFAULT 'pending', -- pending, fixed, ignored
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_drifts_repository ON drifts(repository_id);
CREATE INDEX idx_drifts_status ON drifts(status);
CREATE INDEX idx_drifts_severity ON drifts(severity);

-- Generated tests from drift
CREATE TABLE drift_generated_tests (
    id UUID PRIMARY KEY,
    drift_id UUID REFERENCES drifts(id),
    test_name VARCHAR(255) NOT NULL,
    test_code TEXT NOT NULL,
    confidence_score INTEGER,
    status VARCHAR(50) DEFAULT 'generated', -- generated, reviewed, rejected
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_drift_generated_tests_drift ON drift_generated_tests(drift_id);
```

---

## Sprint 13-14: GitHub Webhook Integration (Weeks 25-28)

### Task 13.1: Webhook Endpoint Setup

**Objective:** Create REST endpoint to receive GitHub webhook events

**Implementation:**

```go
// internal/api/webhook.go
package api

import (
  "crypto/hmac"
  "crypto/sha256"
  "encoding/hex"
  "encoding/json"
  "io"
  "log"
  "net/http"
  "os"

  "github.com/go-go-golems/gotest-agent/internal/drift"
)

type GitHubWebhookPayload struct {
  Ref        string `json:"ref"`
  Repository struct {
    FullName string `json:"full_name"`
    CloneURL string `json:"clone_url"`
    HTMLURL  string `json:"html_url"`
  } `json:"repository"`
  Commits []struct {
    ID        string   `json:"id"`
    Message   string   `json:"message"`
    Added     []string `json:"added"`
    Modified  []string `json:"modified"`
    Removed   []string `json:"removed"`
    Author    struct {
      Name  string `json:"name"`
      Email string `json:"email"`
    } `json:"author"`
    Timestamp string `json:"timestamp"`
  } `json:"commits"`
  Sender struct {
    Login string `json:"login"`
  } `json:"sender"`
}

// HandleGitHubWebhook handles GitHub webhook events
func (s *Server) HandleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
  // Verify webhook signature
  signature := r.Header.Get("X-Hub-Signature-256")
  if signature == "" {
    log.Printf("Missing webhook signature")
    http.Error(w, "Missing signature", http.StatusUnauthorized)
    return
  }

  // Read body
  body, err := io.ReadAll(r.Body)
  if err != nil {
    log.Printf("Failed to read webhook body: %v", err)
    http.Error(w, "Failed to read body", http.StatusInternalServerError)
    return
  }

  // Verify HMAC signature
  secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
  if secret == "" {
    log.Printf("GITHUB_WEBHOOK_SECRET not configured")
    http.Error(w, "Webhook secret not configured", http.StatusInternalServerError)
    return
  }

  mac := hmac.New(sha256.New, []byte(secret))
  mac.Write(body)
  expectedMAC := hex.EncodeToString(mac.Sum(nil))
  
  // Signature format: "sha256=<hash>"
  if len(signature) < 7 || signature[:7] != "sha256=" {
    log.Printf("Invalid signature format: %s", signature)
    http.Error(w, "Invalid signature format", http.StatusUnauthorized)
    return
  }
  
  if !hmac.Equal([]byte(signature[7:]), []byte(expectedMAC)) {
    log.Printf("Invalid webhook signature")
    http.Error(w, "Invalid signature", http.StatusUnauthorized)
    return
  }

  // Parse payload
  var payload GitHubWebhookPayload
  if err := json.Unmarshal(body, &payload); err != nil {
    log.Printf("Failed to parse webhook payload: %v", err)
    http.Error(w, "Invalid payload", http.StatusBadRequest)
    return
  }

  log.Printf("Received webhook for %s from %s", payload.Repository.FullName, payload.Sender.Login)

  // Extract changed files
  var changedFiles []string
  for _, commit := range payload.Commits {
    changedFiles = append(changedFiles, commit.Added...)
    changedFiles = append(changedFiles, commit.Modified...)
    changedFiles = append(changedFiles, commit.Removed...)
  }

  log.Printf("Detected %d changed files in %s", len(changedFiles), payload.Repository.FullName)

  // Trigger drift detection asynchronously
  go func() {
    if s.driftDetector != nil {
      s.driftDetector.DetectDrift(payload.Repository.CloneURL, changedFiles)
    }
  }()

  // Return success response
  w.WriteHeader(http.StatusOK)
  json.NewEncoder(w).Encode(map[string]string{
    "status":  "accepted",
    "message": "Drift detection triggered",
    "files":   string(len(changedFiles)),
  })
}
```

**Acceptance Criteria:**
- Webhook endpoint accepts POST requests at `/api/v1/webhooks/github`
- HMAC signature verification works correctly
- Payload parsing extracts changed files correctly
- Drift detection triggered asynchronously
- Returns 200 OK with JSON response

**Testing:**

```go
// internal/api/webhook_test.go
package api

import (
  "bytes"
  "crypto/hmac"
  "crypto/sha256"
  "encoding/hex"
  "encoding/json"
  "net/http"
  "net/http/httptest"
  "os"
  "testing"
)

func TestHandleGitHubWebhook_ValidSignature(t *testing.T) {
  // Setup
  os.Setenv("GITHUB_WEBHOOK_SECRET", "test-secret")
  defer os.Unsetenv("GITHUB_WEBHOOK_SECRET")

  server := NewServer(nil, nil, nil)
  
  // Create payload
  payload := map[string]interface{}{
    "ref": "refs/heads/main",
    "repository": map[string]interface{}{
      "full_name": "test/repo",
      "clone_url": "https://github.com/test/repo.git",
    },
    "commits": []map[string]interface{}{
      {
        "id":      "abc123",
        "message": "Test commit",
        "added":   []string{"file1.go", "file2.go"},
        "modified": []string{"file3.go"},
        "removed":  []string{},
      },
    },
    "sender": map[string]interface{}{
      "login": "testuser",
    },
  }

  payloadBytes, _ := json.Marshal(payload)

  // Generate signature
  mac := hmac.New(sha256.New, []byte("test-secret"))
  mac.Write(payloadBytes)
  signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

  // Create request
  req := httptest.NewRequest("POST", "/api/v1/webhooks/github", bytes.NewReader(payloadBytes))
  req.Header.Set("X-Hub-Signature-256", signature)
  req.Header.Set("Content-Type", "application/json")

  w := httptest.NewRecorder()

  // Execute
  server.HandleGitHubWebhook(w, req)

  // Verify
  if w.Code != http.StatusOK {
    t.Errorf("Expected status 200, got %d", w.Code)
  }

  var response map[string]string
  json.Unmarshal(w.Body.Bytes(), &response)

  if response["status"] != "accepted" {
    t.Errorf("Expected status 'accepted', got %s", response["status"])
  }
}

func TestHandleGitHubWebhook_InvalidSignature(t *testing.T) {
  os.Setenv("GITHUB_WEBHOOK_SECRET", "test-secret")
  defer os.Unsetenv("GITHUB_WEBHOOK_SECRET")

  server := NewServer(nil, nil, nil)

  payload := map[string]interface{}{
    "ref": "refs/heads/main",
  }
  payloadBytes, _ := json.Marshal(payload)

  req := httptest.NewRequest("POST", "/api/v1/webhooks/github", bytes.NewReader(payloadBytes))
  req.Header.Set("X-Hub-Signature-256", "sha256=invalid-signature")
  req.Header.Set("Content-Type", "application/json")

  w := httptest.NewRecorder()
  server.HandleGitHubWebhook(w, req)

  if w.Code != http.StatusUnauthorized {
    t.Errorf("Expected status 401, got %d", w.Code)
  }
}

func TestHandleGitHubWebhook_MissingSignature(t *testing.T) {
  os.Setenv("GITHUB_WEBHOOK_SECRET", "test-secret")
  defer os.Unsetenv("GITHUB_WEBHOOK_SECRET")

  server := NewServer(nil, nil, nil)

  payload := map[string]interface{}{"ref": "refs/heads/main"}
  payloadBytes, _ := json.Marshal(payload)

  req := httptest.NewRequest("POST", "/api/v1/webhooks/github", bytes.NewReader(payloadBytes))
  req.Header.Set("Content-Type", "application/json")
  // No signature header

  w := httptest.NewRecorder()
  server.HandleGitHubWebhook(w, req)

  if w.Code != http.StatusUnauthorized {
    t.Errorf("Expected status 401, got %d", w.Code)
  }
}
```

---

### Task 13.2: Webhook Registration Service

**Objective:** Allow users to register their GitHub repositories for monitoring

**Implementation:**

```go
// internal/webhook/service.go
package webhook

import (
  "bytes"
  "context"
  "database/sql"
  "encoding/json"
  "fmt"
  "io"
  "log"
  "net/http"
  "os"
  "strings"
  "time"

  "github.com/google/uuid"
)

type WebhookRegistration struct {
  ID            string    `json:"id" db:"id"`
  UserID        string    `json:"user_id" db:"user_id"`
  RepositoryID  string    `json:"repository_id" db:"repository_id"`
  RepositoryURL string    `json:"repository_url" db:"repository_url"`
  WebhookID     string    `json:"webhook_id" db:"webhook_id"`
  Status        string    `json:"status" db:"status"`
  LastSyncAt    time.Time `json:"last_sync_at" db:"last_sync_at"`
  CreatedAt     time.Time `json:"created_at" db:"created_at"`
  UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type Service struct {
  db *sql.DB
}

func NewService(db *sql.DB) *Service {
  return &Service{db: db}
}

// Register registers a GitHub repository for webhook monitoring
func (s *Service) Register(ctx context.Context, userID, repoURL, githubToken string) (*WebhookRegistration, error) {
  // Parse repository owner and name from URL
  // Example: https://github.com/owner/repo
  parts := strings.Split(repoURL, "/")
  if len(parts) < 2 {
    return nil, fmt.Errorf("invalid repository URL: %s", repoURL)
  }
  
  owner := parts[len(parts)-2]
  repo := strings.TrimSuffix(parts[len(parts)-1], ".git")

  // Create GitHub webhook via GitHub API
  webhookID, err := s.createGitHubWebhook(owner, repo, githubToken)
  if err != nil {
    log.Printf("Failed to create GitHub webhook: %v", err)
    return nil, fmt.Errorf("failed to create GitHub webhook: %w", err)
  }

  // Store registration
  reg := &WebhookRegistration{
    ID:            uuid.New().String(),
    UserID:        userID,
    RepositoryURL: repoURL,
    WebhookID:     webhookID,
    Status:        "active",
    CreatedAt:     time.Now(),
    UpdatedAt:     time.Now(),
  }

  _, err = s.db.ExecContext(ctx,
    `INSERT INTO webhook_registrations 
     (id, user_id, repository_url, webhook_id, status, created_at, updated_at)
     VALUES ($1, $2, $3, $4, $5, $6, $7)`,
    reg.ID, reg.UserID, reg.RepositoryURL, reg.WebhookID, reg.Status,
    reg.CreatedAt, reg.UpdatedAt)

  if err != nil {
    log.Printf("Failed to store webhook registration: %v", err)
    return nil, fmt.Errorf("failed to store registration: %w", err)
  }

  log.Printf("Registered webhook for %s (ID: %s)", repoURL, webhookID)
  return reg, nil
}

// createGitHubWebhook creates a webhook in GitHub via API
func (s *Service) createGitHubWebhook(owner, repo, githubToken string) (string, error) {
  webhookURL := os.Getenv("WEBHOOK_URL")
  if webhookURL == "" {
    return "", fmt.Errorf("WEBHOOK_URL not configured")
  }

  webhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET")
  if webhookSecret == "" {
    return "", fmt.Errorf("GITHUB_WEBHOOK_SECRET not configured")
  }

  // GitHub API: POST /repos/{owner}/{repo}/hooks
  apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/hooks", owner, repo)

  payload := map[string]interface{}{
    "name":   "web",
    "active": true,
    "events": []string{"push"},
    "config": map[string]interface{}{
      "url":          webhookURL,
      "content_type": "json",
      "secret":       webhookSecret,
    },
  }

  payloadBytes, err := json.Marshal(payload)
  if err != nil {
    return "", fmt.Errorf("failed to marshal payload: %w", err)
  }

  req, err := http.NewRequest("POST", apiURL, bytes.NewReader(payloadBytes))
  if err != nil {
    return "", fmt.Errorf("failed to create request: %w", err)
  }

  req.Header.Set("Authorization", "token "+githubToken)
  req.Header.Set("Accept", "application/vnd.github.v3+json")
  req.Header.Set("Content-Type", "application/json")

  client := &http.Client{Timeout: 30 * time.Second}
  resp, err := client.Do(req)
  if err != nil {
    return "", fmt.Errorf("failed to call GitHub API: %w", err)
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusCreated {
    body, _ := io.ReadAll(resp.Body)
    return "", fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
  }

  var response struct {
    ID int `json:"id"`
  }
  if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
    return "", fmt.Errorf("failed to decode response: %w", err)
  }

  return fmt.Sprintf("%d", response.ID), nil
}

// List lists all webhook registrations for a user
func (s *Service) List(ctx context.Context, userID string) ([]*WebhookRegistration, error) {
  rows, err := s.db.QueryContext(ctx,
    `SELECT id, user_id, repository_url, webhook_id, status, last_sync_at, created_at, updated_at
     FROM webhook_registrations WHERE user_id = $1 ORDER BY created_at DESC`,
    userID)

  if err != nil {
    return nil, fmt.Errorf("failed to query registrations: %w", err)
  }
  defer rows.Close()

  var registrations []*WebhookRegistration
  for rows.Next() {
    var reg WebhookRegistration
    err := rows.Scan(
      &reg.ID, &reg.UserID, &reg.RepositoryURL, &reg.WebhookID,
      &reg.Status, &reg.LastSyncAt, &reg.CreatedAt, &reg.UpdatedAt)

    if err != nil {
      return nil, fmt.Errorf("failed to scan row: %w", err)
    }

    registrations = append(registrations, &reg)
  }

  return registrations, nil
}

// Unregister removes a webhook registration
func (s *Service) Unregister(ctx context.Context, registrationID string) error {
  // Get registration
  var reg WebhookRegistration
  err := s.db.QueryRowContext(ctx,
    `SELECT webhook_id, repository_url FROM webhook_registrations WHERE id = $1`,
    registrationID).Scan(&reg.WebhookID, &reg.RepositoryURL)

  if err != nil {
    return fmt.Errorf("failed to get registration: %w", err)
  }

  // Parse repository owner and name
  parts := strings.Split(reg.RepositoryURL, "/")
  owner := parts[len(parts)-2]
  repo := strings.TrimSuffix(parts[len(parts)-1], ".git")

  // Delete webhook from GitHub API
  githubToken := os.Getenv("GITHUB_TOKEN")
  if githubToken == "" {
    log.Printf("Warning: GITHUB_TOKEN not configured, skipping GitHub API cleanup")
  } else {
    apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/hooks/%s", owner, repo, reg.WebhookID)
    
    req, err := http.NewRequest("DELETE", apiURL, nil)
    if err != nil {
      log.Printf("Failed to create delete request: %v", err)
    } else {
      req.Header.Set("Authorization", "token "+githubToken)
      req.Header.Set("Accept", "application/vnd.github.v3+json")

      client := &http.Client{Timeout: 30 * time.Second}
      resp, err := client.Do(req)
      if err != nil {
        log.Printf("Failed to delete webhook from GitHub: %v", err)
      } else {
        resp.Body.Close()
        log.Printf("Deleted webhook %s from GitHub", reg.WebhookID)
      }
    }
  }

  // Delete from database
  _, err = s.db.ExecContext(ctx,
    `DELETE FROM webhook_registrations WHERE id = $1`,
    registrationID)

  if err != nil {
    return fmt.Errorf("failed to delete registration: %w", err)
  }

  log.Printf("Unregistered webhook %s", registrationID)
  return nil
}
```

**Acceptance Criteria:**
- Users can register repositories via API
- GitHub webhook created via GitHub API
- Registration stored in database
- Users can list their registrations
- Users can unregister repositories (deletes from GitHub and database)

---

*(Document continues with remaining tasks...)*

**Summary:**
Phase 3 implements continuous synchronization between code and tests. The system will:
- Monitor GitHub repositories via webhooks
- Auto-detect drift when code changes
- Auto-generate/update tests as needed
- Alert teams on outdated tests
- Maintain 95%+ test coverage automatically

**Timeline:** 10 weeks (5 sprints)  
**Key Deliverables:**
1. GitHub webhook integration
2. Webhook registration service
3. Drift detection service
4. Auto-generation service
5. Alert system

**Success Criteria:**
- Zero manual test updates needed
- 95% test coverage maintained
- Drift detected within 5 minutes of code change
- Auto-generation completes within 10 minutes
- Team alerted within 1 minute of drift detection

---

*End of Phase 3 Implementation Guide*
