# Production Readiness

**Owner:** Engineering (SRE/production ownership UNKNOWN)  
**Authoritative sources:** Compose config, Dockerfiles, Go entry points, environment config, documentation  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-26  
**Verification performed:** Static inspection; no production deployment was observed or tested  
**Confidence:** High for identified gaps; actual production posture and external controls are UNKNOWN

## Current Deployment State

The repository provides exactly one tracked deployment method: `docker compose up` starting six services on a single host (`docker-compose.yml:1-136`). The README explicitly labels this arrangement as optimized for local development and demo with required production hardening steps (`README.md:182-206`).

This document assesses the gap between the current Compose deployment and a production-ready deployment.

## Deployment and Rollback

| Concern | Current state | Gap |
|---|---|---|
| Deployment method | `docker compose up` on a single host | No automated deploy pipeline, no rolling or blue-green deployment strategy |
| Image registry | Local build only; no published images | No tag strategy, no image promotion between environments |
| Rollback | Manual `docker compose down && docker compose up` with prior image | No automated rollback, no canary or gradual rollout |
| Downtime during deploy | Compose restarts services sequentially | No zero-downtime deploy mechanism (connection draining, health-gated rollout) |
| Infrastructure as Code | None found | No Terraform, Pulumi, CloudFormation, Ansible, or other IaC |
| CI/CD | No tracked `.github/workflows` or equivalent | No automated build, test, scan, or deploy pipeline |

## Configuration and Secrets

| Concern | Current state | Gap |
|---|---|---|
| Environment variables | Plain `.env` file interpolation in Compose | No environment-specific overrides (dev/staging/prod); no validation of required variables at deploy time |
| Secrets | Compose `environment:` blocks with literal values; `.env.example` recommends Docker secrets/external managers but none are implemented | Database password, API key, LLM API key, sidecar key, and project credentials are plain environment strings |
| Configuration drift | Several `.env.example` variables have no runtime consumer; several code-consumed variables are absent from `.env.example` | Operators cannot confidently configure the application from documentation alone |
| Feature flags | `GOTEST_AI_PLANNING`, `GOTEST_APPROVED_CASE_RUNNER` consumed by code but absent from Compose and `.env.example` | Desired runtime behavior is undocumented and default paths are synthetic |
| API key | Empty default disables all authentication | Production deployment depends on operator remembering to override the default |

## Networking and TLS

| Concern | Current state | Gap |
|---|---|---|
| TLS | None configured | All inter-service and client-facing traffic is plain HTTP |
| Frontend API URL | Hard-coded at `NEXT_PUBLIC_API_URL` during build (`frontend/Dockerfile:10-11`) | Cannot retarget a built frontend image to a different API origin without rebuild |
| Database TLS | `sslmode=disable` | PostgreSQL connections are unencrypted |
| Redis TLS | Not configured in repository | Unencrypted if used |
| CORS | Hard-coded wildcard (`*`) at `internal/api/server.go:98-108` | No origin restriction; documented `CORS_ALLOWED_ORIGINS` config does not exist in code |
| Service exposure | All six services publish to all host interfaces | No private network, no reverse proxy, no ingress controller |
| Rate limiting | Explicitly not implemented (`README.md:204`) | No protection against abuse, credential stuffing, or resource exhaustion |

## Timeouts, Resilience, and Graceful Shutdown

| Concern | Current state | Gap |
|---|---|---|
| HTTP server timeouts | Configured: ReadHeaderTimeout 5s, ReadTimeout 15s, WriteTimeout 30s, IdleTimeout 60s | — ✅ (TODO-018) |
| Request body limits | `bodyLimitMiddleware` 1 MiB on `/api/v1` group | — ✅ (TODO-018) |
| Graceful shutdown | SIGINT/SIGTERM → cancel scheduler → Shutdown with 10s deadline | — ✅ (TODO-018) |
| Run execution timeout | No deadline or cancellation propagation beyond individual Playwright action timeouts | A hung LLM call or browser action holds the goroutine indefinitely |
| Outbound LLM client timeouts | No explicit timeout on OpenAI-compatible client construction (`internal/agent/llm_openai.go:29-34`; `internal/ai/client.go:51`) | Hung upstream blocks run goroutines |
| Goroutine lifecycle | Semaphore-bounded (default 10, `MAX_CONCURRENT_RUNS`); panic recovery via `Agent.Launch` | — ✅ (AUDIT S-01 + TODO-007) |
| Retry strategy | No retry at the orchestrator level; inner Playwright actions retry up to 3 times | Transient LLM or database errors cause immediate terminal failure |
| Circuit breaking | None | No backpressure or degradation strategy |

## Resource Management

| Concern | Current state | Gap |
|---|---|---|
| Container resource limits | None specified in Compose or Dockerfiles | No CPU, memory, or I/O limits; a single run can consume unbounded host resources |
| Concurrency cap | Semaphore at `launchRun` (non-blocking, warns on overflow, `MAX_CONCURRENT_RUNS`=10) | — ✅ (AUDIT S-01) |
| Event store growth | Capped at 10,000 events/run with FIFO pruning; persisted to PostgreSQL | — ✅ (AUDIT S-02 + ADR-003) |
| Temporary artifacts | Videos under `/tmp/agent_test/videos` and `/data/videos` without cleanup | Disk exhaustion from accumulated recordings, screenshots, and temp files |
| Playwright installation | Re-installed on every run (`internal/agent/playwright_runner.go:34-37`) | Latency spike per run; concurrent installations compete for shared state |

## Persistence, Backup, and Recovery

| Concern | Current state | Gap |
|---|---|---|
| PostgreSQL backups | None found (`docs/docker.md:92-101` documents only destructive volume reset) | No backup schedule, no pg_dump automation, no point-in-time recovery, no restore procedure |
| Volume persistence | `pgdata` and `app_data` are Docker named volumes | Survives container restart but not host or volume failure; no off-host replication |
| Migration safety | Migrations are non-atomic and non-fatal on failure (`internal/db/migrate.go:45-69`; `cmd/server/main.go:28-34`) | Schema can be partially applied; database fallback silently degrades to memory |
| Restart durability | Events, recordings, visuals, releases, notifications, reviews, suites, and sidecar jobs are memory-only | Restart loses all operational state; this affects monitoring, audit trails, and alert history |
| Retention and pruning | None implemented | Runs, settings, projects, and artifact files accumulate without bound |
| Disaster recovery | No documented runbook | No failover, no replication, no restore validation |

## Horizontal Scaling

| Concern | Current state | Gap |
|---|---|---|
| Multi-instance support | In-process event bus, local SSE subscribers, singleton scheduler (`internal/events/store.go:42-145`; `cmd/server/main.go:41-67`) | Multiple instances partition state and duplicate schedules |
| Schedule claiming | `SELECT ... WHERE enabled AND next_run_at <= now` without lock or lease (`internal/schedule/store.go:245-269`) | Concurrent instances create duplicate runs |
| Session affinity | Per-run SSE must reach the instance that hosts the event channel | No session-routing mechanism |
| Shared state | Run state is in PostgreSQL (shared), but events/operational stores are local | Cross-instance visibility of execution progress is broken |

## Observability

| Concern | Current state | Gap |
|---|---|---|
| Health check | `GET /health` returns `{"status":"ok"}` without probing PostgreSQL, Redis, Steel, or sidecar | Unhealthy dependencies are invisible to orchestration |
| Readiness/Liveness | No separate readiness or liveness endpoints | Orchestration cannot distinguish "starting but not ready" from "dead" |
| Structured logging | `slog` used for lifecycle events; request logging via Chi middleware | No correlation IDs propagated to downstream calls; no log level configuration by environment |
| Metrics export | `/api/v1/metrics/*` are product QA analytics, not infrastructure telemetry | No Prometheus endpoint, no OpenTelemetry integration, no metric export for platform monitoring |
| Distributed tracing | None found | Cannot trace a request across the Go process, LLM calls, and browser execution |
| Error reporting | None found (no Sentry, error aggregation, or alerting integration) | Production failures are visible only through log inspection |
| Profiling | None found | No `pprof` endpoint or continuous profiling |

## Operational Runbooks

| Concern | Current state | Gap |
|---|---|---|
| Startup procedure | `make up` or `docker compose up -d` | No pre-flight checks, no environment validation, no migration dry-run |
| Shutdown procedure | `docker compose down` | No in-flight run drain, no data integrity verification |
| Incident response | None documented | No escalation paths, no log collection procedure, no state-recovery steps |
| Monitoring and alerting | None configured | No alert definitions, no notification channels, no SLO definitions |
| Capacity planning | None | No resource-usage baselines, no growth projections |

## Production Readiness Summary

| Category | Assessment |
|---|---|
| Deployment | Not production-ready |
| Configuration and secrets | Not production-ready |
| Networking and TLS | Not production-ready |
| Timeouts and resilience | Not production-ready |
| Resource management | Not production-ready |
| Persistence and backup | Not production-ready |
| Horizontal scaling | Not production-ready |
| Observability | Not production-ready |
| Operational runbooks | Not production-ready |

The current repository state is a functioning local-development and demo stack. Every category above requires deliberate engineering work before any environment exposed to production traffic or data. The gap list is defined; prioritization of which gaps to close first depends on owner-approved production requirements (RPO/RTO, SLOs, expected traffic, tenancy model, deployment environment) that are all currently UNKNOWN.
