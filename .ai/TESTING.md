# Testing Strategy

**Owner:** Engineering (QA/testing owner pending nomination)  
**Authoritative sources:** Existing test files, Go module layout, frontend package.json, sidecar directory  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-26  
**Verification performed:** Static test-inventory inspection; no test suite run was completed  
**Confidence:** High for what exists; execution results are UNKNOWN

## Current Coverage Inventory

### Go tests (exist, not yet executed)

| Package | Test file | Coverage area | Notes |
|---|---|---|---|
| `internal/agent` | `agent_test.go` | Agent state machine, mocks | Interface-mock approach for LLM/Runner |
| `internal/api` | `api_test.go` | Health endpoint, auth, create/run, full lifecycle | Uses in-memory stores; exercises project→plan→case→list→run→schedule→monitor path |
| `internal/api` | `planning_test.go` | Plan generation, case approval, maintenance, refinement, proposals, test lists, schedules | Primary regression test for planning lifecycle |
| `internal/auth` | `auth_test.go` | JWT creation/validation, API key generation | |
| `internal/compare` | `compare_test.go` | Run comparison | |
| `internal/events` | `events_test.go` | Event emit/subscribe | |
| `internal/gitdiff` | `gitdiff_test.go` | Diff analysis | |
| `internal/intelligence` | * | Risk scoring, confidence | Package-level tests |
| `internal/metrics` | * | Statistics computation | |
| `internal/schedule` | `schedule_test.go` | Store operations | Uses memory store |
| `internal/visual` | * | Visual comparison | |

### Missing tests by surface

| Surface | Current state | Gap |
|---|---|---|
| PostgreSQL stores | No integration tests found | All existing tests use memory stores. Schema, query, and migration correctness untested against a real database. |
| Planning package (reconstructed) | Package tests absent; lifecycle covered through API test | No direct unit tests for `MemoryStore` or `DBStore`. |
| Frontend | No test scripts, no test dependencies in `package.json` | Zero unit, component, integration, or E2E coverage. |
| Python sidecar | No `*test*.py` files found | Zero coverage for FastAPI endpoints, graph nodes, or executor callbacks. |
| End-to-end flows | None found | No cross-service E2E test exercising Compose-deployed full stack. |
| Performance/load | None found | No throughput, latency, or resource-usage tests. |
| Security | None found | No auth bypass, injection, or access-control tests. |

## Target Testing Tiers

### Tier 1 — Unit (immediate priority)

Every package boundary should have focused unit tests with mocked dependencies.

**Go packages:**

1. `internal/planning` — direct `MemoryStore` and `DBStore` unit tests (creation, retrieval, update, concurrent mutation, deep-copy isolation).
2. Agent state machine transitions through `Agent.Execute` with mock LLM/Runner.
3. Domain/calculation packages: compare, metrics, intelligence, visual — verify determinism and edge cases.
4. Config loading with various environment values.

**Frontend components:**

5. Shared UI components (Badge, Card, Section, Tabs, ScreenshotStrip, Timeline) using Vitest + Testing Library.
6. `frontend/src/lib/api.ts` typed fetch functions with mocked `fetch`.
7. Settings, run-detail, and create pages with mocked API responses and SSE events.

**Sidecar:**

8. FastAPI endpoint handlers with mocked dependencies.
9. Individual graph node functions (planner, writer, critic, executor, fixer) with mocked LLM responses.

### Tier 2 — Integration (high priority)

Each persistence and transport adapter against its real backing service.

10. `planning.DBStore` against a test PostgreSQL instance — verify migration schema, CRUD, transaction rollback, FK constraint, JSON round-trip, and null handling for each entity.
11. Existing `db.Store`, `project.DBStore`, and `schedule.DBStore` against PostgreSQL — same coverage.
12. Migration runner: verify full application, partial failure, concurrency, and at-least-once idempotency.
13. API handlers through `httptest` with real memory stores, exercising error paths, auth, validation, and response format.
14. SSE event replay, subscription, drop, and reconnect semantics.

### Tier 3 — End-to-End (medium priority)

15. Compose-deployed full stack: guided run creation (create → `202` → SSE/run detail → terminal state).
16. Project→plan→approve→case→list→schedule→run flow end-to-end.
17. Settings update → AI provider test → run execution with that provider.
18. GitHub webhook delivery with and without HMAC secret.
19. Schedule due processing cycle.

### Tier 4 — Specialized (targeted)

20. Concurrent schedule claiming under multiple backend instances or overlapping scheduler ticks.
21. SSE behavior under connection loss, slow clients, and resource exhaustion.
22. Auth bypass and malformed-input fuzzing for critical endpoints.
23. Load test: sustained run creation rate, concurrent SSE viewers, list performance with 10k+ runs.

## Tooling Recommendations

| Layer | Tool | Rationale |
|---|---|---|
| Go unit/integration | `go test`, `httptest`, `testcontainers-go` for PostgreSQL | Already in use; pgx pool setup needed for integration tests |
| Frontend unit/component | Vitest + `@testing-library/react` | Lightweight, Vite-native, matches Next.js ecosystem |
| Frontend E2E | Playwright Test | Already a project dependency; can drive full-stack flows |
| Sidecar | pytest + `httpx` test client + `pytest-asyncio` | Standard Python async testing; mock `langchain-anthropic` |
| CI | Any runner that can start PostgreSQL + app | `go test ./...` + `npm test` + `pytest` as minimum gate |

## Testing Patterns from Existing Code

### API test server setup (reuse this pattern)

```go
// internal/api/api_test.go — existing pattern
func newTestServer() *api.Server {
    store := db.NewMemoryStore()
    return api.NewServer(&config.Config{}, store, nil)
}
```

Extend with a PostgreSQL-backed variant:

```go
func newPGTestServer(t *testing.T) *api.Server {
    t.Helper()
    pool := testPGPool(t)
    store, _ := db.NewStore(context.Background(), pool)
    db.RunMigrations(context.Background(), pool)
    return api.NewServer(&config.Config{}, store, db.NewSettingsStore(pool))
}
```

### Planning store tests (new, follow existing project store tests)

Create `internal/planning/store_test.go`:

```go
func TestDraftPlanLifecycle(t *testing.T) {
    s := planning.NewMemoryStore()
    p := &planning.DraftPlan{ProjectID: "p1", Cases: []planning.DraftCase{{Title: "Login"}}}
    require.NoError(t, s.CreateDraft(t.Context(), p))
    require.NotEmpty(t, p.ID)

    got, err := s.GetDraft(t.Context(), p.ID)
    require.NoError(t, err)
    require.Equal(t, p.ID, got.ID)
    require.Len(t, got.Cases, 1)

    got.Cases[0].Title = "mutated"
    fresh, _ := s.GetDraft(t.Context(), p.ID)
    require.Equal(t, "Login", fresh.Cases[0].Title, "clone must prevent aliasing")
}
```

## Verification Gating

Minimum gate before considering the repository testable:

1. `go build ./...` — confirmed passing (2026-07-26)
2. `gofmt -d .` — zero diff (pending)
3. `go test -count=1 ./...` — all pass (pending)
4. `go vet ./...` — zero findings (confirmed for `internal/planning`; full scope pending)

Pre-commit or CI gates should add:

5. `npm test` (once frontend tests exist)
6. `pytest` (once sidecar tests exist)
7. `npm run lint` (ESLint for frontend)

## Priority Order

1. `go test` all existing packages — confirms current baseline
2. `internal/planning/store_test.go` — locks reconstructed package behavior
3. Frontend unit test scaffold — one page, one component, one API function
4. PostgreSQL integration test scaffold — testcontainers + one migration/CRUD cycle
5. One E2E smoke test — guided run through Compose
6. Sidecar pytest scaffold
7. Remaining packages and flows by risk
