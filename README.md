# GoTest Agent

Self-hosted AI testing agent that reads your codebase, generates tests, executes them, and auto-fixes failures — all from your IDE or a web dashboard.

## What it does

1. **Analyzes** your project (language, framework, routes, endpoints)
2. **Generates** a test plan with prioritized scenarios
3. **Writes** Playwright test files automatically
4. **Executes** tests in a sandboxed Steel Browser
5. **Auto-fixes** failing tests (up to 3 attempts)
6. **Reports** results with screenshots, HTML reports, and live streaming

## Quick Start

```bash
# Clone and start the full stack
git clone https://github.com/sanhaji182/agent_test.git
cd agent_test
cp .env.example .env
# Add your ANTHROPIC_API_KEY to .env

# Start everything
make up

# Verify
make smoke-test
```

Open http://localhost:3001 for the dashboard, http://localhost:8080 for the API.

## Local Development

```bash
# Backend
go run ./cmd/server    # API on :8080
go run ./cmd/mcp       # MCP server (stdio)

# Frontend
cd frontend && npm install && npm run dev  # Dashboard on :3000
```

## Docker Full Stack

```bash
make up        # Start all 6 services
make down      # Stop
make logs      # Tail logs
make rebuild   # Rebuild and restart
make ps        # Show status
make smoke-test # End-to-end verification
```

Services: backend (:8080), frontend (:3001), postgres (:5432), redis (:6379), steel-browser (:3010), langgraph-sidecar (:8000)

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `ANTHROPIC_API_KEY` | Yes | Anthropic API key for LLM |
| `LLM_MODEL` | No | Model (default: claude-sonnet-4-5) |
| `API_KEY` | No | API auth key (empty = no auth) |
| `DATABASE_URL` | No | PostgreSQL URL (falls back to in-memory) |
| `REDIS_URL` | No | Redis for job queue |
| `STEEL_API_URL` | No | Steel Browser URL |
| `OPENAI_API_KEY` | No | For visual regression (GPT-4o Vision) |

## How to Use

### Connect a Project

```bash
curl -X POST http://localhost:8080/api/v1/runs \
  -H "Content-Type: application/json" \
  -d '{"project_path": "/path/to/your/project", "requirements": "test login and checkout flows"}'
```

Or use the MCP tool in Cursor/VS Code:
```
run_tests(project_path="/path/to/project", requirements="test login")
```

### Run Tests

Via API:
```bash
curl -X POST http://localhost:8080/api/v1/runs \
  -d '{"project_path": "/app", "requirements": "test all user flows"}'
```

Via MCP (in IDE): type "test my project" and the agent handles everything.

### Set Up Schedules

```bash
curl -X POST http://localhost:8080/api/v1/schedules \
  -d '{
    "name": "nightly",
    "project_path": "/app",
    "requirements": "regression tests",
    "frequency": "daily",
    "environment": "staging",
    "base_url": "http://staging.example.com",
    "enabled": true,
    "notify_on_fail": true,
    "webhook_url": "https://hooks.slack.com/..."
  }'
```

Supported frequencies: `daily`, `weekly`, `monthly`, `cron` (with `cron_expr`).

### Use Risk & Recommendations

```bash
# Get risk scores
curl http://localhost:8080/api/v1/metrics/risk

# Get actionable recommendations
curl http://localhost:8080/api/v1/metrics/recommendations

# Smart suite selection
curl -X POST http://localhost:8080/api/v1/suite-selection \
  -d '{"mode": "high_risk", "all_tests": ["login", "checkout", "signup"]}'
```

### Compare Runs

```bash
curl http://localhost:8080/api/v1/runs/{runA}/compare/{runB}
```

### Review Workflow

```bash
# Create review for a test plan
curl -X POST http://localhost:8080/api/v1/reviews \
  -d '{"run_id": "...", "type": "test_plan"}'

# Approve
curl -X POST http://localhost:8080/api/v1/reviews/{id}/approve \
  -d '{"reviewer": "alice", "comment": "LGTM"}'
```

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│  Frontend   │────▶│  Go Backend  │────▶│  Steel Browser   │
│  (Next.js)  │     │  (Chi API)   │     │  (Playwright)    │
└─────────────┘     └──────┬───────┘     └─────────────────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
        ┌──────────┐ ┌─────────┐ ┌──────────────┐
        │PostgreSQL│ │  Redis  │ │LangGraph     │
        │          │ │         │ │Sidecar       │
        └──────────┘ └─────────┘ └──────────────┘
```

## API Reference

See the full endpoint list at `GET /health` or browse the source in `internal/api/server.go`.

Key endpoints:
- `POST /api/v1/runs` — create test run
- `GET /api/v1/runs/:id` — get run detail
- `GET /api/v1/runs/:id/stream` — SSE live events
- `GET /api/v1/runs/:id/events` — step-level events
- `POST /api/v1/runs/:id/rerun` — rerun
- `GET /api/v1/runs/:id/compare/:other` — compare
- `GET /api/v1/metrics/risk` — risk scores
- `GET /api/v1/metrics/recommendations` — suggestions
- `POST /api/v1/schedules` — create schedule
- `POST /api/v1/releases` — create release
- `GET /api/v1/releases/:id/confidence` — confidence score

## License

MIT
