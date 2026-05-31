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

type Logger struct {
	apiKey  string
	project string
	baseURL string
	http    *http.Client
}

type LogEntry struct {
	ID        string         `json:"id"`
	Project   string         `json:"project"`
	Input     any            `json:"input"`
	Output    any            `json:"output"`
	Expected  any            `json:"expected,omitempty"`
	Scores    map[string]float64 `json:"scores,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt string         `json:"created_at"`
}

func NewLogger(apiKey, project string) *Logger {
	baseURL := "https://api.braintrust.dev/v1"
	return &Logger{
		apiKey:  apiKey,
		project: project,
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (l *Logger) LogTestPlanGeneration(ctx context.Context, input any, output any, score float64) error {
	return l.log(ctx, LogEntry{
		Project: l.project,
		Input:   input,
		Output:  output,
		Scores:  map[string]float64{"test_plan_quality": score},
		Metadata: map[string]any{"type": "test_plan_generation"},
	})
}

func (l *Logger) LogScriptGeneration(ctx context.Context, input any, output any, passed bool) error {
	score := 0.0
	if passed {
		score = 1.0
	}
	return l.log(ctx, LogEntry{
		Project: l.project,
		Input:   input,
		Output:  output,
		Scores:  map[string]float64{"script_validity": score},
		Metadata: map[string]any{"type": "script_generation"},
	})
}

func (l *Logger) LogFixAttempt(ctx context.Context, input any, output any, succeeded bool) error {
	score := 0.0
	if succeeded {
		score = 1.0
	}
	return l.log(ctx, LogEntry{
		Project: l.project,
		Input:   input,
		Output:  output,
		Scores:  map[string]float64{"fix_success": score},
		Metadata: map[string]any{"type": "fix_attempt"},
	})
}

func (l *Logger) log(ctx context.Context, entry LogEntry) error {
	if l.apiKey == "" {
		return nil // No-op if not configured
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
