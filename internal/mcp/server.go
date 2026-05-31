package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	mcpServer *server.MCPServer
	agent     *agent.Agent
	llm       agent.LLM
	runs      map[string]*agent.TestRun
}

func NewServer(a *agent.Agent, llm agent.LLM) *Server {
	s := server.NewMCPServer("gotest-agent", "1.0.0")

	srv := &Server{mcpServer: s, agent: a, llm: llm, runs: make(map[string]*agent.TestRun)}
	srv.registerTools()
	return srv
}

func (s *Server) registerTools() {
	s.mcpServer.AddTool(
		mcp.NewTool("run_tests",
			mcp.WithDescription("Run tests on a project"),
			mcp.WithString("project_path", mcp.Required(), mcp.Description("Path to project")),
			mcp.WithString("requirements", mcp.Description("Test requirements")),
		),
		s.handleRunTests,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("analyze_project",
			mcp.WithDescription("Analyze a project codebase"),
			mcp.WithString("project_path", mcp.Required(), mcp.Description("Path to project")),
		),
		s.handleAnalyzeProject,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("generate_test_plan",
			mcp.WithDescription("Generate a test plan for a project"),
			mcp.WithString("project_path", mcp.Required(), mcp.Description("Path to project")),
			mcp.WithString("requirements", mcp.Description("Test requirements")),
		),
		s.handleGenerateTestPlan,
	)

	s.mcpServer.AddTool(
		mcp.NewTool("get_run_status",
			mcp.WithDescription("Get status of a test run"),
			mcp.WithString("run_id", mcp.Required(), mcp.Description("Run ID")),
		),
		s.handleGetRunStatus,
	)
}

func (s *Server) handleRunTests(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, _ := req.Params.Arguments["project_path"].(string)
	reqs, _ := req.Params.Arguments["requirements"].(string)

	run := &agent.TestRun{
		ID:           uuid.New().String(),
		ProjectPath:  path,
		Requirements: reqs,
		State:        agent.StateIdle,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	s.runs[run.ID] = run

	if err := s.agent.Execute(ctx, run); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("run failed: %v", err)), nil
	}

	result, _ := json.Marshal(run)
	return mcp.NewToolResultText(string(result)), nil
}

func (s *Server) handleAnalyzeProject(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, _ := req.Params.Arguments["project_path"].(string)

	analysis, err := s.llm.AnalyzeCodebase(ctx, path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("analysis failed: %v", err)), nil
	}

	return mcp.NewToolResultText(analysis), nil
}

func (s *Server) handleGenerateTestPlan(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, _ := req.Params.Arguments["project_path"].(string)
	reqs, _ := req.Params.Arguments["requirements"].(string)

	analysis, err := s.llm.AnalyzeCodebase(ctx, path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("analysis failed: %v", err)), nil
	}

	plan, err := s.llm.GenerateTestPlan(ctx, analysis, reqs)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("plan generation failed: %v", err)), nil
	}

	result, _ := json.Marshal(plan)
	return mcp.NewToolResultText(string(result)), nil
}

func (s *Server) handleGetRunStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	runID, _ := req.Params.Arguments["run_id"].(string)

	run, ok := s.runs[runID]
	if !ok {
		return mcp.NewToolResultError(fmt.Sprintf("run %s not found", runID)), nil
	}

	result, _ := json.Marshal(run)
	return mcp.NewToolResultText(string(result)), nil
}

func (s *Server) Serve() error {
	return server.ServeStdio(s.mcpServer)
}
