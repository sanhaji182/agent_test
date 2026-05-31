package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/steel"
)

type SteelRunner struct {
	steel   *steel.Client
	timeout int
}

func NewSteelRunner(steelClient *steel.Client, timeout int) *SteelRunner {
	return &SteelRunner{steel: steelClient, timeout: timeout}
}

func (r *SteelRunner) Run(ctx context.Context, testFiles []agent.TestFile, projectURL string) (*agent.RunResult, error) {
	session, err := r.steel.CreateSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("create steel session: %w", err)
	}
	defer r.steel.DestroySession(ctx, session.ID)

	tmpDir, err := os.MkdirTemp("", "gotest-steel-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	for _, f := range testFiles {
		if err := os.WriteFile(filepath.Join(tmpDir, f.Name), []byte(f.Content), 0644); err != nil {
			return nil, fmt.Errorf("write %s: %w", f.Name, err)
		}
	}

	// Playwright config connecting to Steel CDP
	config := fmt.Sprintf(`import { defineConfig, devices } from '@playwright/test';
export default defineConfig({
  reporter: [['json', { outputFile: '%s/results.json' }]],
  use: {
    baseURL: '%s',
    connectOptions: { wsEndpoint: '%s' },
  },
});`, tmpDir, projectURL, session.CDPURL)

	if err := os.WriteFile(filepath.Join(tmpDir, "playwright.config.ts"), []byte(config), 0644); err != nil {
		return nil, fmt.Errorf("write config: %w", err)
	}

	cmd := exec.CommandContext(ctx, "npx", "playwright", "test",
		"--config="+filepath.Join(tmpDir, "playwright.config.ts"))
	cmd.Dir = tmpDir
	cmd.CombinedOutput()

	return parsePlaywrightResults(filepath.Join(tmpDir, "results.json"), "")
}
