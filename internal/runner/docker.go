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
)

type DockerRunner struct {
	image   string
	timeout int
}

func NewDockerRunner(timeout int) *DockerRunner {
	return &DockerRunner{
		image:   "mcr.microsoft.com/playwright:v1.40.0-jammy",
		timeout: timeout,
	}
}

func (r *DockerRunner) Run(ctx context.Context, testFiles []agent.TestFile, projectURL string) (*agent.RunResult, error) {
	tmpDir, err := os.MkdirTemp("", "gotest-run-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write test files
	for _, f := range testFiles {
		path := filepath.Join(tmpDir, f.Name)
		if err := os.WriteFile(path, []byte(f.Content), 0644); err != nil {
			return nil, fmt.Errorf("write test file %s: %w", f.Name, err)
		}
	}

	// Write playwright config
	config := `import { defineConfig } from '@playwright/test';
export default defineConfig({
  reporter: [['json', { outputFile: '/results/results.json' }]],
  use: { baseURL: '` + projectURL + `' },
});`
	if err := os.WriteFile(filepath.Join(tmpDir, "playwright.config.ts"), []byte(config), 0644); err != nil {
		return nil, fmt.Errorf("write config: %w", err)
	}

	// Run in Docker
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
		"-v", tmpDir+":/tests",
		"-v", tmpDir+":/results",
		"--network", "host",
		r.image,
		"npx", "playwright", "test", "--config=/tests/playwright.config.ts",
	)
	cmd.Dir = tmpDir
	output, _ := cmd.CombinedOutput()

	// Parse results
	resultsPath := filepath.Join(tmpDir, "results.json")
	return parsePlaywrightResults(resultsPath, string(output))
}

type playwrightReport struct {
	Suites []playwrightSuite `json:"suites"`
	Stats  struct {
		Expected   int `json:"expected"`
		Unexpected int `json:"unexpected"`
		Skipped    int `json:"skipped"`
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

func parsePlaywrightResults(path, output string) (*agent.RunResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// Fallback: parse from output
		return parseFromOutput(output), nil
	}

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

func stripJSONMarkers(s string) string {
	s = strings.TrimPrefix(s, "```json\n")
	s = strings.TrimPrefix(s, "```\n")
	s = strings.TrimSuffix(s, "\n```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
