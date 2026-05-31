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

const TypeTestRun = "tests:run"

type TestRunPayload struct {
	ProjectPath  string `json:"project_path"`
	Requirements string `json:"requirements"`
	ProjectURL   string `json:"project_url"`
}

type Worker struct {
	srv   *asynq.Server
	agent *agent.Agent
}

func NewWorker(redisAddr string, a *agent.Agent, concurrency int) *Worker {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: concurrency},
	)
	return &Worker{srv: srv, agent: a}
}

func (w *Worker) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeTestRun, w.handleTestRun)
	return w.srv.Start(mux)
}

func (w *Worker) Stop() {
	w.srv.Stop()
}

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
