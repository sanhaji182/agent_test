package api

import (
	"context"
	"log/slog"

	"github.com/go-go-golems/gotest-agent/internal/github"
	"github.com/go-go-golems/gotest-agent/internal/parser"
	"github.com/go-go-golems/gotest-agent/internal/webhook"
)

const webhookCloneDir = "/tmp/agent_test/repos"

// processPushWithTestGen runs the AI test-generation pipeline (clone, parse,
// synthesize test plan, launch run) for a GitHub push event. Returns false if
// AI planning is unavailable or the pipeline fails, so the caller can fall
// back to a plain auto-triggered run.
func (s *Server) processPushWithTestGen(event webhook.PushEvent) bool {
	ctx := context.Background()
	llm := s.aiClient(ctx)
	if llm == nil {
		return false
	}

	integration := github.NewIntegration(s.cfg.GitHubWebhookSecret, webhookCloneDir)
	svc := github.NewTestGenerationService(integration, parser.NewDefaultRegistry(), llm, s.launchRun, s.store)

	ghEvent := &github.PushEvent{
		Ref: event.Ref,
		Repository: github.Repository{
			FullName: event.Repository.FullName,
			CloneURL: event.Repository.CloneURL,
		},
		Commits: make([]github.Commit, 0, len(event.Commits)),
	}
	for _, c := range event.Commits {
		ghEvent.Commits = append(ghEvent.Commits, github.Commit{
			ID:       c.ID,
			Message:  c.Message,
			Added:    c.Added,
			Removed:  c.Removed,
			Modified: c.Modified,
		})
	}

	if err := svc.ProcessPushEvent(ctx, ghEvent); err != nil {
		slog.Warn("webhook test generation failed, falling back to plain run",
			"repo", event.Repository.FullName, "ref", event.Ref, "error", err)
		return false
	}
	return true
}
