package api_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/api"
	"github.com/go-go-golems/gotest-agent/internal/config"
	"github.com/go-go-golems/gotest-agent/internal/db"
)

// TestLaunchRun_UsesEnqueuerWhenConfigured verifies that when a durable-queue
// enqueuer is installed via SetRunEnqueuer, creating a run routes through the
// enqueuer instead of executing in-process.
func TestLaunchRun_UsesEnqueuerWhenConfigured(t *testing.T) {
	store := db.NewMemoryStore()
	srv := api.NewServer(&config.Config{MaxConcurrentRuns: 2}, store, nil)

	enqueued := make(chan string, 1)
	srv.SetRunEnqueuer(func(runID string) error {
		enqueued <- runID
		return nil
	})

	body := strings.NewReader(`{"project_path": "https://example.com", "requirements": "smoke"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runs", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated && rec.Code != http.StatusAccepted && rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
	}

	select {
	case id := <-enqueued:
		if id == "" {
			t.Fatal("expected non-empty run ID enqueued")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected run to be enqueued via SetRunEnqueuer")
	}
}

// TestExecuteRunByID_TerminalStateIsNoop verifies idempotent retry safety:
// a run already in a terminal state is not re-executed.
func TestExecuteRunByID_TerminalStateIsNoop(t *testing.T) {
	store := db.NewMemoryStore()
	srv := api.NewServer(&config.Config{MaxConcurrentRuns: 2}, store, nil)

	run := &agent.TestRun{
		ID:          "done-run",
		ProjectPath: "https://example.com",
		State:       agent.StateDone,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := store.CreateRun(context.Background(), run); err != nil {
		t.Fatalf("create run: %v", err)
	}

	if err := srv.ExecuteRunByID(context.Background(), "done-run"); err != nil {
		t.Fatalf("expected noop for terminal run, got error: %v", err)
	}

	got, err := store.GetRun(context.Background(), "done-run")
	if err != nil {
		t.Fatalf("get run: %v", err)
	}
	if got.State != agent.StateDone {
		t.Fatalf("terminal state should be untouched, got %s", got.State)
	}
}

// TestExecuteRunByID_MissingRunErrors verifies unknown IDs surface an error
// (so Asynq marks the job failed instead of silently succeeding).
func TestExecuteRunByID_MissingRunErrors(t *testing.T) {
	srv := api.NewServer(&config.Config{MaxConcurrentRuns: 2}, db.NewMemoryStore(), nil)
	if err := srv.ExecuteRunByID(context.Background(), "nope"); err == nil {
		t.Fatal("expected error for missing run")
	}
}

// TestLaunchRun_EnqueuerErrorFallsBack verifies in-process fallback when the
// enqueuer fails: the run must not be lost (state eventually leaves idle or
// an unsupported-provider error is logged — here we just assert no panic and
// the request succeeds).
func TestLaunchRun_EnqueuerErrorFallsBack(t *testing.T) {
	store := db.NewMemoryStore()
	srv := api.NewServer(&config.Config{MaxConcurrentRuns: 2}, store, nil)
	srv.SetRunEnqueuer(func(runID string) error {
		return fmt.Errorf("redis down")
	})

	body := strings.NewReader(`{"project_path": "https://example.com", "requirements": "smoke"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runs", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code >= 500 {
		t.Fatalf("request must not fail when enqueue fails (fallback), got %d", rec.Code)
	}
}
