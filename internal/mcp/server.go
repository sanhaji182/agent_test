// Package mcp menyediakan MCP (Model Context Protocol) server untuk integrasi IDE.
// Tools: run_tests, analyze_project, generate_test_plan, get_run_status
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

// Server adalah MCP server yang mengekspos tools ke IDE (Cursor, VS Code)
type Server struct {
	mcpServer *server.MCPServer
	agent     *agent.Agent
	llm       agent.LLM
	runs      map[string]*agent.TestRun // Penyimpanan run di memori
}

// NewServer membuat MCP server baru dengan agent dan LLM
func NewServer(a *agent.Agent, llm agent.LLM) *Server {
	s := server.NewMCPServer("gotest-agent", "1.0.0")
	srv := &Server{mcpServer: s, agent: a, llm: llm, runs: make(map[string]*agent.TestRun)}
	srv.registerTools()
	return srv
}

// registerTools mendaftarkan semua MCP tools
func (s *Server) registerTools() {
	// Tool: jalankan test lengkap
	s.mcpServer.AddTool(
		mcp.NewTool("run_tests",
			mcp.WithDescription("Run tests on a project"),
			mcp.WithString("project_path", mcp.Required(), mcp.Description("Path to project")),
			mcp.WithString("requirements", mcp.Description("Test requirements")),
		),
		s.handleRunTests,
	)

	// Tool: analisis kode project
	s.mcpServer.AddTool(
		mcp.NewTool("analyze_project",
			mcp.WithDescription("Analyze a project codebase"),
			mcp.WithString("project_path", mcp.Required(), mcp.Description("Path to project")),
		),
		s.handleAnalyzeProject,
	)

	// Tool: buat test plan
	s.mcpServer.AddTool(
		mcp.NewTool("generate_test_plan",
			mcp.WithDescription("Generate a test plan for a project"),
			mcp.WithString("project_path", mcp.Required(), mcp.Description("Path to project")),
			mcp.WithString("requirements", mcp.Description("Test requirements")),
		),
		s.handleGenerateTestPlan,
	)

	// Tool: cek status run
	s.mcpServer.AddTool(
		mcp.NewTool("get_run_status",
			mcp.WithDescription("Get status of a test run"),
			mcp.WithString("run_id", mcp.Required(), mcp.Description("Run ID")),
		),
		s.handleGetRunStatus,
	)
}

// handleRunTests menjalankan full test pipeline dan mengembalikan hasil
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

// handleAnalyzeProject menganalisis kode project via LLM
func (s *Server) handleAnalyzeProject(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, _ := req.Params.Arguments["project_path"].(string)

	analysis, err := s.llm.AnalyzeCodebase(ctx, path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("analysis failed: %v", err)), nil
	}
	return mcp.NewToolResultText(analysis), nil
}

// handleGenerateTestPlan membuat test plan dari analisis + requirements
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

// handleGetRunStatus mengembalikan status run berdasarkan ID
func (s *Server) handleGetRunStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	runID, _ := req.Params.Arguments["run_id"].(string)

	run, ok := s.runs[runID]
	if !ok {
		return mcp.NewToolResultError(fmt.Sprintf("run %s not found", runID)), nil
	}

	result, _ := json.Marshal(run)
	return mcp.NewToolResultText(string(result)), nil
}

// Serve memulai MCP server via stdio (untuk koneksi dari IDE)
func (s *Server) Serve() error {
	return server.ServeStdio(s.mcpServer)
}
