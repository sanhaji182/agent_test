# Performance Plan

**Owner:** Engineering  
**Authoritative sources:** Execution code paths, runner implementations, database queries, SSE handlers, event store  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-26  
**Verification performed:** Static code-path and resource-usage analysis; no profiling, load testing, or production traffic observation was performed  
**Confidence:** High for identified patterns; actual throughput/latency numbers and production cardinality are UNKNOWN

## Current Performance Characteristics

### Known bottlenecks (static analysis)

| Bottleneck | Location | Mechanism | Severity |
|---|---|---|---|
| Playwright re-installation per run | `internal/agent/playwright_runner.go:34-37` | `playwright.Install()` called on every `Run()` invocation | MEDIUM |
| Unbounded goroutine creation per run | `internal/api/server.go:1581-1589,2186-2190` | Each create/rerun/webhook/approved-case launches a `go` goroutine with no pool or cap | HIGH |
| N+1 run fetches in analytical endpoints | `internal/api/server.go:2841-2865` | `ListRuns` then `GetRun` for every run in hotspots, flaky, and trend handlers | MEDIUM |
| In-memory scan over runs for monitoring | `internal/api/server.go:2064-2147` | `handleMonitoringSummary` iterates all runs in Go to compute list health | MEDIUM |
| Database poll in per-run SSE | `internal/api/server.go:2224-2261` | Every 2 seconds per active SSE connection queries run state | MEDIUM |
| Event store unbounded growth | `internal/events/store.go:42-100` | Append-only slice with no retention, no eviction; non-blocking sends drop events on full channel | MEDIUM |
| JSON round-trips for planning entities | `internal/planning/db.go` | Testcase steps/assertions/tags, draft cases, and proposal snapshots are marshaled/unmarshaled as JSONB on every read/write | LOW |
| Direct Playwright in API process | `internal/api/server.go:1922-1929` | Browser execution shares the API process; no resource isolation | MEDIUM |

### Concurrency characteristics

| Pattern | Evidence | Risk |
|---|---|---|
| Goroutine per run (unbounded) | `internal/api/server.go:1581-1589` | Memory, FD, and CPU exhaustion under burst |
| Mutex-serialized event store | `internal/events/store.go:42-48` | Write contention under high event volume |
| Non-blocking event subscriber sends | `internal/events/store.go:77-91` | Silent event loss under slow SSE clients |
| Single scheduler goroutine | `cmd/server/main.go:41-67` | Single point of scheduling; no distributed coordination |
| `pgxpool` connection pool | `internal/db/store.go:24-33` | Go default pool sizing; not tuned for expected concurrency |

## Recommended Improvements

### Tier 1 — Immediate (low effort, high impact, no architectural change)

1. **Install Playwright once at startup or image build.**
   Move `playwright.Install()` from `Run()` to a startup initialization step. Verify installation with a readiness check. Eliminates per-run latency and concurrent-install contention.
   **Evidence:** `internal/agent/playwright_runner.go:34-37`

2. **Add bounded goroutine concurrency.**
   Use a semaphore channel (`make(chan struct{}, maxConcurrent)`) or worker pool to cap concurrent run executions. Reject or queue runs above the cap with `429`/`503`.
   **Evidence:** `internal/api/server.go:1581-1589`

3. **Reduce per-run SSE poll frequency or use event-driven state.**
   Increase the poll interval from 2 seconds to 5-10 seconds, or push terminal state changes as events rather than polling.
   **Evidence:** `internal/api/server.go:2224-2261`

4. **Add event store retention limits.**
   Cap per-run events at a fixed maximum (e.g., 1,000) and total global events with a configurable limit. Drop oldest first or use a ring buffer.
   **Evidence:** `internal/events/store.go:42-100`

5. **Add `http.Server` timeouts.**
   Replace `http.ListenAndServe` with a configured `http.Server` that has `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, and `IdleTimeout`.
   **Evidence:** `cmd/server/main.go:44-46`

### Tier 2 — Standard (medium effort, moderate benefit)

6. **Push analytical queries to SQL.**
   Replace Go-level filtering/sorting of runs in monitoring and metric endpoints with PostgreSQL queries that return only the needed rows. Add database indexes for common filter/sort combinations.
   **Evidence:** `internal/api/server.go:2064-2147`

7. **Cache monitoring/comparison results.**
   Cache pass rates, hotspot results, and comparison diffs with a short TTL (30-60 seconds) rather than recomputing on every request.

8. **Move browser execution to a separate worker.**
   Run Playwright in a dedicated worker container (Docker runner or Steel) with resource limits, isolating API responsiveness from browser resource consumption.
   **Evidence:** `internal/api/server.go:1922-1929` hard-codes local Playwright. The `DockerRunner` at `internal/runner/docker.go` already implements this pattern but is not wired into the primary web path.

9. **Add PostgreSQL connection pool tuning.**
   Expose `PG_MIN_CONNS`, `PG_MAX_CONNS`, and `PG_MAX_CONN_LIFETIME` through configuration. Default to reasonable values based on expected concurrent runs.
   **Evidence:** `internal/db/store.go:24-33` uses default pool configuration.

### Tier 3 — Production (higher effort, platform investment)

10. **Replace in-process goroutines with a durable queue.**
    Route all async execution (create, rerun, schedule, webhook) through Redis/Asynq or a database-backed job table. This provides durability, retry, bounded concurrency, and observability. The `internal/queue` package already implements the Asynq pattern.
    **Evidence:** `internal/queue/worker.go:27-92`

11. **Implement horizontal scaling support.**
    Move events to Redis Streams or a database-backed event table for cross-instance delivery. Add session affinity or fan-out for SSE. Atomic schedule claiming via `FOR UPDATE SKIP LOCKED` or advisory locks.
    **Evidence:** `internal/events/store.go:42-145`; `internal/schedule/store.go:245-269`

12. **Add performance regression tests.**
    Benchmark run creation latency, list endpoints with increasing dataset sizes, SSE throughput, and concurrent run execution. Run as part of CI to detect regressions.

## Performance Baselines (UNKNOWN)

No throughput, latency, or resource-usage baselines exist. Before tuning:

- Measure run creation end-to-end latency (POST → 202).
- Measure run execution duration distribution (realistic test scenarios).
- Measure monitoring summary response time with 100 / 1,000 / 10,000 runs.
- Measure SSE event delivery latency (emit → client receive).
- Profile memory usage over sustained operation (1 hour, 100 runs).
- Profile goroutine count under concurrent run creation (10 / 50 / 100 concurrent).

## Quick Wins Summary

| # | Change | Effort | Risk |
|---|---|---|---|
| 1 | Move Playwright install to startup | Low | Low |
| 2 | Add goroutine semaphore | Low | Low (backward-compatible; reject above cap) |
| 3 | Reduce SSE poll frequency | Low | Low |
| 4 | Add event retention cap | Low | Medium (lossy; acceptable until durable events) |
| 5 | Add server timeouts | Low | Low (standard Go idiom) |
| 6 | Push filtering to SQL | Medium | Medium (query correctness) |
| 7 | Cache analytics results | Medium | Low |
| 8 | Worker-based browser execution | High | Medium (changes artifact paths/test flow) |
| 9 | Tune connection pool | Low | Low |
| 10 | Durable queue | High | High (changes execution model, may require idempotency) |
| 11 | Horizontal scaling | High | High (architecture change) |
| 12 | Performance regression tests | Medium | Low |
