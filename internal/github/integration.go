package github

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
)

// Integration orchestrates GitHub webhook events with repository operations
type Integration struct {
	webhookHandler *WebhookHandler
	cloneClient    *CloneClient
	cloneDir       string // Base directory for cloning repositories
}

// NewIntegration creates a new GitHub integration
func NewIntegration(webhookSecret string, cloneDir string) *Integration {
	return &Integration{
		webhookHandler: NewWebhookHandler(webhookSecret),
		cloneClient:    NewCloneClient(nil),
		cloneDir:       cloneDir,
	}
}

// repoPath resolves the clone directory for a repository full name from a
// webhook payload, rejecting names that would escape the base clone dir.
func (i *Integration) repoPath(repoFullName string) (string, error) {
	base := filepath.Clean(i.cloneDir)
	dir := filepath.Clean(filepath.Join(base, repoFullName))
	if dir == base || !strings.HasPrefix(dir, base+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid repository name: %q", repoFullName)
	}
	return dir, nil
}

// ProcessWebhookEvent processes a webhook event and triggers appropriate actions
func (i *Integration) ProcessWebhookEvent(ctx context.Context, event *WebhookEvent) error {
	switch event.Type {
	case "push":
		pushEvent, err := i.webhookHandler.ParsePushEvent(event.Payload)
		if err != nil {
			return fmt.Errorf("failed to parse push event: %w", err)
		}
		return i.processPushEvent(ctx, pushEvent)
	case "pull_request":
		prEvent, err := i.webhookHandler.ParsePullRequestEvent(event.Payload)
		if err != nil {
			return fmt.Errorf("failed to parse pull request event: %w", err)
		}
		return i.processPullRequestEvent(ctx, prEvent)
	case "ping":
		log.Println("Received ping event from GitHub")
		return nil
	default:
		log.Printf("Ignoring unsupported event type: %s", event.Type)
		return nil
	}
}

// processPushEvent handles push webhook events
func (i *Integration) processPushEvent(ctx context.Context, pushEvent *PushEvent) error {
	log.Printf("Processing push event for %s (ref: %s)", pushEvent.Repository.FullName, pushEvent.Ref)

	// Skip if not a branch push (tags, etc.)
	if !strings.HasPrefix(pushEvent.Ref, "refs/heads/") {
		log.Printf("Skipping non-branch push: %s", pushEvent.Ref)
		return nil
	}

	// Extract branch name
	branch := strings.TrimPrefix(pushEvent.Ref, "refs/heads/")

	// Clone or pull repository
	repoDir, err := i.repoPath(pushEvent.Repository.FullName)
	if err != nil {
		return err
	}
	if err := i.ensureRepository(ctx, pushEvent.Repository.FullName, repoDir, branch); err != nil {
		return fmt.Errorf("failed to ensure repository: %w", err)
	}

	// Collect changed files from commits
	changedFiles := make([]string, 0)
	for _, commit := range pushEvent.Commits {
		changedFiles = append(changedFiles, commit.Added...)
		changedFiles = append(changedFiles, commit.Modified...)
		// Skip removed files as they don't need testing
	}

	if len(changedFiles) == 0 {
		log.Println("No changed files in push event")
		return nil
	}

	log.Printf("Detected %d changed files in %s", len(changedFiles), pushEvent.Repository.FullName)

	// TODO: Trigger test generation for changed files
	// For now, just log the information
	log.Printf("Would trigger test generation for files: %v", changedFiles)

	return nil
}

// processPullRequestEvent handles pull request webhook events
func (i *Integration) processPullRequestEvent(ctx context.Context, prEvent *PullRequestEvent) error {
	log.Printf("Processing pull request event for %s #%d (action: %s)",
		prEvent.Repository.FullName, prEvent.Number, prEvent.Action)

	// Only process opened, synchronize, and reopened actions
	switch prEvent.Action {
	case "opened", "synchronize", "reopened":
		// Clone repository and checkout PR branch
		repoDir, err := i.repoPath(prEvent.Repository.FullName)
		if err != nil {
			return err
		}
		if err := i.ensureRepository(ctx, prEvent.Repository.FullName, repoDir, prEvent.PullRequest.Head.Ref); err != nil {
			return fmt.Errorf("failed to ensure repository: %w", err)
		}

		// TODO: Trigger test generation for PR changes
		log.Printf("Would trigger test generation for PR #%d", prEvent.Number)

	case "closed":
		if prEvent.PullRequest.Merged {
			log.Printf("PR #%d was merged", prEvent.Number)
			// TODO: Could trigger full test suite after merge
		} else {
			log.Printf("PR #%d was closed without merging", prEvent.Number)
		}

	default:
		log.Printf("Ignoring PR action: %s", prEvent.Action)
	}

	return nil
}

// ensureRepository ensures the repository is cloned and up to date
func (i *Integration) ensureRepository(ctx context.Context, repoFullName, repoDir, branch string) error {
	// Check if repository exists
	if i.cloneClient.IsCloned(repoDir) {
		// Pull latest changes
		log.Printf("Pulling latest changes for %s (branch: %s)", repoFullName, branch)
		if err := i.cloneClient.Pull(ctx, repoDir); err != nil {
			log.Printf("Warning: failed to pull repository %s: %v", repoFullName, err)
			// Continue with existing clone
		}

		// Checkout the correct branch
		if err := i.cloneClient.CheckoutBranch(ctx, repoDir, branch); err != nil {
			return fmt.Errorf("failed to checkout branch %s: %w", branch, err)
		}
	} else {
		// Clone repository
		log.Printf("Cloning repository %s (branch: %s)", repoFullName, branch)
		if err := i.cloneClient.CloneRepository(ctx, repoFullName, CloneOptions{
			TargetDir: repoDir,
			Shallow:   true,
			Branch:    branch,
		}); err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
	}

	return nil
}

// GetWebhookHandler returns the webhook handler for HTTP integration
func (i *Integration) GetWebhookHandler() *WebhookHandler {
	return i.webhookHandler
}
