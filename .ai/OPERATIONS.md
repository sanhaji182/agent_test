# Operations Runbook

**Owner:** Engineering (SRE ownership UNKNOWN)
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`
**Last updated:** 2026-07-27
**Scope:** Day-to-day operational procedures for the Compose-deployed application

## Service Overview

| Service | Port (host) | Port (container) | Health check | Data |
|---|---|---|---|---|
| Backend (Go API) | 8080 | 8080 | `GET /health` | PostgreSQL + /tmp/agent_test/videos |
| Frontend (Next.js) | 3001 | 3001 | HTTP 200 | None (static; API-dependent) |
| PostgreSQL | — (internal) | 5432 | `pg_isready` | `pgdata` volume |
| Redis | — (internal) | 6379 | `redis-cli ping` | None (experimental) |
| Steel Browser | 3010 | 3000 | HTTP | None |
| Sidecar | 8000 | 8000 | `GET /health` | In-memory jobs |

## Startup

```bash
cp .env.example .env
# Edit .env: set ANTHROPIC_API_KEY (or LLM_API_KEY + LLM_PROVIDER)
# For production: set APP_ENV=production and API_KEY=<secure-value>
make up
make smoke-test
```

Access dashboard at `http://localhost:3001` and API at `http://localhost:8080`.

## Shutdown

```bash
make down        # Stop all services; preserve volumes
make down -v     # Stop and destroy volumes (all data lost)
```

**Warning:** `docker compose down -v` permanently destroys all PostgreSQL data, screenshots, and reports. Only do this for a clean reset.

## Health Verification

```bash
# All services
docker compose ps

# Individual health
curl http://localhost:8080/health
curl http://localhost:3001
curl http://localhost:8000/health
```

Note: `/health` returns `{"status":"ok"}` for process liveness only — it does not check PostgreSQL, Redis, or other downstream dependencies.

## Viewing Logs

```bash
make logs                  # All services
docker compose logs app    # Backend only
docker compose logs -f     # Follow all
```

## Common Issues

### Backend can't connect to PostgreSQL

**Symptom:** `slog.Warn("db not available, using in-memory store")` in logs.

**Check:** `docker compose ps postgres` — is it healthy?

**Fix:**
```bash
docker compose restart postgres
docker compose restart app
```

### Frontend shows empty data

**Symptom:** Dashboard loads but shows empty states for runs, projects, etc.

**Causes:**
1. Backend is not reachable at the configured `NEXT_PUBLIC_API_URL`.
2. Backend has fallen back to in-memory storage (PostgreSQL not available).
3. `API_KEY` is configured but the frontend doesn't send it.

**Check:** `curl http://localhost:8080/api/v1/runs` (with `-H "X-Api-Key: <key>"` if API_KEY is set).

### Steel Browser crashes

**Symptom:** Steel logs show crashes, or browser-dependent features fail.

**Notes:**
- Steel has `SYS_ADMIN` capability and 2 GB shared memory.
- Steel uses the mutable `:latest` tag — versions may change between restarts.
- The primary web run path uses local Playwright, not Steel. Steel is only relevant if explicitly configured.

### Sidecar won't start or fails

**Symptom:** LangGraph sidecar logs errors.

**Check:** `docker compose logs sidecar`

**Note:** The sidecar is optional — the primary web path does not require it.

### Runs are stuck in "idle"

**Symptom:** Created runs stay idle and never transition.

**Causes:**
1. No LLM API key configured in settings (resolution chain: DB settings → env vars → defaults).
2. Settings store unavailable (PostgreSQL not connected).
3. Non-list scheduled runs — these create `idle` rows without launching execution (known bug, `TODO-008`).

**Fix:**
1. Go to Settings page and configure LLM provider/model/key.
2. Ensure PostgreSQL is healthy and the backend connected to it.
3. For manual test-list schedules, use the list-creation workflow instead.

## Data Persistence

### What survives restart

- PostgreSQL data (`pgdata` volume): runs, projects, settings, schedules, planning entities
- Application files (`app_data` volume): persistent screenshots and reports

### What is lost on restart

- All events, recordings metadata, visual artifacts, releases, notifications, reviews, suites
- Sidecar job status
- Temporary Playwright videos under `/tmp/agent_test/videos`

### Destructive Reset

```bash
docker compose down -v
docker compose up -d
```

All data is permanently lost. Re-populate via the Demo Seed button on the dashboard (`POST /api/v1/demo/seed`).

## Backup

No automated backup exists at this revision. Manual backup:

```bash
# PostgreSQL
docker compose exec postgres pg_dump -U postgres gotest_agent > backup-$(date +%Y%m%d).sql

# Application files
docker compose cp app:/data ./data-backup-$(date +%Y%m%d)
```

Restore: see `.ai/BACKUP_AND_RECOVERY.md`.

## Environment Variables at a Glance

| Variable | Required? | Default | Notes |
|---|---|---|---|
| `APP_ENV` | No | `development` | Set to `production` to require `API_KEY` |
| `APP_PORT` | No | `8080` | |
| `API_KEY` | Production only | empty | Empty = no auth; required when `APP_ENV=production` |
| `DATABASE_URL` | No | local DSN | Empty = in-memory only |
| `ANTHROPIC_API_KEY` | For Anthropic | empty | Web runs read LLM key from DB settings, not this env |
| `LLM_PROVIDER` | No | (from DB settings) | `anthropic`, `openai`, `custom`, `local` |
| `LLM_MODEL` | No | `claude-sonnet-4-5` | |
| `GITHUB_WEBHOOK_SECRET` | No | Falls back to `API_KEY` | Separate webhook HMAC secret |
| `GOTEST_AI_PLANNING` | No | disabled | Set to `1` or `true` to enable AI feature extraction |
| `GOTEST_APPROVED_CASE_RUNNER` | No | simulated | Set to `docker` for real Playwright execution |

Full inventory: `.env.example`. Documentation gap analysis: `.ai/DOCUMENTATION_GAP.md`.

## Resource Monitoring

No production monitoring exists at this revision. Watch:

```bash
docker compose stats          # Per-container CPU/memory
make logs | grep ERROR        # Error log extraction
curl http://localhost:8080/health  # Liveness
```

Event store growth, goroutine count, and temporary video accumulation require manual inspection.

## Emergency Procedures

### Database corruption or accidental deletion

1. Stop the application: `make down`
2. If you have a backup: restore per `.ai/BACKUP_AND_RECOVERY.md`
3. If you do not have a backup: data recovery is not supported at this revision
4. Restart: `make up`

### Application is unresponsive

1. Check all services: `docker compose ps`
2. Check backend logs: `docker compose logs app | tail -100`
3. If memory-exhausted: check for heavy goroutine count or event store growth
4. Restart backend: `docker compose restart app`
5. Note: restart loses all memory-backed state (events, recordings, notifications)

### Suspicious activity

If the backend is exposed without authentication (`API_KEY` empty or `APP_ENV=development`):

1. All `/api/v1` endpoints are publicly accessible.
2. Check for unauthorized runs, settings changes, or data tampering.
3. Set `APP_ENV=production` and `API_KEY=<strong-value>` immediately.
4. Review `.ai/SECURITY.md` for all known default-configuration risks.
