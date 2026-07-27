# Migration Plan

**Owner:** Engineering  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-27  
**Verification performed:** Static inspection; no deployment migration was performed or tested  
**Confidence:** High for identified safe improvements; product decisions (ADRs) may change sequencing

## Migration Principles

1. **Small, reviewable changes.** Each phase produces a compilable, testable state.
2. **Backward-compatible where feasible.** No breaking API change without documented migration.
3. **Improve evidence with each phase.** Build/tests run after every meaningful change.
4. **Stop on evidence of regression.** If a gate fails, fix before advancing.
5. **Keep the knowledge base synchronized per phase.**

## Phase 0 — Establish a Trustworthy Baseline

**Status:** Partially done (planning package restored; format/test gates pending)

### Steps

1. Finish formatting and lifecycle regression (`TODO-001` remaining gates).
2. Run full `go test ./...` and fix failures (`TODO-014`).
3. Run `govulncheck`, `npm audit`, `pip-audit` (`TODO-015`).
4. Commit the planning package and updated knowledge base with recorded evidence.

### Gates

- `go build ./...` ✅ (passed 2026-07-26)
- `gofmt -d .` produces zero diff ⏳
- `go test -count=1 ./...` all pass ⏳
- Scanner results recorded (not "clean" without evidence) ⏳

## Phase 1 — Safe Defaults (Security Baseline)

**Motivation:** The default Compose configuration has confirmed vulnerabilities that must close before any environment is exposed to untrusted networks.

### Steps

1. Add `APP_ENV` config field; fail startup if `API_KEY` is empty and `APP_ENV != development` (`TODO-002`).
2. Add separate `GITHUB_WEBHOOK_SECRET` config; wire to webhook handler (`TODO-004`).
3. Remove host port from sidecar or add inbound authentication (`TODO-003`).
4. Bind LLM credentials to approved provider origins (`TODO-006`).
5. Add `http.Server` timeouts and body limits (`TODO-018` scope).
6. Remove public `/videos/{filename}` handler; serve all recordings through one authenticated endpoint.

### Gates

- `go test ./...` pass
- Integration test: startup fails without `API_KEY` in production mode
- Integration test: sidecar rejects unauthenticated requests
- Integration test: LLM key cannot be exfiltrated via base URL change

## Phase 2 — Execution Convergence

**Motivation:** Five divergent execution paths with different state machines create bug divergence and confuse operators.

### Prerequisites

- `ADR-001` accepted (canonical orchestrator)
- `ADR-002` accepted (canonical browser backend)

### Steps

1. Extract `ExecutionService` that wraps `agent.Agent` or equivalent (`TODO-007`).
2. Route web create/rerun, MCP, schedules, webhooks, and approved cases through it.
3. Eliminate synthetic success in default paths (`TODO-010`).
4. Fix non-list schedule execution (`TODO-008`).
5. Implement real run deletion (`TODO-009`).
6. Unify provider routing matrix across test-connection and execution paths.
7. Remove `simulateMockRun` dead code (`DC-1`).

### Gates

- All five entry points produce consistent terminal states.
- `run.Mode` and `run.TestType` route to correct execution backend.
- Non-list scheduled run reaches terminal state.
- Run deletion removes row and artifacts.

### Breaking Changes

- Approved cases will no longer synthesize passing assertions by default.
- `simulateMockRun` removed; demo seed may need adjustment.

## Phase 3 — Durability and Persistence

**Motivation:** Multiple stores are memory-only; restart silently loses audit and operational state.

### Prerequisites

- `ADR-003` accepted (durable vs. ephemeral entities)
- `ADR-004` accepted (deployment cardinality)

### Steps

1. Persist or document ephemeral scope for events, recordings, visuals, releases, notifications, reviews, suites (`TODO-011`).
2. Make schedule claiming atomic (`TODO-017`) if multi-instance is supported.
3. Add event retention limits.
4. Add PostgreSQL integration tests for all stores (`TODO-014` scope).
5. Implement backup/restore procedures (reference: `.ai/BACKUP_AND_RECOVERY.md`).

### Database Migrations

- Add `FOR UPDATE SKIP LOCKED` or advisory lock to `GetDue` (if multi-instance).
- Add event table (if events are persisted).
- Add junction table for test-list membership (if relational integrity is required).

## Phase 4 — Testing

**Motivation:** Zero frontend, Python, integration, or E2E tests means every change risks regression.

### Steps

1. Frontend Vitest + Testing Library scaffold; cover one page, one component, one API function (`TODO-012`).
2. Python pytest scaffold; cover FastAPI endpoints (`TODO-013`).
3. Add `planning` package unit tests using the test patterns in `.ai/TESTING.md`.
4. Add PostgreSQL integration tests with testcontainers or a test database.
5. One E2E smoke test: guided run through Compose.

### Gates

- `go test ./...` passes (including new planning and integration tests).
- `npm test` exits zero.
- `pytest` exits zero.
- E2E smoke test passes.

## Phase 5 — Code Organization

**Motivation:** The 3,556-line `server.go` is the primary maintainability risk.

### Prerequisites

- Phase 4 testing provides regression safety.

### Steps

1. Extract `internal/api/routes.go` (route registration only).
2. Extract `internal/api/handlers_projects.go`, `handlers_runs.go`, `handlers_schedules.go`, `handlers_releases.go`, `handlers_metrics.go`, `handlers_settings.go`.
3. Create application-service structs: `ProjectService`, `PlanningService`, `ExecutionService`, `ScheduleService`.
4. Standardize validation: typed request structs with `DisallowUnknownFields`, JSON error envelope (`TODO-018`).
5. Consolidate LLM client layers into one factory.
6. Remove `class-variance-authority` if confirmed unused (`TODO-021`).
7. Replace `frontend/README.md` (`TODO-022`).
8. Correct root README knowledge-base reference (`TODO-019`).

### Breaking Changes

- PATCH handlers switch from `map[string]interface{}` to typed structs.
- Error responses switch from `http.Error` to JSON envelope.

## Phase 6 — Production Readiness

**Motivation:** Current Compose deployment is not suitable for production traffic.

### Steps

1. Add production Compose overlay with private networking, non-root containers, secret management.
2. Add TLS termination guidance (reverse proxy or internal).
3. Add dependency-aware readiness probe.
4. Add Prometheus metrics endpoint (`.ai/OBSERVABILITY_PLAN.md`).
5. Add structured error responses with correlation IDs.
6. Add graceful shutdown with run draining.
7. Replace bare goroutines with bounded worker pool or Asynq.
8. Create CI/CD pipeline (GitHub Actions or equivalent).
9. Add resource limits to all Compose services.
10. Pin all container images by digest.

## Phase 7 — Documentation Reconciliation

1. Apply all documentation fixes from `.ai/DOCUMENTATION_GAP.md` (`TODO-019`, `TODO-020`).
2. Remove stale variables from `.env.example`; add missing code-consumed variables.
3. Update `docs/docker.md` for current topology.
4. Reconcile embedded product docs with active implementation.
5. Verify `.ai/` knowledge base against source after all phases complete.

## Dependency Map

```
Phase 0 (baseline)
  └─→ Phase 1 (security)
       └─→ Phase 2 (execution)  ← ADR-001, ADR-002
            └─→ Phase 3 (durability)  ← ADR-003, ADR-004
                 └─→ Phase 4 (testing)
                      └─→ Phase 5 (organization)
                           └─→ Phase 6 (production)
                                └─→ Phase 7 (docs)
```

Phases 1 and 4 can begin in parallel with their prerequisites. Phase 5 depends on Phase 4. Phase 6 depends on Phase 3 and Phase 5.

## Risk Assessment

| Phase | Risk of Regression | Risk of Breakage | Mitigation |
|---|---|---|---|
| 0 | Low | Low | Planning package is additive; no existing code changed |
| 1 | Medium | High (auth) | Feature-flag or development-mode escape hatch |
| 2 | High | High (execution) | Keep old path behind flag until new path is proven |
| 3 | Medium | Medium (data) | New migrations forward-only; test on copy of data |
| 4 | Low | Low | Tests are additive |
| 5 | Medium | Medium (API) | Existing tests must pass before and after each extraction |
| 6 | Medium | Medium (ops) | New Compose overlay; keep dev Compose intact |
| 7 | Low | Low | Documentation-only |
