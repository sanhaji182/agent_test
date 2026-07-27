# Observability Plan

**Owner:** Engineering (SRE/observability ownership UNKNOWN)  
**Authoritative sources:** Server entry points, logging/middleware code, health endpoint, SSE handlers, Compose config  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-26  
**Verification performed:** Static inspection; no live telemetry was observed  
**Confidence:** High for current state; target architecture is a recommendation pending owner approval

## Current Observability State

### What exists

| Signal | Implementation | Coverage | Evidence |
|---|---|---|---|
| Structured logging | `log/slog` for lifecycle, scheduler, migration, and webhook events | Backend process events only | `cmd/server/main.go:25-30,45-47,53-66` |
| Request logging | Chi middleware: RequestID, Logger, Recoverer | All HTTP requests | `internal/api/server.go:55-60` |
| Health check | `GET /health` returns `{"status":"ok"}` | Process liveness only; no dependency probes | `internal/api/server.go:268-271` |
| Sidecar health | FastAPI health endpoint | Process liveness only | `sidecar/main.py:28-30` |
| Compose health checks | Per-service health probes | All six Compose services | `docker-compose.yml:29-35,49-55,68-74,81-87,104-110,126-132` |
| SSE heartbeat | 15-second heartbeat on global stream, 2-second poll on per-run stream | SSE connection liveness | `internal/api/server.go:2224-2261,3181-3215` |
| Request correlation | Chi RequestID middleware assigns per-request IDs | HTTP request scope only | `internal/api/server.go:56` |

### What is absent

| Signal | Gap |
|---|---|
| Metrics export | No Prometheus endpoint, no OpenTelemetry metrics, no `expvar` or custom metric registry |
| Distributed tracing | No trace context propagation across the Go process, LLM calls, browser execution, or sidecar calls |
| Dependency health | `/health` does not probe PostgreSQL, Redis, Steel, or sidecar |
| Readiness/Liveness split | One health endpoint serves both purposes with no distinction |
| Error aggregation | No Sentry, error-tracking, or crash-reporting integration |
| Profiling | No `net/http/pprof` endpoint or continuous profiling |
| Log levels | `slog` is used but log level is not configurable through environment |
| Log shipping | No log forwarding to external aggregation; Compose logs are local-only (`make logs` tails `docker compose logs`) |
| Alerting | No alert definitions, no notification integration for infrastructure failures |
| SLO/SLI definitions | None documented |

## Recommended Observability Architecture

### Tier 1 — Immediate (low effort, high value)

These can be added without new dependencies or architectural changes.

1. **Dependency-aware health check.**
   Add a configurable readiness endpoint that probes PostgreSQL (`pgxpool.Ping`), Redis if wired (`PING`), and Steel if wired. Keep the existing `/health` as a lightweight liveness probe.
   ```go
   // Proposed: internal/api/server.go readiness handler
   func (s *Server) handleReadiness(w http.ResponseWriter, r *http.Request) {
       if err := s.pgPool.Ping(r.Context()); err != nil {
           w.WriteHeader(http.StatusServiceUnavailable)
           json.NewEncoder(w).Encode(map[string]string{"status": "not ready", "pg": err.Error()})
           return
       }
       w.WriteHeader(http.StatusOK)
       json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
   }
   ```
   **Evidence motivation:** `GET /health` at `internal/api/server.go:268-271` returns process-only status. A dependency-aware readiness endpoint allows orchestration to route traffic only to healthy instances.

2. **Configurable log level.**
   Read `LOG_LEVEL` from environment; default to `info`. Pass to `slog`.
   **Evidence motivation:** `cmd/server/main.go` uses `slog` with default level. Production debugging requires runtime level changes.

3. **Request ID propagation to logs and error responses.**
   Add the Chi RequestID to `slog` context in middleware so all log lines carry the request ID. Return it in error response bodies.
   **Evidence motivation:** `internal/api/server.go:56` assigns request IDs but handlers use plain `http.Error` with no correlation ID.

4. **`pprof` endpoint (optionally enabled).**
   Register `net/http/pprof` on a separate port or behind a debug flag for production profiling.
   **Evidence motivation:** No profiling endpoint exists. Goroutine leaks, memory growth, and CPU hot spots cannot be diagnosed in production.

5. **Run lifecycle metrics.**
   Track active runs, completed runs, failed runs, and average run duration as atomic counters/gauges exposed through a simple `/metrics` JSON endpoint or Prometheus format.
   ```go
   var (
       runsActive    int64
       runsCompleted int64
       runsFailed    int64
   )
   ```
   **Evidence motivation:** `executeRealRun` at `internal/api/server.go:1858-1953` creates goroutines without instrumentation. The existing `/api/v1/metrics/*` endpoints compute QA analytics over run data rather than exposing operational metrics.

### Tier 2 — Standard (medium effort, structural improvements)

6. **Prometheus metrics endpoint (`/metrics`).**
   Use `promhttp` (minimal dependency) to expose the default Go runtime metrics (goroutines, memory, GC) plus custom application metrics. Separate this from the existing `/api/v1/metrics/*` analytics endpoints.
   **Metrics to expose:**
   - `gotest_runs_total{state="done|failed|active"}`
   - `gotest_run_duration_seconds` (histogram)
   - `gotest_llm_calls_total{provider="anthropic|openai"}`
   - `gotest_llm_call_duration_seconds` (histogram)
   - `gotest_db_connections_active`
   - `gotest_events_dropped_total`
   - `gotest_sse_connections_active`

7. **Structured error responses with correlation IDs.**
   Replace `http.Error(w, "text", code)` with a JSON error envelope containing `request_id`, `code`, and `message`. This allows log-to-response correlation without log inspection.
   **Evidence motivation:** All handlers use ad-hoc `http.Error` strings. See `internal/api/server.go` handlers throughout.

8. **Run execution tracing.**
   Add a span/trace ID to each run at creation time; propagate it through LLM calls and browser execution; include it in event metadata and SSE frames. This enables end-to-end correlation without a full tracing backend.
   **Evidence motivation:** `executeRealRun` at `server.go:1858-1953` uses `context.Background()` with no trace context.

### Tier 3 — Production (higher effort, platform investment)

9. **OpenTelemetry SDK integration.**
   Instrument the Go server, sidecar, and frontend/server components with OTel auto-instrumentation. Export to an OTLP collector (self-hosted or vendor).
   **Spans:** HTTP request → LLM API call → Playwright execution → database query.
   **Metrics:** Replace custom Prometheus metrics with OTel metrics where appropriate.

10. **Centralized log aggregation.**
    Ship structured logs from all services to Loki, Elasticsearch, or a vendor. Configure log retention and search.

11. **Alerting and SLOs.**
    Define service-level indicators (run creation latency, run completion rate, SSE delivery rate, API error rate) and objectives. Wire alerts to the chosen notification channel once `internal/notify` has a production trigger path.

12. **Distributed tracing for multi-service flows.**
    Propagate trace context across:
    - Browser → Go REST → LLM → Playwright
    - Sidecar → Go API
    - GitHub webhook → Go API → async run

## Implementation Sequence

| Phase | Items | Prerequisites |
|---|---|---|
| 1 — Immediate | Dependency-aware readiness, configurable log level, request-ID propagation, optional `pprof`, basic run counters | None; can be done in the existing `internal/api` package |
| 2 — Standard | Prometheus metrics, structured error responses, run trace IDs | `promhttp` dependency; error-response standardization |
| 3 — Production | OpenTelemetry, log aggregation, alerting, distributed tracing | Platform decisions (collector, aggregator, alerting channel); requires owner input on target infrastructure |

## Observability Gating

Before considering the application production-observable:

- [ ] Dependency-aware readiness endpoint exists and is used by orchestration.
- [ ] Application-level metrics (runs, errors, LLM calls, event drops) are exported.
- [ ] Request and trace IDs connect logs, responses, and events.
- [ ] Log level is runtime-configurable.
- [ ] Alert definitions exist for critical failures (database unavailable, LLM quota exhausted, event drops, run queue saturation).
- [ ] Health, metrics, and logs are exercised in a deployed environment.
