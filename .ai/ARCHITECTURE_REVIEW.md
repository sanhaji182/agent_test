# Architecture Review

**Owner:** Engineering  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-27  
**Verification performed:** Static inspection of all Go packages, Compose topology, frontend routing, and sidecar graph  
**Confidence:** High for observed patterns and violations; canonical intent partly UNKNOWN

## Architecture Style Assessment

**Primary: Modular monolith.** One Go process composes domain packages and serves REST, SSE, scheduler, and orchestration from a single runtime (`cmd/server/main.go:15-49`; `internal/api/server.go:38-89`).

**Supporting:**
- Client/server — browser calls REST/SSE directly; no Next.js BFF layer.
- In-process event-driven projection — transient pub/sub for SSE.
- Partial ports-and-adapters — interfaces at persistence and external-call boundaries, but the primary web path bypasses the reusable Agent composition.
- Distributed deployment shell — Compose runs six services, but Redis/Steel/sidecar are not wired into the active execution path.

## Layer Violations

### L1 — Transport/application/persistence mixing in `Server`

**Severity:** HIGH  
**Evidence:** `internal/api/server.go:38-89` makes `Server` the composition root, router, handler registry, and direct delegator to every store. Handlers at `server.go:1543-1589,1858-1953` directly call stores, LLM clients, runners, and event emitters in a single scope.  
**Impact:** Every feature change touches `server.go`. Testing individual handlers requires the composed `Server`. Response format cannot change without touching every handler.  
**Recommendation:** Extract route modules and application-service structs. Handlers own only transport decode/response encode. Services own business logic and transaction boundaries.

### L2 — Missing service layer between transport and persistence

**Severity:** HIGH  
**Evidence:** Approval flow (`server.go:566-597`) performs case creation, draft update, and response encoding in one handler. Proposal approval (`server.go:850-899`) updates test case then proposal in separate store operations. Monitoring summary (`server.go:2064-2147`) scans runs in memory for analytical display.  
**Impact:** Business rules (transaction boundaries, consistency, validation) are scattered across handlers.  
**Recommendation:** Create domain services per aggregate: `ProjectService`, `PlanningService`, `ExecutionService`, `ScheduleService`, `ReleaseService`.

### L3 — Orchestration divergence

**Severity:** HIGH  
**Evidence:** Five separate execution implementations with different state machines:
1. `Server.launchRun` → `Agent.Launch` (web) — `idle→analyzing→plan_generated→writing_tests→running↔fixing→done|failed`
2. `Agent.Execute` (MCP) — full state machine with fix loop
3. `executeApprovedTestCaseRun` — simulated or Docker path
4. Asynq worker — implemented but unwired
5. LangGraph sidecar — separate graph, executor doesn't await results

**Impact:** Bug fixes don't propagate. Users get different behavior from different entry points.  
**Recommendation:** Route all entry points through one orchestration boundary. Retire or explicitly mark experimental paths.

## Cyclic Dependencies

None found. Package graph is acyclic: `cmd` → `internal/api` → `internal/{agent,db,project,schedule,...}` with no reverse imports.

## God Objects / God Services

### `api.Server`

**Severity:** HIGH  
**Evidence:** 14 fields mixing transport (`router`), configuration (`cfg`), persistence (`store`, `settings`, `projects`, `planning`, `schedules`), and runtime state (`events`, `recordings`, `visuals`, `releases`, `notifs`, `reviews`, `suites`). 85+ handler methods. Route registration, middleware, and compositon in one file.  
**Recommendation:** Split by domain; each domain service owns its stores and business logic.

## Large Files

| File | Lines | Severity | Recommendation |
|---|---|---|---|
| `internal/api/server.go` | ~3,556 | HIGH | Split into route modules + service packages. Target: <500 lines per file. |

## Duplicate Logic

| Pattern | Locations | Severity | Recommendation |
|---|---|---|---|
| Execution orchestration | `server.go:1858-1953`, `agent.go:223-345`, `server.go:1203-1286`, `queue/worker.go` | HIGH | Consolidate behind one `Executor`/`Orchestrator` interface |
| LLM provider routing | `server.go:1889-1894` vs `ai/client.go:57-70` vs `server.go:3530-3535` | HIGH | One `LLMFactory` with consistent routing |
| Draft plan creation | `server.go:389-400` and `474-485` | LOW | Extract `createDraft` helper |
| Idle schedule creation | `server.go:2671-2691` and `2721-2739` | MEDIUM | Route through one `createScheduledRun` function |

## Poor Package Organization

| Issue | Evidence | Recommendation |
|---|---|---|
| `internal/api` is a single package with a single file | `server.go` contains 100% of API code | Split: `internal/api/routes.go`, `internal/api/handlers_runs.go`, `internal/api/handlers_projects.go`, etc. |
| Planning package was absent | Reconciled 2026-07-26 | Now at `internal/planning/{types,memory,db}.go`; follow same pattern for future domain packages |
| `internal/ai` duplicates LLM logic already in `internal/agent` | `ai/client.go` vs `agent/llm_*.go` | Consolidate under `internal/agent` or a shared `internal/llm` |

## SRP Violations

| Type | Responsibilities | Evidence |
|---|---|---|
| `api.Server` | Router, middleware, composition root, 85+ handlers, business logic, orchestration, provider settings, demo seeding, SSE | `server.go:38-3556` |
| `PlaywrightRunner.Run` | Browser install, launch, action execution, self-healing with LLM, video capture, result assembly | `playwright_runner.go:32-186` |

## Tight Coupling

| Coupling | Evidence | Recommendation |
|---|---|---|
| Handlers → concrete LLM/Runner construction | `server.go:1889-1929` constructs `NewAnthropicLLM`/`NewPlaywrightRunner` inline | Inject via constructor or factory |
| Handlers → concrete store types | `server.go:62-69` type-asserts to `*db.Store` to select PostgreSQL | Use capability interface or explicit store config |
| Webhook → API key reuse | `server.go:237` passes `cfg.APIKey` as webhook secret | Separate config field |

## Missing Abstractions

- No canonical `Executor` or `Orchestrator` — `Agent.Execute` is the closest candidate but is unused by the web path.
- No `LLMFactory` — two independent factory paths exist.
- No `ArtifactStore` — videos, screenshots, and recordings use different paths with no shared lifecycle.
- No `JobQueue` — goroutines are the only async mechanism despite Asynq being available.

## Unnecessary Abstractions

None identified. The existing repository interfaces (RunStore, project Store, schedule Repository, Store, LLM, Runner) are all consumed and appropriate.

## High Cognitive Complexity

| Location | Issue | Recommendation |
|---|---|---|
| `executeRealRun` | LLM loading, plan generation, script generation, runner construction, and result handling in one function | Extract: `loadLLMConfig`, `executeWithLLM`, `runWithPlaywright` |
| Playwright self-healing loop | Nested try/heal/retry with DOM capture, screenshot, LLM call, action retry | Extract `healAction` method; flatten retry logic |
| `buildPlaywrightSpec` + Docker config | Playwright spec uses `json.Marshal` for title/URL; Docker config uses string concatenation | Unify escaping; both should use `json.Marshal` |

## Summary

| Category | Count | Highest Severity |
|---|---|---|
| Layer violations | 3 | HIGH |
| Cyclic dependencies | 0 | — |
| God objects | 1 | HIGH |
| Large files | 1 | HIGH |
| Duplicate logic | 4 | HIGH |
| SRP violations | 2 | HIGH |
| Tight coupling | 3 | HIGH |
| Missing abstractions | 4 | HIGH |
| Unnecessary abstractions | 0 | — |
| High cognitive complexity | 3 | MEDIUM |

The architecture is fundamentally sound in its package-per-domain organization and interface boundaries at external edges. The primary repair needed is consolidating execution paths and separating the monolithic `api.Server` into domain services, neither of which requires redesign — just extraction and wiring of already-existing code.
