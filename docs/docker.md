# GoTest Agent — Docker Deployment

## Prerequisites

- Docker 24+ with Compose V2
- 4GB+ RAM (8GB recommended for Steel Browser)
- Ports available: 8080, 3001, 5432, 6379, 3010, 8000

## Quick Start

```bash
# 1. Copy and configure environment
cp .env.example .env
# Edit .env and add your ANTHROPIC_API_KEY

# 2. Start the full stack
make up

# 3. Verify everything is running
make ps
make smoke-test
```

## Services

| Service | Port | Description |
|---------|------|-------------|
| backend | 8080 | Go API server |
| frontend | 3001 | Next.js dashboard |
| postgres | 5432 | PostgreSQL database |
| redis | 6379 | Job queue (Asynq) |
| steel-browser | 3010 | Headless browser for Playwright (container port 3000) |
| langgraph-sidecar | 8000 | Python multi-agent orchestrator |

## Commands

```bash
make up           # Start all services
make down         # Stop all services
make logs         # Tail logs from all services
make ps           # Show service status
make rebuild      # Rebuild and restart all services
make smoke-test   # Run end-to-end smoke test
```

## Architecture (Docker Network)

```
frontend:3001 → (browser) → backend:8080
backend:8080 → postgres:5432
backend:8080 → redis:6379
backend:8080 → steel-browser:3000
backend:8080 → langgraph-sidecar:8000
langgraph-sidecar:8000 → backend:8080
```

All services communicate via the default Docker Compose network using service names as hostnames.

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `ANTHROPIC_API_KEY` | Yes (default provider) | Anthropic API key — used when `LLM_PROVIDER` is unset or `anthropic` |
| `LLM_PROVIDER` | No | `anthropic` (default), `openai`, `google`, `deepseek`, `mistral`, `groq`, `openrouter`, `custom`, `local`, `ollama` |
| `LLM_API_KEY` | No | API key for non-Anthropic providers (falls back to `ANTHROPIC_API_KEY`/`OPENAI_API_KEY` per provider) |
| `LLM_BASE_URL` | No | Override endpoint; hosted providers get correct per-provider defaults automatically |
| `LLM_MODEL` | No | Model name (default: claude-sonnet-4-5) |
| `QUEUE_ENABLED` | No | `true` routes run execution through the Redis/Asynq durable queue (default: in-process goroutines) |
| `CORS_ALLOWED_ORIGINS` | No | Comma-separated origin allowlist for production (default: wildcard, development only) |
| `STEEL_API_KEY` | No | Steel Browser auth key |
| `API_KEY` | No | API authentication key (required in production when APP_ENV ≠ development) |
| `APP_ENV` | No | Environment: "development" (default) or "production". Production requires API_KEY. |
| `JWT_SECRET` | No | Dashboard cookie-auth secret; random per-process secret generated if unset |
| `GITHUB_WEBHOOK_SECRET` | No | Shared secret for GitHub webhook HMAC verification (falls back to `API_KEY` if unset) |
| `GOTEST_AI_PLANNING` | No | `1` enables AI feature extraction and draft-plan generation |
| `GOTEST_APPROVED_CASE_RUNNER` | No | `docker` enables real Docker Playwright execution for approved cases |
| `MAX_FIX_ATTEMPTS` / `DEFAULT_TIMEOUT_SECONDS` / `MAX_CONCURRENT_RUNS` | No | Execution limits (defaults: 3 / 300 / 10) |

## Health Checks

All services have Docker health checks configured:
- **backend**: `GET /health` every 5s
- **frontend**: HTTP check on port 3001 every 10s
- **postgres**: `pg_isready` every 5s
- **redis**: `redis-cli ping` every 5s
- **steel-browser**: HTTP check on port 3000 every 10s
- **langgraph-sidecar**: `GET /health` every 10s

## Troubleshooting

### Steel Browser crashes
Ensure `/dev/shm` is mounted and `shm_size: 2gb` is set. Chrome requires shared memory.

### Backend can't connect to PostgreSQL
The backend waits for PostgreSQL health check before starting. Check `docker compose logs postgres`.

### Frontend shows empty data
The frontend calls the backend at `http://localhost:8080` from the browser. Ensure port 8080 is exposed and accessible.

### LangGraph sidecar fails to start
Check that Python dependencies install correctly: `docker compose logs langgraph-sidecar`.

## Data Persistence

- PostgreSQL data: `pgdata` Docker volume
- Screenshots/reports: `app_data` Docker volume

To reset all data:
```bash
make down
docker volume rm agent_test_pgdata agent_test_app_data
```
