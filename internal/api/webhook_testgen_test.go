package api

import (
	"context"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/config"
	"github.com/go-go-golems/gotest-agent/internal/db"
	"github.com/go-go-golems/gotest-agent/internal/webhook"
)

func TestProcessPushWithTestGenDisabledWithoutAIPlanning(t *testing.T) {
	t.Setenv("GOTEST_AI_PLANNING", "")
	store := db.NewMemoryStore()
	s := NewServer(&config.Config{MaxConcurrentRuns: 1}, store, nil)

	event := webhook.PushEvent{
		Ref:        "refs/heads/main",
		Repository: webhook.Repository{FullName: "acme/app", CloneURL: "https://example.com/acme/app.git"},
	}
	if s.processPushWithTestGen(event) {
		t.Fatal("expected fallback path when AI planning is disabled")
	}

	runs, err := store.ListRuns(context.Background(), 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 0 {
		t.Fatalf("test-gen skip must not create runs, got %d", len(runs))
	}
}
