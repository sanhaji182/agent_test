// Package runner berisi implementasi test executor (Docker dan Steel)
package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/execution"
	"github.com/go-go-golems/gotest-agent/internal/reporter"
)

// DockerRunner menjalankan Playwright test di dalam Docker container
type DockerRunner struct {
	image   string // Docker image Playwright
	timeout int    // Timeout dalam detik
	exec    *execution.Context
	runID   string // Set sebelum Run() dipanggil
}

// NewDockerRunner membuat runner baru dengan timeout tertentu
func NewDockerRunner(timeout int) *DockerRunner {
	return &DockerRunner{
		image:   "mcr.microsoft.com/playwright:v1.40.0-jammy",
		timeout: timeout,
	}
}

// SetExecContext mengatur execution context untuk emit events
func (r *DockerRunner) SetExecContext(exec *execution.Context, runID string) {
	r.exec = exec
	r.runID = runID
}

// Run menjalankan test files di Docker container dan mengembalikan hasil
func (r *DockerRunner) Run(ctx context.Context, testFiles []agent.TestFile, projectURL string) (*agent.RunResult, error) {
	// Buat direktori sementara untuk file test
	tmpDir, err := os.MkdirTemp("", "gotest-run-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Tulis semua file test ke direktori sementara
	for _, f := range testFiles {
		path := filepath.Join(tmpDir, f.Name)
		if err := os.WriteFile(path, []byte(f.Content), 0644); err != nil {
			return nil, fmt.Errorf("write test file %s: %w", f.Name, err)
		}
	}

	// Buat konfigurasi Playwright dengan JSON reporter + video recording
	escapedURL, _ := json.Marshal(projectURL)
	config := `import { defineConfig } from '@playwright/test';
export default defineConfig({
  reporter: [['json', { outputFile: '/results/results.json' }]],
  use: {
    baseURL: ` + string(escapedURL) + `,
    video: { mode: 'on', size: { width: 1280, height: 720 } },
  },
});`
	if err := os.WriteFile(filepath.Join(tmpDir, "playwright.config.ts"), []byte(config), 0644); err != nil {
		return nil, fmt.Errorf("write config: %w", err)
	}

	// Jalankan test di Docker container
	// Use host.docker.internal for host access (Docker Desktop Mac/Win, or --add-host on Linux)
	// instead of --network host, which grants full host network access.
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
		"-v", tmpDir+":/tests",
		"-v", tmpDir+":/results",
		"--add-host", "host.docker.internal:host-gateway",
		r.image,
		"npx", "playwright", "test", "--config=/tests/playwright.config.ts",
	)
	cmd.Dir = tmpDir
	output, _ := cmd.CombinedOutput()

	// Parse hasil dari JSON reporter
	resultsPath := filepath.Join(tmpDir, "results.json")

	// Emit step-level events dari Playwright report
	if r.exec != nil && r.runID != "" {
		reporter.ParseAndEmit(r.exec, r.runID, resultsPath)
	}

	result, err := parsePlaywrightResults(resultsPath, string(output))
	if err != nil {
		return result, err
	}

	// Cari file video yang dihasilkan Playwright (biasanya di test-results/)
	videoPath := findVideoFile(tmpDir)
	if videoPath != "" && r.runID != "" {
		// Salin video ke direktori persisten
		destDir := filepath.Join("/data/videos", r.runID)
		os.MkdirAll(destDir, 0755)
		destPath := filepath.Join(destDir, "recording.webm")
		if data, err := os.ReadFile(videoPath); err == nil {
			os.WriteFile(destPath, data, 0644)
			result.VideoPath = "/videos/" + r.runID + "/recording.webm"
		}
	}

	return result, nil
}

// Struktur untuk parsing JSON report Playwright
type playwrightReport struct {
	Suites []playwrightSuite `json:"suites"`
	Stats  struct {
		Expected   int `json:"expected"`   // Test yang pass
		Unexpected int `json:"unexpected"` // Test yang gagal
		Skipped    int `json:"skipped"`    // Test yang dilewati
	} `json:"stats"`
}

type playwrightSuite struct {
	Title string           `json:"title"`
	Specs []playwrightSpec `json:"specs"`
}

type playwrightSpec struct {
	Title string `json:"title"`
	Tests []struct {
		Status  string `json:"status"`
		Results []struct {
			Status string `json:"status"`
			Error  *struct {
				Message string `json:"message"`
			} `json:"error"`
		} `json:"results"`
	} `json:"tests"`
}

// parsePlaywrightResults membaca dan parse file JSON report Playwright
func parsePlaywrightResults(path, output string) (*agent.RunResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// Fallback: parse dari stdout jika file tidak ada
		return parseFromOutput(output), nil
	}

	// Bersihkan markdown fence jika ada
	data = []byte(stripJSONMarkers(string(data)))

	var report playwrightReport
	if err := json.Unmarshal(data, &report); err != nil {
		return parseFromOutput(output), nil
	}

	result := &agent.RunResult{
		Passed: report.Stats.Expected,
		Failed: report.Stats.Unexpected,
		Total:  report.Stats.Expected + report.Stats.Unexpected + report.Stats.Skipped,
	}

	// Kumpulkan detail kegagalan dari setiap test
	for _, suite := range report.Suites {
		for _, spec := range suite.Specs {
			for _, test := range spec.Tests {
				for _, r := range test.Results {
					if r.Status == "failed" || r.Status == "unexpected" {
						msg := ""
						if r.Error != nil {
							msg = r.Error.Message
						}
						result.Failures = append(result.Failures, agent.Failure{
							Test:    spec.Title,
							Message: msg,
						})
					}
				}
			}
		}
	}

	return result, nil
}

// parseFromOutput adalah fallback parser dari stdout Playwright
func parseFromOutput(output string) *agent.RunResult {
	result := &agent.RunResult{}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "passed") {
			result.Passed++
		}
		if strings.Contains(line, "failed") {
			result.Failed++
			result.Failures = append(result.Failures, agent.Failure{
				Test:    "unknown",
				Message: line,
			})
		}
	}
	result.Total = result.Passed + result.Failed
	return result
}

// findVideoFile mencari file video (.webm) di direktori test results
func findVideoFile(dir string) string {
	var found string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".webm") || strings.HasSuffix(path, ".mp4") {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

// stripJSONMarkers membersihkan markdown code fence
func stripJSONMarkers(s string) string {
	s = strings.TrimPrefix(s, "```json\n")
	s = strings.TrimPrefix(s, "```\n")
	s = strings.TrimSuffix(s, "\n```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
