// Package gitdiff parses git diff output to determine impacted files,
// then maps them to relevant test suites for impacted test selection.
package gitdiff

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// ChangedFiles returns files changed in the working tree vs HEAD
func ChangedFiles(projectPath string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "HEAD")
	cmd.Dir = projectPath
	out, err := cmd.Output()
	if err != nil {
		// Fallback: try diff against last commit
		cmd = exec.Command("git", "diff", "--name-only", "HEAD~1", "HEAD")
		cmd.Dir = projectPath
		out, err = cmd.Output()
		if err != nil {
			return nil, err
		}
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

// MapToTests maps changed files to relevant test names using simple heuristics:
// - source file "src/auth/login.ts" → test "login"
// - controller "controllers/checkout.go" → test "checkout"
// - route file → test matching the route name
func MapToTests(changedFiles []string, allTests []string) []string {
	if len(changedFiles) == 0 {
		return allTests
	}

	// Extract keywords from changed file paths
	keywords := map[string]bool{}
	for _, f := range changedFiles {
		base := filepath.Base(f)
		name := strings.TrimSuffix(base, filepath.Ext(base))
		// Remove common suffixes
		for _, suffix := range []string{"_test", "_spec", ".test", ".spec", "_controller", "_handler", "_service"} {
			name = strings.TrimSuffix(name, suffix)
		}
		keywords[strings.ToLower(name)] = true

		// Also add parent directory name
		dir := filepath.Base(filepath.Dir(f))
		if dir != "." && dir != "" {
			keywords[strings.ToLower(dir)] = true
		}
	}

	// Match tests that contain any keyword
	var matched []string
	for _, test := range allTests {
		testLower := strings.ToLower(test)
		for kw := range keywords {
			if strings.Contains(testLower, kw) {
				matched = append(matched, test)
				break
			}
		}
	}

	if len(matched) == 0 {
		return allTests // Fallback: run all if no match
	}
	return matched
}
