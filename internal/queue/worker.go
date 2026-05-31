// Package queue menyediakan job queue menggunakan Asynq (Redis-backed).
// Digunakan untuk menjalankan test secara async di background.
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// TypeTestRun adalah tipe job untuk menjalankan test
const TypeTestRun = "tests:run"

// TestRunPayload adalah data yang dikirim bersama job
type TestRunPayload struct {
	ProjectPath  string `json:"project_path"`
	Requirements string `json:"requirements"`
	ProjectURL   string `json:"project_url"`
}

// Worker memproses job dari Redis queue
type Worker struct {
	srv   *asynq.Server
	agent *agent.Agent
}

// NewWorker membuat worker baru dengan koneksi Redis dan agent
func NewWorker(redisAddr string, a *agent.Agent, concurrency int) *Worker {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: concurrency},
	)
	return &Worker{srv: srv, agent: a}
}

// Start memulai worker untuk memproses job
func (w *Worker) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeTestRun, w.handleTestRun)
	return w.srv.Start(mux)
}

// Stop menghentikan worker
func (w *Worker) Stop() {
	w.srv.Stop()
}

// handleTestRun memproses satu job test run
func (w *Worker) handleTestRun(_ context.Context, t *asynq.Task) error {
	var payload TestRunPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	ctx := context.Background()
	run := &agent.TestRun{
		ID:           uuid.New().String(),
		ProjectPath:  payload.ProjectPath,
		Requirements: payload.Requirements,
		State:        agent.StateIdle,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	slog.Info("starting test run", "id", run.ID, "path", payload.ProjectPath)
	if err := w.agent.Execute(ctx, run); err != nil {
		slog.Error("test run failed", "id", run.ID, "error", err)
		return err
	}

	slog.Info("test run completed", "id", run.ID, "state", run.State)
	return nil
}

// EnqueueTestRun menambahkan job test run ke queue
func EnqueueTestRun(client *asynq.Client, payload TestRunPayload) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	task := asynq.NewTask(TypeTestRun, data)
	info, err := client.Enqueue(task, asynq.MaxRetry(3), asynq.Timeout(10*time.Minute))
	if err != nil {
		return "", err
	}
	return info.ID, nil
}
