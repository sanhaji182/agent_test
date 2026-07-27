package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
)

// TypeRunByID is the job type for executing an already-persisted run by ID.
// Unlike TypeTestRun (which carries the full payload), this references a run
// created by the API server so state/events remain consistent (ADR-001).
const TypeRunByID = "runs:execute"

// RunByIDPayload carries only the run ID; the worker loads the run from the store.
type RunByIDPayload struct {
	RunID string `json:"run_id"`
}

// RunExecutor executes a persisted run by ID. Implemented by api.Server.
type RunExecutor interface {
	ExecuteRunByID(ctx context.Context, runID string) error
}

// RunWorker processes runs:execute jobs from Redis.
type RunWorker struct {
	srv  *asynq.Server
	exec RunExecutor
}

// NewRunWorker creates a worker bound to a RunExecutor.
func NewRunWorker(redisAddr string, exec RunExecutor, concurrency int) *RunWorker {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: concurrency},
	)
	return &RunWorker{srv: srv, exec: exec}
}

// Start begins processing jobs (non-blocking).
func (w *RunWorker) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeRunByID, w.handleRunByID)
	return w.srv.Start(mux)
}

// Stop gracefully stops the worker.
func (w *RunWorker) Stop() {
	w.srv.Stop()
	w.srv.Shutdown()
}

func (w *RunWorker) handleRunByID(ctx context.Context, t *asynq.Task) error {
	var payload RunByIDPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}
	slog.Info("queue: executing run", "run_id", payload.RunID)
	if err := w.exec.ExecuteRunByID(ctx, payload.RunID); err != nil {
		slog.Error("queue: run failed", "run_id", payload.RunID, "error", err)
		return err
	}
	slog.Info("queue: run completed", "run_id", payload.RunID)
	return nil
}

// EnqueueRunByID enqueues a run-execution job referencing a persisted run.
func EnqueueRunByID(client *asynq.Client, runID string) error {
	data, err := json.Marshal(RunByIDPayload{RunID: runID})
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeRunByID, data)
	_, err = client.Enqueue(task, asynq.MaxRetry(2), asynq.Timeout(15*time.Minute))
	return err
}
