# Repository Discovery

**Owner:** Engineering  
**Authoritative sources:** Tracked source, manifests, migrations, lockfiles, and Git metadata  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-26  
**Verification performed:** Static repository inspection  
**Confidence:** High for repository topology and static evidence; runtime behavior and business intent are partly UNKNOWN

## Executive Summary

**Verified:** GoTest Agent is a self-hosted testing application with a Go API/MCP layer, Next.js dashboard, PostgreSQL schema, browser execution implementations, Redis/Asynq code, and an optional Python LangGraph sidecar (`README.md:13-30`; `go.mod:1-15`; `frontend/package.json:5-29`; `docker-compose.yml:1-136`).

**Verified:** The active dashboard web-run path, reusable Agent/MCP path, approved-case path, Asynq worker, and sidecar graph implement different orchestration semantics (`internal/api/server.go:1203-1286,1858-1953`; `internal/agent/agent.go:223-345`; `internal/queue/worker.go:27-92`; `sidecar/graph.py:18-42`).

**Verified:** The repository has an unresolved structural gap: the server imports `internal/planning`, but no tracked package exists (`internal/api/server.go:25,44,63,67`). **Inferred:** this blocks ordinary compilation. **UNKNOWN:** actual compiler output.

## Business Understanding

### Problem and users

- **Verified:** Product documentation describes an AI-assisted testing agent that analyzes a project, generates test plans and Playwright tests, executes them, attempts repairs, and reports evidence (`README.md:13-20`; `frontend/src/lib/docs.ts:23-46`).
- **Verified:** Explicit personas are developers, QA teams, technical leads, and teams wanting self-hosted control (`frontend/src/lib/docs.ts:35-46`).
- **Inferred:** The connected project, planning, reusable-case, list, schedule, monitoring, review, and release surfaces indicate a product direction toward a governed continuous-QA lifecycle (`internal/api/server.go:135-234`; `frontend/src/components/sidebar.tsx:11-24`).
- **UNKNOWN:** Business owner, pricing strategy, adoption metrics, target scale, service levels, supported-browser policy, and formal production-readiness criteria.

### Main workflows

1. **Verified:** Guided one-off run from requirements/target through run detail (`frontend/src/app/create/page.tsx:63-85,104-291`; `internal/api/server.go:1543-1953`).
2. **Verified:** Project intake → feature extraction → draft test plan → approved reusable cases (`frontend/src/app/projects/page.tsx:71-155,322-419`; `internal/api/server.go:273-598`).
3. **Verified:** Case execution, refinement, proposal review, and case versioning (`frontend/src/app/tests/page.tsx:55-95`; `internal/api/server.go:788-935`).
4. **Verified:** Test lists, schedules, monitoring, and run history (`frontend/src/app/suites/page.tsx:55-205`; `internal/api/server.go:1045-1200,2529-2741`).
5. **Verified:** Run audit, events, recordings, visuals, comparison, report, and failure analysis (`frontend/src/app/runs/[id]/page.tsx:38-165`; `internal/api/server.go:2197-2505`).
6. **Verified:** MCP/IDE workflow through four stdio tools (`internal/mcp/server.go:33-143`).

## Technology Stack

| Area | Verified technology | Evidence |
|---|---|---|
| Backend | Go 1.26.4, Go Modules, Chi | `go.mod:1-14`; `internal/api/server.go:13-14` |
| Frontend | TypeScript, React 19.2.7, Next.js 16.2.7 App Router, Tailwind 4, npm | `frontend/package.json:5-29`; `frontend/postcss.config.mjs:1-7` |
| Sidecar | Python 3.12, FastAPI, LangGraph, Uvicorn | `sidecar/Dockerfile:1-7`; `sidecar/requirements.txt:1-6` |
| Database | PostgreSQL 16.14, pgx/v5 | `docker-compose.yml:57-74`; `go.mod:10` |
| Queue | Redis 7 + Asynq code; primary server wiring absent | `docker-compose.yml:76-87`; `internal/queue/worker.go:27-92`; `cmd/server/main.go:15-49` |
| Browser execution | Direct Playwright-Go, Docker runner, Steel runner/client | `internal/agent/playwright_runner.go`; `internal/runner/docker.go`; `internal/runner/steel.go` |
| Build | Make, Go toolchain, npm, Docker/Compose | `Makefile:3-48`; Dockerfiles |
| Deployment | Local/demo Docker Compose is the only tracked full-stack deployment | `docker-compose.yml:1-136`; `README.md:182-206` |

Details belong in [`DEPENDENCIES.md`](DEPENDENCIES.md).

## Repository Structure

| Path | Verified responsibility |
|---|---|
| `cmd/server/` | HTTP server entry point, PostgreSQL/memory selection, migrations, scheduler, listener |
| `cmd/mcp/` | Independent stdio MCP process |
| `internal/api/` | Router, middleware, composition, handlers, web orchestration |
| `internal/agent/` | Run model, state machines, LLM interfaces/clients, direct Playwright, sidecar client |
| `internal/db/` | Run/settings persistence and migration runner |
| `internal/project/` | Project memory/PostgreSQL repositories |
| `internal/schedule/` | Schedule memory/PostgreSQL repositories and next-run calculation |
| `internal/events/`, `recordings/`, `visual/` | Process-local execution evidence stores |
| `internal/release/`, `notify/`, `workflow/` | Process-local release, notification, review, and suite stores |
| `internal/metrics/`, `intelligence/`, `compare/`, `gitdiff/` | Stateless analytical/domain calculations |
| `internal/queue/` | Asynq task producer/worker code; unwired from executable entry points |
| `internal/runner/`, `steel/`, `vision/`, `evals/`, `webhook/` | Execution and external integrations |
| `frontend/src/app/` | Next.js routes/pages |
| `frontend/src/lib/` | Typed API client, product docs, helpers |
| `frontend/src/components/` | Shared UI and run-console components |
| `sidecar/` | FastAPI/LangGraph optional multi-agent service |
| `docs/` | Docker operator guide |
| `.ai/` | Tracked internal engineering knowledge base |

See [`CODEMAP.md`](CODEMAP.md) for directory- and symbol-level navigation.

## Architecture Summary

- **Verified:** One Go HTTP process composes most domain stores and performs web orchestration. This supports a **modular-monolith** classification (`cmd/server/main.go:15-49`; `internal/api/server.go:38-89`).
- **Verified:** Compose adds separate frontend, PostgreSQL, Redis, Steel, and sidecar processes, but Redis/Steel/sidecar are not part of the primary web-run composition (`docker-compose.yml:1-136`; `cmd/server/main.go:15-49`; `internal/api/server.go:1858-1953`).
- **Verified:** Runtime events are transient in-process notifications, not event-sourced state (`internal/events/store.go:42-145`; `internal/db/store.go:46-249`).
- **Verified:** Repository, LLM, and runner interfaces create partial ports-and-adapters boundaries, but the web path directly constructs concrete LLM/runner clients (`internal/agent/agent.go:110-145`; `internal/api/server.go:1889-1929`).
- **UNKNOWN:** The canonical orchestrator, browser backend, queue strategy, and multi-instance model.

See [`ARCHITECTURE.md`](ARCHITECTURE.md).

## Evidence Collected

### Primary implementation evidence

- Route registry and handlers: `internal/api/server.go:124-3556`
- Web-run orchestration: `internal/api/server.go:1543-1953`
- Reusable Agent: `internal/agent/agent.go:110-345`
- Database migrations: `internal/db/migrations/001_init.sql`–`008_change_proposals.sql`
- Runtime composition: `cmd/server/main.go:15-67`; `docker-compose.yml:1-136`
- Frontend contracts and workflows: `frontend/src/lib/api.ts:3-691`; `frontend/src/app/**/page.tsx`
- Sidecar endpoints/graph: `sidecar/main.py:8-69`; `sidecar/graph.py:18-42`
- Dependency manifests: `go.mod`, `go.sum`, `frontend/package*.json`, `sidecar/requirements.txt`

### Documentation evidence and conflicts

- **Verified:** Root README is product/operator-facing but points to an absent ignored `planning/` directory (`README.md:32`; `.gitignore:72-73`).
- **Verified:** `docs/docker.md` predates later configuration changes and lists Steel host port as 3000, while Compose publishes `3010:3000` (`docs/docker.md:26-33`; `docker-compose.yml:92-93`).
- **Verified:** Embedded product docs claim Steel execution, Slack/Telegram notification, and PHPUnit generation; active code does not establish those as primary web capabilities (`frontend/src/lib/docs.ts:31,396,615`; `internal/api/server.go:1858-1953`; `internal/notify/store.go:58-99`; `internal/agent/llm_anthropic.go:60-97`).
- **Verified:** `frontend/README.md` remains generic create-next-app boilerplate (`frontend/README.md:1-36`).

## Confidence Summary

| Area | Confidence |
|---|---|
| Tracked paths, route registrations, schema, manifests | High |
| Static wiring and absence findings | High |
| Runtime behavior derived from direct code paths | Medium until exercised |
| Build/test status | UNKNOWN |
| Dependency vulnerability status | UNKNOWN |
| Production topology and external controls | UNKNOWN |
| Business intent beyond explicit docs | Low/UNKNOWN |

## UNKNOWN Questions

- `UNK-001` — Which execution path and runner should be canonical?
- `UNK-002` — Is the product single-admin or multi-user/tenant?
- `UNK-003` — Is horizontal backend scaling required?
- `UNK-004` — Which operational entities must survive restarts?
- `UNK-005` — What deployment controls exist outside this repository?
- `UNK-006` — What are the current build, test, and CVE results?
- `UNK-007` — Are stored credential fields references or live secrets?

Current status and priorities: [`PROJECT_STATE.md`](PROJECT_STATE.md). Actionable work: [`TODO.md`](TODO.md).
