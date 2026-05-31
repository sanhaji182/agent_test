package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SidecarClient adalah HTTP client untuk berkomunikasi dengan LangGraph Python sidecar
type SidecarClient struct {
	baseURL string
	http    *http.Client
}

// NewSidecarClient membuat client baru untuk sidecar di URL tertentu
func NewSidecarClient(baseURL string) *SidecarClient {
	return &SidecarClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 5 * time.Minute},
	}
}

// Request body untuk memulai run di sidecar
type sidecarRunRequest struct {
	RunID        string `json:"run_id"`
	ProjectPath  string `json:"project_path"`
	Requirements string `json:"requirements"`
	CodeAnalysis string `json:"code_analysis"`
	MaxFixes     int    `json:"max_fixes"`
}

// Response dari sidecar saat memulai run
type sidecarRunResponse struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

// Response status dari sidecar
type sidecarStatusResponse struct {
	Status string `json:"status"` // "running", "completed", "failed"
	Result any    `json:"result"`
	Error  string `json:"error"`
}

// StartRun memulai eksekusi multi-agent di sidecar (async)
func (c *SidecarClient) StartRun(ctx context.Context, run *TestRun, maxFixes int) (string, error) {
	body, _ := json.Marshal(sidecarRunRequest{
		RunID:        run.ID,
		ProjectPath:  run.ProjectPath,
		Requirements: run.Requirements,
		CodeAnalysis: run.CodeAnalysis,
		MaxFixes:     maxFixes,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/agent/run", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("sidecar request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("sidecar: status %d: %s", resp.StatusCode, string(b))
	}

	var result sidecarRunResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.JobID, nil
}

// GetStatus mengecek status job yang sedang berjalan di sidecar
func (c *SidecarClient) GetStatus(ctx context.Context, jobID string) (*sidecarStatusResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/agent/"+jobID, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var status sidecarStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}
	return &status, nil
}
