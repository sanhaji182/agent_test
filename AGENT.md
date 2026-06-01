# AGENT.md — GoTest Agent Execution Guide

## Project Status: Production-Ready MVP ✅

All core phases are complete. The system is running as a full Docker stack.

## What Is Built

### Backend (Go)
- HTTP API (Chi router) on `:8080`
- MCP server (stdio) for Cursor/VS Code
- Anthropic Claude-based code analysis, test plan generation, script writing, auto-fix loop
- Steel Browser integration for Playwright execution
- PostgreSQL persistence + in-memory fallback
- Redis job queue + background worker
- SSE live event stream (global `/api/v1/stream` + per-run `/api/v1/runs/:id/stream`)
- Event bus for instant push (no polling lag)
- Browser video recording & replay with step→timestamp mapping
- Visual regression (GPT-4o Vision)
- Braintrust eval logging
- Git diff-based impacted test selection
- Background scheduler (daily/weekly/monthly/cron)
- Webhook notifications + alert rules
- Review workflow (approve/reject/request-changes)
- Risk scoring, suite selection, release confidence
- HTML report generation
- Auth middleware (API key)
- Production hardening: rate limiting, input validation, security headers

### Frontend (Next.js)
- Real-time control room dashboard with SSE live updates
- Failure toast notifications
- Progress bar with stage hints (early/midway/near completion) for running tests
- Run detail audit view: video replay, step breakdown, screenshots, failure summary
- Step-to-video timestamp sync
- Risk dashboard, review queue, suite browser
- Monitoring (schedules), alerts, releases, exports pages
- Bilingual documentation (Indonesian + English)
- Filter + search on run list (by ID, state, requirements, project path)
- Onboarding flow with demo seed

### LangGraph Sidecar (Python)
- FastAPI server on `:8000`
- Multi-agent graph: planner → writer → executor → critic → fixer

## Architecture

```
Frontend (:3001) → Go Backend (:8080) → Steel Browser (:3010)
                         ↓
              PostgreSQL / Redis / LangGraph Sidecar (:8000)
```

## Running the Stack

```bash
make up          # start all 6 services
make rebuild     # rebuild images and restart
make down        # stop
make logs        # tail logs
make smoke-test  # end-to-end verification
```

## Key Files

| Path | Purpose |
|------|---------|
| `internal/agent/agent.go` | Core agent pipeline |
| `internal/api/server.go` | All HTTP routes |
| `internal/runner/docker.go` | Playwright execution |
| `internal/events/store.go` | Event bus + SSE |
| `frontend/src/app/page.tsx` | Overview / control room |
| `frontend/src/app/runs/page.tsx` | Run list with filter/search |
| `frontend/src/app/runs/[id]/page.tsx` | Run detail / audit view |
| `sidecar/graph.py` | LangGraph multi-agent graph |

## What Remains / Next Priorities

1. **Analytics** — pass rate trends, failure hotspots, flaky detection (backend metrics exist, frontend charts partial)
2. **Alert rules UI** — backend exists, frontend page is placeholder
3. **Suite management** — tags and pinning UI
4. **Export** — JSON download buttons for runs, comparisons, risk reports

## Operating Rules

1. Read this file and the relevant PRD before changing code.
2. Do not break existing behavior — run `go build ./...` and `npm run build` after changes.
3. Prefer minimal, focused changes over large rewrites.
4. Keep polling as fallback whenever touching SSE/live transport.
5. Do not fake precision (timestamps, percentages) — use real data or approximate labels.
6. After each meaningful change: commit, push, then `make rebuild`.
