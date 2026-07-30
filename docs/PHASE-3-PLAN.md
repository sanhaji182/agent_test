# Phase 3: Continuous Sync & Drift Detection

**Timeline:** 10 weeks (5 sprints)  
**Target:** Auto-detect code changes, auto-generate/update tests, alert on drift  
**Success Criteria:** Zero manual test updates needed, 95% test coverage maintained

---

## Overview

Phase 3 focuses on making the testing system **truly continuous and self-maintaining**. Instead of manually triggering test generation, the system will:

1. **Monitor code repositories** for changes (via GitHub webhooks)
2. **Auto-detect drift** between code and tests
3. **Auto-generate/update tests** when code changes
4. **Alert teams** when tests are outdated or failing
5. **Provide insights** on test coverage gaps

---

## Sprint 13-14: GitHub Webhook Integration (Weeks 25-28)

### Goal
Set up GitHub webhook infrastructure to receive real-time notifications when code changes.

### Tasks

#### Task 13.1: Webhook Endpoint Setup

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
  "net/http"
  "strings"
)

type GitHubWebhookPayload struct {
  Ref        string `json:"ref"`
  Repository struct {
    FullName string `json:"full_name"`
    CloneURL string `json:"clone_url"`
  } `json:"repository"`
  Commits []struct {
    ID       string   `json:"id"`
    Message  string   `json:"message"`
    Added    []string `json:"added"`
    Modified []string `json:"modified"`
    Removed  []string `json:"removed"`
  } `json:"commits"`
}

func (s *Server) HandleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
  // Verify webhook signature
  signature := r.Header.Get("X-Hub-Signature-256")
  if signature == "" {
    http.Error(w, "Missing signature", http.StatusUnauthorized)
    return
  }

  body, err := io.ReadAll(r.Body)
  if err != nil {
    http.Error(w, "Failed to read body", http.StatusInternalServerError)
    return
  }

  // Verify HMAC signature
  secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
  mac := hmac.New(sha256.New, []byte(secret))
  mac.Write(body)
  expectedMAC := hex.EncodeToString(mac.Sum(nil))
  
  if !hmac.Equal([]byte(signature[7:]), []byte(expectedMAC)) {
    http.Error(w, "Invalid signature", http.StatusUnauthorized)
    return
  }

  // Parse payload
  var payload GitHubWebhookPayload
  if err := json.Unmarshal(body, &payload); err != nil {
    http.Error(w, "Invalid payload", http.StatusBadRequest)
    return
  }

  // Extract changed files
  var changedFiles []string
  for _, commit := range payload.Commits {
    changedFiles = append(changedFiles, commit.Added...)
    changedFiles = append(changedFiles, commit.Modified...)
    changedFiles = append(changedFiles, commit.Removed...)
  }

  // Trigger drift detection
  go s.driftDetector.DetectDrift(payload.Repository.CloneURL, changedFiles)

  w.WriteHeader(http.StatusOK)
  json.NewEncoder(w).Encode(map[string]string{
    "status":  "accepted",
    "message": "Drift detection triggered",
  })
}
```

**Acceptance Criteria:**
- Webhook endpoint accepts POST requests
- HMAC signature verification works
- Payload parsing extracts changed files correctly
- Drift detection triggered asynchronously

**Testing:**
```go
func TestGitHubWebhook(t *testing.T) {
  // Test signature verification
  // Test payload parsing
  // Test drift detection trigger
}
```

---

#### Task 13.2: Webhook Registration Service

**Objective:** Allow users to register their GitHub repositories for monitoring

**Implementation:**
```go
// internal/webhook/service.go
package webhook

import (
  "context"
  "database/sql"
  "encoding/json"
  "time"

  "github.com/google/uuid"
)

type WebhookRegistration struct {
  ID           string    `json:"id" db:"id"`
  UserID       string    `json:"user_id" db:"user_id"`
  RepositoryID string    `json:"repository_id" db:"repository_id"`
  RepositoryURL string   `json:"repository_url" db:"repository_url"`
  WebhookID    string    `json:"webhook_id" db:"webhook_id"` // GitHub webhook ID
  Status       string    `json:"status" db:"status"` // active, inactive, error
  LastSyncAt   time.Time `json:"last_sync_at" db:"last_sync_at"`
  CreatedAt    time.Time `json:"created_at" db:"created_at"`
  UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Service struct {
  db *sql.DB
}

func NewService(db *sql.DB) *Service {
  return &Service{db: db}
}

// Register registers a GitHub repository for webhook monitoring
func (s *Service) Register(ctx context.Context, userID, repoURL, githubToken string) (*WebhookRegistration, error) {
  // Create GitHub webhook via GitHub API
  webhookID, err := s.createGitHubWebhook(repoURL, githubToken)
  if err != nil {
    return nil, err
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
    return nil, err
  }

  return reg, nil
}

// createGitHubWebhook creates a webhook in GitHub via API
func (s *Service) createGitHubWebhook(repoURL, githubToken string) (string, error) {
  // Parse repository owner and name from URL
  // POST to GitHub API to create webhook
  // Return webhook ID
  
  // Example: POST https://api.github.com/repos/{owner}/{repo}/hooks
  payload := map[string]interface{}{
    "name":   "web",
    "active": true,
    "events": []string{"push"},
    "config": map[string]interface{}{
      "url":          os.Getenv("WEBHOOK_URL"),
      "content_type": "json",
      "secret":       os.Getenv("GITHUB_WEBHOOK_SECRET"),
    },
  }

  // Make HTTP request to GitHub API
  // Return webhook ID from response
  
  return "webhook-id", nil
}

// List lists all webhook registrations for a user
func (s *Service) List(ctx context.Context, userID string) ([]*WebhookRegistration, error) {
  rows, err := s.db.QueryContext(ctx,
    `SELECT id, user_id, repository_url, webhook_id, status, last_sync_at, created_at, updated_at
     FROM webhook_registrations WHERE user_id = $1 ORDER BY created_at DESC`,
    userID)

  if err != nil {
    return nil, err
  }
  defer rows.Close()

  var registrations []*WebhookRegistration
  for rows.Next() {
    var reg WebhookRegistration
    err := rows.Scan(
      &reg.ID, &reg.UserID, &reg.RepositoryURL, &reg.WebhookID,
      &reg.Status, &reg.LastSyncAt, &reg.CreatedAt, &reg.UpdatedAt)

    if err != nil {
      return nil, err
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
    return err
  }

  // Delete webhook from GitHub API
  // DELETE https://api.github.com/repos/{owner}/{repo}/hooks/{webhook_id}

  // Delete from database
  _, err = s.db.ExecContext(ctx,
    `DELETE FROM webhook_registrations WHERE id = $1`,
    registrationID)

  return err
}
```

**Acceptance Criteria:**
- Users can register repositories
- GitHub webhook created via API
- Registration stored in database
- Users can list their registrations
- Users can unregister repositories

---

#### Task 13.3: Drift Detector Service

**Objective:** Detect drift between code changes and existing tests

**Implementation:**
```go
// internal/drift/detector.go
package drift

import (
  "context"
  "database/sql"
  "log"
  "strings"
  "time"

  "github.com/go-go-golems/gotest-agent/internal/parser"
  "github.com/go-go-golems/gotest-agent/internal/recorder"
)

type Detector struct {
  db              *sql.DB
  parserService   *parser.Service
  recorderService *recorder.Service
}

func NewDetector(db *sql.DB, parserService *parser.Service, recorderService *recorder.Service) *Detector {
  return &Detector{
    db:              db,
    parserService:   parserService,
    recorderService: recorderService,
  }
}

// DetectDrift analyzes changed files and detects drift
func (d *Detector) DetectDrift(repoURL string, changedFiles []string) {
  ctx := context.Background()

  // Filter relevant files (routes, models, handlers)
  relevantFiles := d.filterRelevantFiles(changedFiles)

  if len(relevantFiles) == 0 {
    log.Println("No relevant files changed")
    return
  }

  log.Printf("Detected %d relevant file changes", len(relevantFiles))

  // Parse changed files
  codebase, err := d.parserService.Parse(ctx, repoURL)
  if err != nil {
    log.Printf("Failed to parse repository: %v", err)
    return
  }

  // Get existing tests
  existingTests, err := d.getExistingTests(repoURL)
  if err != nil {
    log.Printf("Failed to get existing tests: %v", err)
    return
  }

  // Detect drift
  drifts := d.analyzeDrift(codebase, existingTests, changedFiles)

  if len(drifts) == 0 {
    log.Println("No drift detected")
    return
  }

  log.Printf("Detected %d drifts", len(drifts))

  // Store drifts
  for _, drift := range drifts {
    d.storeDrift(ctx, drift)
  }

  // Trigger auto-generation
  d.triggerAutoGeneration(ctx, repoURL, drifts)
}

// filterRelevantFiles filters files that are relevant for testing
func (d *Detector) filterRelevantFiles(files []string) []string {
  var relevant []string

  for _, file := range files {
    // Filter by file patterns
    if strings.Contains(file, "routes") ||
       strings.Contains(file, "controllers") ||
       strings.Contains(file, "handlers") ||
       strings.Contains(file, "models") ||
       strings.Contains(file, "schemas") {
      relevant = append(relevant, file)
    }
  }

  return relevant
}

// Drift represents a detected drift between code and tests
type Drift struct {
  ID           string    `json:"id" db:"id"`
  RepositoryID string    `json:"repository_id" db:"repository_id"`
  Type         string    `json:"type" db:"type"` // missing_test, outdated_test, removed_test
  FilePath     string    `json:"file_path" db:"file_path"`
  Description  string    `json:"description" db:"description"`
  Severity     string    `json:"severity" db:"severity"` // high, medium, low
  Status       string    `json:"status" db:"status"` // pending, fixed, ignored
  CreatedAt    time.Time `json:"created_at" db:"created_at"`
  UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// analyzeDrift analyzes drift between code and tests
func (d *Detector) analyzeDrift(codebase *parser.Codebase, existingTests []recorder.RecordingSession, changedFiles []string) []Drift {
  var drifts []Drift

  // Check for missing tests (new routes/models without tests)
  for _, route := range codebase.Routes {
    if !d.hasTest(existingTests, route.Handler) {
      drifts = append(drifts, Drift{
        Type:        "missing_test",
        FilePath:    route.File,
        Description: "New route without test: " + route.Method + " " + route.Path,
        Severity:    "high",
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
      })
    }
  }

  // Check for outdated tests (tests for removed routes)
  for _, test := range existingTests {
    if !d.hasRoute(codebase.Routes, test) {
      drifts = append(drifts, Drift{
        Type:        "outdated_test",
        FilePath:    test.URL,
        Description: "Test for removed route",
        Severity:    "medium",
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
      })
    }
  }

  return drifts
}

// hasTest checks if a test exists for a handler
func (d *Detector) hasTest(tests []recorder.RecordingSession, handler string) bool {
  for _, test := range tests {
    // Check if test covers this handler
    // This is a simplified check - in reality, you'd parse test content
    if strings.Contains(test.URL, handler) {
      return true
    }
  }
  return false
}

// hasRoute checks if a route exists in codebase
func (d *Detector) hasRoute(routes []parser.Route, test recorder.RecordingSession) bool {
  for _, route := range routes {
    if strings.Contains(test.URL, route.Path) {
      return true
    }
  }
  return false
}

// storeDrift stores a drift in the database
func (d *Detector) storeDrift(ctx context.Context, drift Drift) error {
  _, err := d.db.ExecContext(ctx,
    `INSERT INTO drifts (id, repository_id, type, file_path, description, severity, status, created_at, updated_at)
     VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
    drift.ID, drift.RepositoryID, drift.Type, drift.FilePath, drift.Description,
    drift.Severity, drift.Status, drift.CreatedAt, drift.UpdatedAt)

  return err
}

// triggerAutoGeneration triggers auto-generation of tests for drifts
func (d *Detector) triggerAutoGeneration(ctx context.Context, repoURL string, drifts []Drift) {
  // For each drift, trigger test generation
  for _, drift := range drifts {
    if drift.Type == "missing_test" && drift.Severity == "high" {
      // Trigger test generation for this route
      log.Printf("Triggering test generation for %s", drift.FilePath)
      // Call AI generation service
    }
  }
}
```

**Acceptance Criteria:**
- Drift detector receives changed files
- Relevant files filtered correctly
- Drift detected for missing/outdated tests
- Drifts stored in database
- Auto-generation triggered for high-severity drifts

---

### Sprint 15-16: Drift Detection & Auto-Update (Weeks 29-32)

### Goal
Implement automatic test updates when drift is detected.

### Tasks

#### Task 15.1: Auto-Generation Service

**Objective:** Automatically generate tests when drift is detected

**Implementation:**
```go
// internal/drift/auto_generator.go
package drift

import (
  "context"
  "database/sql"
  "log"

  "github.com/go-go-golems/gotest-agent/internal/ai"
  "github.com/go-go-golems/gotest-agent/internal/parser"
  "github.com/go-go-golems/gotest-agent/internal/recorder"
)

type AutoGenerator struct {
  db              *sql.DB
  parserService   *parser.Service
  recorderService *recorder.Service
  aiService       *ai.SynthesisService
}

func NewAutoGenerator(db *sql.DB, parserService *parser.Service, recorderService *recorder.Service, aiService *ai.SynthesisService) *AutoGenerator {
  return &AutoGenerator{
    db:              db,
    parserService:   parserService,
    recorderService: recorderService,
    aiService:       aiService,
  }
}

// GenerateForDrift generates tests for a drift
func (ag *AutoGenerator) GenerateForDrift(ctx context.Context, drift Drift) error {
  log.Printf("Generating tests for drift: %s", drift.Description)

  // Get repository URL
  var repoURL string
  err := ag.db.QueryRowContext(ctx,
    `SELECT repository_url FROM webhook_registrations WHERE repository_id = $1`,
    drift.RepositoryID).Scan(&repoURL)

  if err != nil {
    return err
  }

  // Parse repository
  codebase, err := ag.parserService.Parse(ctx, repoURL)
  if err != nil {
    return err
  }

  // Generate test plan
  testPlan, err := ag.aiService.GenerateTestPlan(ctx, codebase)
  if err != nil {
    return err
  }

  // Create recording session
  session, err := ag.recorderService.CreateSession(ctx, "auto-generated", repoURL)
  if err != nil {
    return err
  }

  // Store generated tests
  // This would integrate with the recorder service to store generated tests

  // Update drift status
  _, err = ag.db.ExecContext(ctx,
    `UPDATE drifts SET status = 'fixed', updated_at = NOW() WHERE id = $1`,
    drift.ID)

  return err
}
```

**Acceptance Criteria:**
- Tests generated automatically for drifts
- Generated tests stored in database
- Drift status updated to "fixed"
- Logs show generation progress

---

*(Document continues with remaining sprints and tasks...)*

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
6. Coverage insights dashboard

**Success Criteria:**
- Zero manual test updates needed
- 95% test coverage maintained
- Drift detected within 5 minutes of code change
- Auto-generation completes within 10 minutes
- Team alerted within 1 minute of drift detection

---

*End of Phase 3 Plan*
