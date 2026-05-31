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

// SteelRunner menjalankan Playwright test melalui Steel Browser CDP endpoint
type SteelRunner struct {
	steel   *steel.Client
	timeout int
}

// NewSteelRunner membuat runner yang terhubung ke Steel Browser
func NewSteelRunner(steelClient *steel.Client, timeout int) *SteelRunner {
	return &SteelRunner{steel: steelClient, timeout: timeout}
}

// Run membuat sesi Steel, menjalankan test via CDP, dan mengembalikan hasil
func (r *SteelRunner) Run(ctx context.Context, testFiles []agent.TestFile, projectURL string) (*agent.RunResult, error) {
	// Buat sesi browser baru di Steel
	session, err := r.steel.CreateSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("create steel session: %w", err)
	}
	defer r.steel.DestroySession(ctx, session.ID) // Bersihkan setelah selesai

	// Siapkan direktori sementara untuk file test
	tmpDir, err := os.MkdirTemp("", "gotest-steel-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Tulis file test ke disk
	for _, f := range testFiles {
		if err := os.WriteFile(filepath.Join(tmpDir, f.Name), []byte(f.Content), 0644); err != nil {
			return nil, fmt.Errorf("write %s: %w", f.Name, err)
		}
	}

	// Konfigurasi Playwright untuk terhubung ke Steel via CDP WebSocket
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

	// Jalankan Playwright test
	cmd := exec.CommandContext(ctx, "npx", "playwright", "test",
		"--config="+filepath.Join(tmpDir, "playwright.config.ts"))
	cmd.Dir = tmpDir
	cmd.CombinedOutput()

	// Parse hasil dari JSON reporter
	return parsePlaywrightResults(filepath.Join(tmpDir, "results.json"), "")
}
