package github

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CloneOptions configures repository cloning behavior
type CloneOptions struct {
	// Target directory where repo will be cloned
	TargetDir string

	// Use shallow clone (depth=1) for performance
	Shallow bool

	// GitHub personal access token for private repos
	Token string

	// Specific branch to clone (empty = default branch)
	Branch string
}

// CloneClient handles repository cloning operations
type CloneClient struct {
	oauth *OAuthClient
}

// NewCloneClient creates a new clone client
func NewCloneClient(oauth *OAuthClient) *CloneClient {
	return &CloneClient{oauth: oauth}
}

// CloneRepository clones a GitHub repository to the specified directory
func (c *CloneClient) CloneRepository(ctx context.Context, repoFullName string, opts CloneOptions) error {
	if repoFullName == "" {
		return fmt.Errorf("repository name is required")
	}

	// Validate repository name format (owner/repo)
	parts := strings.Split(repoFullName, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid repository name format: %s (expected owner/repo)", repoFullName)
	}

	// Create target directory if it doesn't exist
	if opts.TargetDir == "" {
		return fmt.Errorf("target directory is required")
	}

	if err := os.MkdirAll(opts.TargetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Build git clone URL
	var cloneURL string
	if opts.Token != "" {
		// Use HTTPS with token for private repos
		cloneURL = fmt.Sprintf("https://%s@github.com/%s.git", opts.Token, repoFullName)
	} else {
		// Use public HTTPS URL
		cloneURL = fmt.Sprintf("https://github.com/%s.git", repoFullName)
	}

	// Build git clone command
	args := []string{"clone"}

	if opts.Shallow {
		args = append(args, "--depth", "1")
	}

	if opts.Branch != "" {
		args = append(args, "--branch", opts.Branch)
		args = append(args, "--single-branch")
	}

	args = append(args, cloneURL, opts.TargetDir)

	// Execute git clone
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = os.Environ()

	// Disable git credential prompts
	cmd.Env = append(cmd.Env, "GIT_TERMINAL_PROMPT=0")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// CloneToTemp clones a repository to a temporary directory
func (c *CloneClient) CloneToTemp(ctx context.Context, repoFullName string, opts CloneOptions) (string, error) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "gotest-clone-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	opts.TargetDir = tempDir

	if err := c.CloneRepository(ctx, repoFullName, opts); err != nil {
		// Clean up on failure
		os.RemoveAll(tempDir)
		return "", err
	}

	return tempDir, nil
}

// GetRepoInfo returns information about a cloned repository
type RepoInfo struct {
	CurrentBranch string
	RemoteURL     string
	CommitHash    string
	IsShallow     bool
}

// GetRepoInfo retrieves information about a cloned repository
func (c *CloneClient) GetRepoInfo(repoDir string) (*RepoInfo, error) {
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory is not a git repository: %s", repoDir)
	}

	info := &RepoInfo{}

	// Get current branch
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoDir
	output, err := cmd.Output()
	if err == nil {
		info.CurrentBranch = strings.TrimSpace(string(output))
	}

	// Get remote URL
	cmd = exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = repoDir
	output, err = cmd.Output()
	if err == nil {
		info.RemoteURL = strings.TrimSpace(string(output))
	}

	// Get commit hash
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoDir
	output, err = cmd.Output()
	if err == nil {
		info.CommitHash = strings.TrimSpace(string(output))
	}

	// Check if shallow clone
	if _, err := os.Stat(filepath.Join(repoDir, ".git", "shallow")); err == nil {
		info.IsShallow = true
	}

	return info, nil
}

// PullUpdates pulls latest changes from remote
func (c *CloneClient) PullUpdates(ctx context.Context, repoDir string) error {
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); os.IsNotExist(err) {
		return fmt.Errorf("directory is not a git repository: %s", repoDir)
	}

	cmd := exec.CommandContext(ctx, "git", "pull", "--ff-only")
	cmd.Dir = repoDir
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GIT_TERMINAL_PROMPT=0")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// CleanRepo removes all untracked files and resets to clean state
func (c *CloneClient) CleanRepo(ctx context.Context, repoDir string) error {
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); os.IsNotExist(err) {
		return fmt.Errorf("directory is not a git repository: %s", repoDir)
	}

	// Reset to HEAD
	cmd := exec.CommandContext(ctx, "git", "reset", "--hard", "HEAD")
	cmd.Dir = repoDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git reset failed: %w\nOutput: %s", err, string(output))
	}

	// Clean untracked files
	cmd = exec.CommandContext(ctx, "git", "clean", "-fdx")
	cmd.Dir = repoDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clean failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// DeleteRepo removes a cloned repository directory
func (c *CloneClient) DeleteRepo(repoDir string) error {
	if repoDir == "" {
		return fmt.Errorf("repository directory is required")
	}

	// Safety check: ensure we're not deleting root or home directory
	absPath, err := filepath.Abs(repoDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if absPath == "/" || absPath == os.Getenv("HOME") {
		return fmt.Errorf("refusing to delete critical directory: %s", absPath)
	}

	if err := os.RemoveAll(repoDir); err != nil {
		return fmt.Errorf("failed to delete repository: %w", err)
	}

	return nil
}
