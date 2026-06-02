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

## Security Considerations

This setup is optimized for **local development and demo**. For production deployment:

### 🔒 Required Changes

- **Change database credentials** — Update `POSTGRES_PASSWORD` in `docker-compose.yml` to a strong random value
- **Rotate secrets** — Generate strong random values for `JWT_SECRET`, `API_KEY`, and `GITHUB_WEBHOOK_SECRET` in your `.env`
- **Remove exposed ports** — Remove or bind database/Redis ports to `127.0.0.1`:
  ```yaml
  postgres:
    ports:
      - "127.0.0.1:5432:5432"  # instead of "5432:5432"
  redis:
    ports:
      - "127.0.0.1:6379:6379"  # instead of "6379:6379"
  ```
- **Add TLS termination** — Use a reverse proxy (Caddy, nginx, Traefik) with automatic HTTPS in front of the backend

### 🛡️ Recommended Enhancements

- **Restrict CORS** — Set `CORS_ALLOWED_ORIGINS` env var to your frontend domain(s) instead of wildcard `*`
- **Enable rate limiting** — Add rate limiting middleware to prevent API abuse (currently not implemented)
- **Use Docker secrets** — For sensitive values, use Docker secrets or external secret managers instead of env vars
- **Network isolation** — Consider running database/Redis on a private Docker network without host port exposure
- **Steel Browser hardening** — The `SYS_ADMIN` capability is required for headless Chrome but increases attack surface. In production:
  - Use a dedicated container runtime (gVisor, Firecracker)
  - Apply seccomp profiles to restrict syscalls
  - Monitor and limit concurrent sessions

### 📋 Security Checklist

Before deploying to production:

- [ ] Changed `POSTGRES_PASSWORD` from `password`
- [ ] Set strong `JWT_SECRET` (32+ random characters)
- [ ] Set strong `API_KEY` for backend authentication
- [ ] Removed or restricted exposed ports (5432, 6379)
- [ ] Configured reverse proxy with TLS (HTTPS)
- [ ] Set `CORS_ALLOWED_ORIGINS` to specific domains
- [ ] Reviewed and rotated all API keys (Anthropic, Steel, etc.)
- [ ] Enabled database SSL (`sslmode=require` in `DATABASE_URL`)
- [ ] Set up monitoring and logging for security events
- [ ] Regular security updates for base images

### 🔍 Automated Security Scanning

This project uses [gitleaks](https://github.com/gitleaks/gitleaks) to prevent accidental secret commits. See `.github/workflows/security.yml` for CI configuration.

### 🐛 Reporting Vulnerabilities

If you discover a security vulnerability, please email [security@yourdomain.com] instead of opening a public issue.

## License

MIT
