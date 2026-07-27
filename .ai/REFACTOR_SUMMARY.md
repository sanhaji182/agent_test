# Refactor Summary

**Owner:** Engineering  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-27  
**Summary of completed and planned refactoring work**

## Completed

### 1. Reconstructed `internal/planning` package (2026-07-26)

**Before:** `internal/api/server.go:25,44,63,67` imported a package that never existed in any Git commit. The entire module could not compile (`go build` failed with "no required module provides package").

**After:** Three additive files (`types.go`, `memory.go`, `db.go`) following the established `internal/project` Store/MemoryStore/DBStore pattern. No existing code changed. The module compiles.

**Risk:** Low — additive change, no existing behavior modified.  
**Evidence:** `internal/planning/`; `TODO-001`  
**Remaining:** Regression test still pending infrastructure availability.

### 2. Updated root README knowledge-base reference (2026-07-27)

**Before:** `README.md:32` pointed to an absent, gitignored `planning/` directory.

**After:** Points to tracked `.ai/` with document-map reference to `.ai/README.md`.

**Risk:** None — documentation change only.  
**Evidence:** `TODO-019`

## Planned (Sequenced by Phase)

### Phase 1 — Security Baseline

| Refactor | Motivation | Risk | Impact |
|---|---|---|---|
| Fail startup without production auth | Current default: all `/api/v1` routes are public | Medium | Adds `APP_ENV` config; operators must set `API_KEY` in production |
| Separate webhook secret from API key | Reusing one secret across auth and HMAC is a domain violation | Low | New `GITHUB_WEBHOOK_SECRET` config field |
| Authenticate/internalize sidecar | Published unauthenticated service with backend credential forwarding | Medium | Remove host port or add auth; operators must configure sidecar auth |
| Bind LLM credentials to approved origins | Stored credential can be sent to arbitrary caller-chosen host | Low | Settings validation; one provider factory |
| Add HTTP server infrastructure | No timeouts, no body limits, no graceful shutdown since inception | Low | Standard Go `http.Server` configuration |

### Phase 2 — Execution Convergence

| Refactor | Motivation | Risk | Impact |
|---|---|---|---|
| Consolidate 5 execution paths into 1 | Different state machines, retries, and error handling per entry point | High | Touches web, MCP, schedule, webhook, and approved-case flows |
| Eliminate synthetic test success | Mock runner, simulated passes, and false-positive results reported as real | Medium | Approved cases will show real failures; demo seed may need adjustment |
| Fix non-list schedule execution | Schedules create idle runs that never transition | Low | Routes schedule execution through the canonical path |
| Implement run deletion | Endpoint returns 204 without deleting anything | Low | Adds `DeleteRun` to `RunStore`; artifact cleanup |
| Unify provider routing | Test-connection and real execution route providers differently | Medium | One `LLMFactory` used everywhere |

### Phase 3 — Durability

| Refactor | Motivation | Risk | Impact |
|---|---|---|---|
| Persist or document ephemeral stores | 7 domain stores are memory-only; restart silently loses data | Medium | May add new database tables; may add retention policies |
| Atomic schedule claiming | Duplicate runs under concurrent schedulers | Medium | Query change (`FOR UPDATE SKIP LOCKED` or advisory lock) |
| Event retention limits | Unbounded in-memory growth | Low | Configurable cap |

### Phase 4 — Code Organization

| Refactor | Motivation | Risk | Impact |
|---|---|---|---|
| Split `server.go` into route modules and services | 3,556-line file with 85+ handlers, composition, and business logic | Medium | Touches every handler; existing tests must gate the extraction |
| Standardize request/response types | Inconsistent validation; PATCH panics on wrong types; ad-hoc error strings | Medium | Breaking change to error response format |
| Consolidate LLM client layers | Two independent factory paths with divergent provider routing | Low | Remove `internal/ai` or merge into `agent` |
| Remove confirmed dead code | `simulateMockRun`, unused API exports, `class-variance-authority` | Low | Deletion only after confirming no callers |

### Phase 5 — Production Infrastructure

| Refactor | Motivation | Risk | Impact |
|---|---|---|---|
| Production Compose overlay | Dev Compose publishes all services, uses dev credentials, no TLS | Medium | New overlay file; dev Compose unchanged |
| Dependency-aware health checks | `/health` reports OK even when database is unreachable | Low | New readiness endpoint |
| Observability (metrics, structured errors, profiling) | Zero production telemetry exists | Low | Additive instrumentation |
| Graceful shutdown with run draining | In-progress runs abandoned on SIGTERM | Medium | Signal handling + run drain |
| Bounded concurrency and durable queue | Unlimited goroutines per request; no job durability | High | Changes async execution model |
| Container immutability | Mutable tags, unlocked deps, suppressed install failures | Low | Pin digests, lock Python, remove `|| true` |

## Refactoring Principles Applied

1. **Additive first, subtractive second.** Planning package added; no existing code broken.
2. **Tests gate extraction.** Code organization refactors wait for regression test coverage.
3. **Evidence over intuition.** Every refactor cites exact file:line evidence from audit and discovery.
4. **Small, reversible diffs.** Split into phases; each phase produces a compilable, testable state.
5. **No premature abstraction.** Consolidation only when duplicate logic is confirmed; no speculative generalization.
