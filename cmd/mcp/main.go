package main

import (
	"log/slog"
	"os"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/config"
	mcpserver "github.com/go-go-golems/gotest-agent/internal/mcp"
	"github.com/go-go-golems/gotest-agent/internal/runner"
)

func main() {
	cfg := config.Load()

	llm := agent.NewAnthropicLLM(cfg.AnthropicAPIKey, cfg.LLMModel)
	r := runner.NewDockerRunner(cfg.TimeoutSeconds)
	a := agent.New(llm, r, cfg.MaxFixAttempts)

	srv := mcpserver.NewServer(a, llm)
	slog.Info("starting MCP server (stdio)")
	if err := srv.Serve(); err != nil {
		slog.Error("mcp server failed", "error", err)
		os.Exit(1)
	}
}
