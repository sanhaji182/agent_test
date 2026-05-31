# GoTest Agent — Docker Deployment

## Prerequisites

- Docker 24+ with Compose V2
- 4GB+ RAM (8GB recommended for Steel Browser)
- Ports available: 8080, 3001, 5432, 6379, 3000, 8000

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
| steel-browser | 3000 | Headless browser for Playwright |
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
| `ANTHROPIC_API_KEY` | Yes (for LLM) | Anthropic API key |
| `LLM_MODEL` | No | Model name (default: claude-sonnet-4-5) |
| `STEEL_API_KEY` | No | Steel Browser auth key |
| `API_KEY` | No | Backend API key (empty = no auth) |

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
