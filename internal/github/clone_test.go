package github

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCloneClient_CloneRepository_Validation(t *testing.T) {
	client := NewCloneClient(nil)
	ctx := context.Background()

	tests := []struct {
		name        string
		repoName    string
		opts        CloneOptions
		expectError string
	}{
		{
			name:        "empty repo name",
			repoName:    "",
			opts:        CloneOptions{TargetDir: "/tmp/test"},
			expectError: "repository name is required",
		},
		{
			name:        "invalid repo format - no slash",
			repoName:    "invalid",
			opts:        CloneOptions{TargetDir: "/tmp/test"},
			expectError: "invalid repository name format",
		},
		{
			name:        "invalid repo format - empty owner",
			repoName:    "/repo",
			opts:        CloneOptions{TargetDir: "/tmp/test"},
			expectError: "invalid repository name format",
		},
		{
			name:        "invalid repo format - empty repo",
			repoName:    "owner/",
			opts:        CloneOptions{TargetDir: "/tmp/test"},
			expectError: "invalid repository name format",
		},
		{
			name:        "empty target dir",
			repoName:    "owner/repo",
			opts:        CloneOptions{TargetDir: ""},
			expectError: "target directory is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.CloneRepository(ctx, tt.repoName, tt.opts)
			if err == nil {
				t.Errorf("expected error containing %q, got nil", tt.expectError)
				return
			}
			if !contains(err.Error(), tt.expectError) {
				t.Errorf("expected error containing %q, got %q", tt.expectError, err.Error())
			}
		})
	}
}

func TestCloneClient_CloneToTemp_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewCloneClient(nil)
	ctx := context.Background()

	// Clone a small public repo
	tempDir, err := client.CloneToTemp(ctx, "octocat/Hello-World", CloneOptions{
		Shallow: true,
	})

	if err != nil {
		t.Fatalf("CloneToTemp failed: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Verify directory exists and contains .git
	if _, err := os.Stat(filepath.Join(tempDir, ".git")); os.IsNotExist(err) {
		t.Error("cloned repository missing .git directory")
	}

	// Verify README exists (Hello-World repo has README)
	if _, err := os.Stat(filepath.Join(tempDir, "README")); os.IsNotExist(err) {
		t.Error("cloned repository missing README file")
	}
}

func TestCloneClient_GetRepoInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewCloneClient(nil)
	ctx := context.Background()

	// Clone a test repo
	tempDir, err := client.CloneToTemp(ctx, "octocat/Hello-World", CloneOptions{
		Shallow: true,
	})

	if err != nil {
		t.Fatalf("CloneToTemp failed: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Get repo info
	info, err := client.GetRepoInfo(tempDir)
	if err != nil {
		t.Fatalf("GetRepoInfo failed: %v", err)
	}

	// Verify info fields
	if info.CurrentBranch == "" {
		t.Error("CurrentBranch should not be empty")
	}

	if info.RemoteURL == "" {
		t.Error("RemoteURL should not be empty")
	}

	if info.CommitHash == "" {
		t.Error("CommitHash should not be empty")
	}

	if !info.IsShallow {
		t.Error("IsShallow should be true for shallow clone")
	}
}

func TestCloneClient_GetRepoInfo_NotGitRepo(t *testing.T) {
	client := NewCloneClient(nil)

	// Create temp directory without .git
	tempDir, err := os.MkdirTemp("", "not-git-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	_, err = client.GetRepoInfo(tempDir)
	if err == nil {
		t.Error("expected error for non-git directory, got nil")
	}

	if !contains(err.Error(), "not a git repository") {
		t.Errorf("expected error containing 'not a git repository', got %q", err.Error())
	}
}

func TestCloneClient_CleanRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewCloneClient(nil)
	ctx := context.Background()

	// Clone a test repo
	tempDir, err := client.CloneToTemp(ctx, "octocat/Hello-World", CloneOptions{
		Shallow: true,
	})

	if err != nil {
		t.Fatalf("CloneToTemp failed: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create an untracked file
	testFile := filepath.Join(tempDir, "untracked.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("test file was not created")
	}

	// Clean repo
	if err := client.CleanRepo(ctx, tempDir); err != nil {
		t.Fatalf("CleanRepo failed: %v", err)
	}

	// Verify untracked file was removed
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("untracked file was not removed by CleanRepo")
	}
}

func TestCloneClient_DeleteRepo(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "delete-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	client := NewCloneClient(nil)

	// Delete the repo
	if err := client.DeleteRepo(tempDir); err != nil {
		t.Fatalf("DeleteRepo failed: %v", err)
	}

	// Verify directory was deleted
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Error("directory was not deleted")
	}
}

func TestCloneClient_DeleteRepo_SafetyChecks(t *testing.T) {
	client := NewCloneClient(nil)

	tests := []struct {
		name        string
		dir         string
		expectError string
	}{
		{
			name:        "empty directory",
			dir:         "",
			expectError: "repository directory is required",
		},
		{
			name:        "root directory",
			dir:         "/",
			expectError: "refusing to delete critical directory",
		},
		{
			name:        "home directory",
			dir:         os.Getenv("HOME"),
			expectError: "refusing to delete critical directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.DeleteRepo(tt.dir)
			if err == nil {
				t.Errorf("expected error containing %q, got nil", tt.expectError)
				return
			}
			if !contains(err.Error(), tt.expectError) {
				t.Errorf("expected error containing %q, got %q", tt.expectError, err.Error())
			}
		})
	}
}

func TestCloneClient_CloneWithBranch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewCloneClient(nil)
	ctx := context.Background()

	// Clone specific branch
	tempDir, err := client.CloneToTemp(ctx, "octocat/Hello-World", CloneOptions{
		Shallow: true,
		Branch:  "master",
	})

	if err != nil {
		t.Fatalf("CloneToTemp with branch failed: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Verify branch
	info, err := client.GetRepoInfo(tempDir)
	if err != nil {
		t.Fatalf("GetRepoInfo failed: %v", err)
	}

	if info.CurrentBranch != "master" {
		t.Errorf("expected branch 'master', got %q", info.CurrentBranch)
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
