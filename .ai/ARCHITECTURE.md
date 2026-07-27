# Architecture

**Owner:** Engineering  
**Authoritative sources:** `cmd/server/main.go`, `cmd/mcp/main.go`, `internal/api/server.go`, `internal/agent/agent.go`, `internal/events/store.go`, `sidecar/`, `docker-compose.yml`  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-26  
**Verification performed:** Static inspection  
**Confidence:** High for tracked boundaries and wiring; canonical intent and runtime behavior are partly UNKNOWN

## Architecture Style

### Primary classification: modular monolith

**Verified:** The backend is one Go HTTP process whose `api.Server` composes run/settings/project/planning/event/recording/visual/schedule/release/notification/review/suite dependencies (`cmd/server/main.go:15-49`; `internal/api/server.go:38-89`). Routes, handlers, application orchestration, and much of dependency construction share `internal/api/server.go` (`internal/api/server.go:124-261,1858-1953`).

**Inferred:** “Modular monolith” is the best description because packages are grouped by domain, but deployed backend ownership is concentrated in one executable/process.

### Supporting classifications

| Classification | Assessment | Evidence |
|---|---|---|
| Client/server | **Verified:** Browser calls Go REST/SSE directly; no observed Next.js BFF | `frontend/src/lib/api.ts:1,222-234,386-415` |
| In-process event-driven projection | **Verified:** Execution emits transient events to memory-backed subscribers | `internal/events/store.go:42-145`; `internal/api/server.go:2197-2262,3168-3215` |
| Partial ports-and-adapters | **Verified:** LLM, Runner, RunStore, project Store, and schedule Repository interfaces exist; web path bypasses some of them | `internal/agent/agent.go:110-145`; `internal/db/memory.go:10-17`; `internal/api/server.go:1889-1929` |
| Distributed deployment shell | **Verified:** Compose has six services, but primary web execution does not use Redis/Steel/sidecar | `docker-compose.yml:1-136`; `cmd/server/main.go:15-49`; `internal/api/server.go:1858-1953` |
| Microservices | Not supported as primary style; optional sidecar is one separately deployed helper | `sidecar/main.py:8-69` |
| CQRS/event sourcing | Not supported; commands/reads use the same stores and events cannot reconstruct durable state | `internal/db/store.go:46-249`; `internal/events/store.go:42-100` |

## Runtime Processes

| Process/service | Verified responsibility | Active in primary web flow? |
|---|---|---|
| Next.js frontend | Serves dashboard on 3001; browser performs REST/SSE calls | Yes |
| Go server | API, SSE, scheduler, orchestration, direct Playwright execution | Yes |
| PostgreSQL | Core durable store when connection succeeds | Yes, optional with memory fallback |
| Redis | Compose service; Asynq code target | No primary server wiring found |
| Steel Browser | Browser service/client/runner implementation | No primary server wiring found |
| Python sidecar | FastAPI + LangGraph job execution | No primary server construction found |
| MCP binary | Separate stdio process with independent Agent and in-memory run map | Separate entry point |

*Evidence:* `docker-compose.yml:1-136`; `cmd/server/main.go:15-67`; `cmd/mcp/main.go:13-25`; `sidecar/main.py:8-69`.

## Layers and Boundaries

### Transport

- **Verified:** Chi REST/SSE and middleware live in `internal/api` (`internal/api/server.go:55-60,124-261`).
- **Verified:** MCP stdio transport lives in `internal/mcp` (`internal/mcp/server.go:33-143`).
- **Verified:** Sidecar HTTP transport lives in `sidecar/main.py:15-69`.

### Application/orchestration

- **Verified:** Primary web orchestration delegates to `agent.Agent.Launch` via `Server.launchRun` (ADR-001 resolved).
- **Verified:** Execution owned by `agent.Agent.Execute` with `RunPersistence` auto-saving at each transition (`internal/agent/agent.go`).
- **Verified:** Approved-case orchestration is separate (`internal/api/server.go:1203-1286`).
- **UNKNOWN:** Which is intended to be canonical. See [`ADR-001`](DECISIONS.md#adr-001--choose-the-canonical-execution-orchestrator).

### Domain calculations

**Verified:** Comparison, metrics, intelligence, Git impact, visual comparison, and reporting are package-level domain/stateless services (`internal/compare`; `internal/metrics`; `internal/intelligence`; `internal/gitdiff`; `internal/visual`; `internal/report`).

### Persistence

- **Verified:** Repository boundaries exist for runs, projects, and schedules (`internal/db/memory.go:10-17`; `internal/project/store.go:33-38`; `internal/schedule/store.go:47-54`).
- **Verified:** Runs/settings/projects/schedules have PostgreSQL-backed paths. Planning is intended to have one, but its package is absent (`internal/api/server.go:62-69`).
- **Verified:** Events, recordings, visuals, releases, notifications, reviews, and suites are always constructed as memory stores (`internal/api/server.go:71-85`).

Physical persistence detail belongs in [`DATABASE.md`](DATABASE.md).

## Component Interactions and Data Flow

### Primary web run

1. **Verified:** Browser submits `POST /api/v1/runs`; on `202` it navigates to run detail (`frontend/src/app/create/page.tsx:63-85`; `frontend/src/lib/api.ts:236-260`).
2. **Verified:** Handler decodes request, derives feature map, persists `idle` run, emits event, starts a goroutine (`internal/api/server.go:1543-1590`).
3. **Verified:** Goroutine uses `context.Background`, loads LLM settings from DB, generates plan/scripts with literal `"Web Application"`, constructs direct Playwright runner, and persists terminal state (`internal/api/server.go:1858-1953`).
4. **Verified:** Run detail fetches run/artifacts and subscribes to per-run SSE only when initial state is active (`frontend/src/app/runs/[id]/page.tsx:38-65`).
5. **Verified:** Per-run SSE replays memory events, subscribes, and polls durable state every two seconds (`internal/api/server.go:2197-2262`).

### Reusable Agent/MCP

**Verified:** MCP constructs its own Anthropic client, Docker runner, Agent, and MCP server. `run_tests` executes synchronously and keeps status in an MCP-local memory map (`cmd/mcp/main.go:13-25`; `internal/mcp/server.go:74-143`). It does not create web-visible durable runs.

### Approved case and list

**Verified:** Approved cases generate Playwright test files. Default execution simulates step completion and pass results; Docker execution is enabled only by `GOTEST_APPROVED_CASE_RUNNER=docker` (`internal/api/server.go:1203-1286`). Lists iterate case IDs and start one run per resolvable case (`internal/api/server.go:1110-1200`).

### Scheduling

**Verified:** One server goroutine polls every 60 seconds (`cmd/server/main.go:41-67`). List-backed schedules start case runs. Non-list schedule paths create idle rows without starting execution (`internal/api/server.go:2637-2741`). Due selection and update are separate (`internal/schedule/store.go:245-269`).

### Sidecar

**Verified:** Reusable Agent can post and poll sidecar jobs (`internal/agent/agent.go:178-220`; `internal/agent/sidecar.go:49-100`). The sidecar graph invokes a backend run endpoint, but its executor treats the asynchronous response as result data rather than waiting for terminal run completion (`sidecar/agents/executor.py:9-38`).

## Concurrency and Events

- **Verified:** Run/rerun/approved-case/webhook paths launch unrestricted goroutines (`internal/api/server.go:253-254,1196-1200,1581-1589,2186-2190`).
- **Verified:** Event store uses one mutex and buffered channels; sends are dropped when buffers are full (`internal/events/store.go:42-91`).
- **Verified:** Event history is unbounded process memory and disappears on restart (`internal/events/store.go:42-100`).
- **Inferred:** Current design assumes one backend instance. **UNKNOWN:** whether this is intended. See `ADR-004`.

## Code-Evidenced Design Principles

- Interface boundaries at external/persistence edges (`internal/agent/agent.go:110-145`; repository interfaces).
- Package-per-domain organization under `internal/`.
- Memory implementations for local development/fallback.
- UUID identifiers and JSON field names using snake_case.
- Parameterized PostgreSQL access and embedded ordered migrations.
- Async run initiation with live event projection.

These are implementation patterns, not immutable owner-approved policy unless recorded in an accepted ADR.

## Architectural Constraints and Risks

1. **Verified:** Missing planning package is an unresolved source boundary (`TODO-001`).
2. **Verified:** The primary API module combines composition, transport, application logic, and integrations (`internal/api/server.go:38-3556`).
3. **Verified:** Execution paths diverge (`ADR-001`, `TODO-007`).
4. **Verified:** Browser backends and artifact paths diverge (`ADR-002`).
5. **Verified:** Durable and transient state lifetimes differ (`ADR-003`).
6. **Verified:** Scheduler and event design lack multi-instance coordination (`ADR-004`).
7. **Verified:** Frontend authentication transport is incompatible with optional API-key enforcement (`ADR-005`).

## Architectural Decision Needs

- [`ADR-001`](DECISIONS.md#adr-001--choose-the-canonical-execution-orchestrator) — orchestrator
- [`ADR-002`](DECISIONS.md#adr-002--choose-the-canonical-browsertest-execution-backend) — runner/backend
- [`ADR-003`](DECISIONS.md#adr-003--define-durable-versus-ephemeral-product-state) — durability
- [`ADR-004`](DECISIONS.md#adr-004--define-deployment-cardinality-and-background-job-ownership) — scaling/scheduler
- [`ADR-005`](DECISIONS.md#adr-005--define-the-authentication-and-authorization-model) — auth/tenancy

Transport details: [`API.md`](API.md). Domain semantics: [`DOMAIN.md`](DOMAIN.md). Change navigation: [`CODEMAP.md`](CODEMAP.md).
