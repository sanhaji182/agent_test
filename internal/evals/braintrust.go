// Package evals menyediakan integrasi dengan Braintrust untuk evaluasi kualitas LLM.
// Mencatat skor untuk: test plan quality, script validity, fix success rate.
package evals

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Logger mencatat hasil evaluasi ke Braintrust API
type Logger struct {
	apiKey  string
	project string
	baseURL string
	http    *http.Client
}

// LogEntry adalah satu entri evaluasi yang dikirim ke Braintrust
type LogEntry struct {
	ID        string             `json:"id"`
	Project   string             `json:"project"`
	Input     any                `json:"input"`
	Output    any                `json:"output"`
	Expected  any                `json:"expected,omitempty"`
	Scores    map[string]float64 `json:"scores,omitempty"`
	Metadata  map[string]any     `json:"metadata,omitempty"`
	CreatedAt string             `json:"created_at"`
}

// NewLogger membuat eval logger baru (no-op jika apiKey kosong)
func NewLogger(apiKey, project string) *Logger {
	return &Logger{
		apiKey:  apiKey,
		project: project,
		baseURL: "https://api.braintrust.dev/v1",
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// LogTestPlanGeneration mencatat kualitas test plan yang dihasilkan (skor 0-10)
func (l *Logger) LogTestPlanGeneration(ctx context.Context, input any, output any, score float64) error {
	return l.log(ctx, LogEntry{
		Project:  l.project,
		Input:    input,
		Output:   output,
		Scores:   map[string]float64{"test_plan_quality": score},
		Metadata: map[string]any{"type": "test_plan_generation"},
	})
}

// LogScriptGeneration mencatat apakah script yang dihasilkan valid
func (l *Logger) LogScriptGeneration(ctx context.Context, input any, output any, passed bool) error {
	score := 0.0
	if passed {
		score = 1.0
	}
	return l.log(ctx, LogEntry{
		Project:  l.project,
		Input:    input,
		Output:   output,
		Scores:   map[string]float64{"script_validity": score},
		Metadata: map[string]any{"type": "script_generation"},
	})
}

// LogFixAttempt mencatat apakah fix berhasil memperbaiki test yang gagal
func (l *Logger) LogFixAttempt(ctx context.Context, input any, output any, succeeded bool) error {
	score := 0.0
	if succeeded {
		score = 1.0
	}
	return l.log(ctx, LogEntry{
		Project:  l.project,
		Input:    input,
		Output:   output,
		Scores:   map[string]float64{"fix_success": score},
		Metadata: map[string]any{"type": "fix_attempt"},
	})
}

// log mengirim entri evaluasi ke Braintrust API
func (l *Logger) log(ctx context.Context, entry LogEntry) error {
	if l.apiKey == "" {
		return nil // No-op jika tidak dikonfigurasi
	}

	entry.CreatedAt = time.Now().Format(time.RFC3339)
	body, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", l.baseURL+"/project_logs/"+l.project+"/insert", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+l.apiKey)

	resp, err := l.http.Do(req)
	if err != nil {
		return fmt.Errorf("braintrust log: %w", err)
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("braintrust: status %d", resp.StatusCode)
	}
	return nil
}
