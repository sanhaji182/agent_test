# AGENT.md — GoTest Agent Execution Guide

You are an autonomous coding agent working on the GoTest Agent project.
Your job is to implement the full blueprint in gotest-agent-master-blueprint.md.

## Core Objective

Build a self-hostable AI software testing agent in Go with:
- MCP server for Cursor/VS Code integration
- HTTP API + async queue
- Anthropic Claude-based code analysis and test generation
- Steel Browser self-hosted sandbox for browser automation
- Playwright test generation and execution
- Visual regression and screenshot analysis
- LangGraph sidecar for advanced multi-agent workflows
- Braintrust-based evals for quality tracking
- PostgreSQL persistence and Redis queue

## Operating Rules

1. Read gotest-agent-master-blueprint.md before changing code.
2. Follow the blueprint in order, starting from Section 12.
3. Do not wait for human confirmation between tasks unless blocked by ambiguity.
4. Prefer making the code compile and run over perfect architecture.
5. If a file is broken, rewrite it fully and continue.
6. After each phase, summarize what changed and what remains.
7. Keep the implementation production-oriented but MVP-first.
8. If a dependency or API is uncertain, inspect docs and implement the simplest working version.
9. Keep all changes aligned with the blueprint; do not invent unrelated features.
10. Treat the blueprint as the source of truth.

## Required Execution Order

### Phase 1 — Foundation
- Fix module name in go.mod
- Install dependencies
- Run go mod tidy
- Rewrite internal/agent/llm_anthropic.go using anthropic-sdk-go
- Create internal/steel/client.go
- Fix internal/mcp/server.go imports and tool wiring
- Fix internal/runner/docker.go JSON parsing
- Ensure go build ./... passes

### Phase 2 — E2E MVP
- Add Steel Browser service to docker-compose
- Add DB store implementation
- Add migration runner
- Persist test runs and results
- Add SSE progress stream
- Add HTML report generation
- Ensure MCP run_tests works end-to-end

### Phase 3 — Vision + Evals
- Add screenshot capture
- Add GPT-4o Vision integration
- Add Braintrust eval logging
- Add visual regression checks

### Phase 4 — LangGraph Sidecar
- Create Python sidecar service
- Add FastAPI server
- Implement planner/writer/critic/executor/fixer agents
- Wire Go backend to sidecar for advanced mode

### Phase 5 — SaaS Layer
- Add auth, API keys, dashboard, webhook integration, billing

## Work Style

- Be aggressive and autonomous.
- Make sensible decisions without asking unless absolutely necessary.
- Fix compile/runtime errors as soon as they appear.
- Keep the implementation incremental and test frequently.

## First Actions

1. Inspect the repository structure.
2. Read the blueprint.
3. Start Phase 1 immediately.
4. Report progress only after meaningful chunks of work.
