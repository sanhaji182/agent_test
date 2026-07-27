# Project State

**Owner:** Engineering  
**Authoritative sources:** Tracked source and manifests listed below  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-26  
**Verification performed:** Static inspection only; no successful build, test, runtime, or vulnerability scan  
**Confidence:** High for static implementation presence; runtime health is UNKNOWN

## Overall Project Health

**Status:** **Compiling; not release-ready**  
**Confidence:** High for build result; runtime health partly UNKNOWN.

**Verified:** The repository contains a broad self-hosted testing product surface: a Go HTTP API, Next.js dashboard, PostgreSQL schema, SSE event delivery, Docker/Steel/local runners, an MCP binary, and an optional LangGraph sidecar (`cmd/server/main.go:15-67`; `internal/api/server.go:38-261`; `frontend/package.json:5-29`; `docker-compose.yml:1-136`).

**Verified:** The previously missing `internal/planning` package was reconstructed from surviving repository evidence and `go vet ./internal/planning` plus `go build ./...` passed (2026-07-26). `gofmt` and focused lifecycle regression (`TestGenerateApprovePlanLifecycle`) are pending infrastructure availability but the structural compilation blocker is removed. See `CHANGELOG_AI.md` and `TODO-001`.

## Verification State

| Check | Status | Evidence |
|---|---|---|
| Static repository discovery | Completed | `.ai/DISCOVERY.md`; tracked-source citations |
| Architecture/security audit | Completed statically | Root `AUDIT.md` is a point-in-time secondary report; canonical risks are linked through TODOs |
| `gofmt` formatting check | Passed | `gofmt -d` produced zero diff on all planning files (2026-07-26) |
| `go vet ./internal/planning` | Passed | Zero findings (2026-07-26) |
| `go build ./cmd/server` | Passed | Planning package restored; full module compiles (2026-07-26) |
| `go test ./...` | Partially verified | `go build` confirmed; lifecycle regression `TestGenerateApprovePlanLifecycle` and full suite pending infrastructure |
| `npm audit`, `govulncheck`, `pip-audit`, `gitleaks` | Pending | Shell safety classifier prevented execution throughout |
| Frontend lint/build/tests | UNKNOWN / not completed | No successful command evidence |
| Python tests | UNKNOWN; no test files found | `sidecar/` static file sweep |
| `npm audit`, `govulncheck`, `pip-audit`, native `gitleaks` | UNKNOWN / not completed | Scanner commands did not execute |

## Completed in Tracked Source

“Completed” here means **present in static source**, not runtime-verified.

### Product and transport surface

- **Verified:** Chi-based HTTP server, request IDs, access logging, panic recovery, and REST/SSE route registration (`internal/api/server.go:55-60,124-261`).
- **Verified:** Project, test-plan, test-case, change-proposal, test-list, run, schedule, release, metrics, review, suite, settings, AI provider, export, demo, health, webhook, and video handlers are registered (`internal/api/server.go:124-261`). Detailed contracts: [`API.md`](API.md).
- **Verified:** Next.js App Router dashboard has pages for the control room, creation, projects, tests, runs, comparison, suites, monitoring, risk, releases, reviews, alerts, exports, docs, and settings (`frontend/src/app/**/page.tsx`; `frontend/src/components/sidebar.tsx:11-24`).
- **Verified:** MCP stdio server defines four tools: `run_tests`, `analyze_project`, `generate_test_plan`, and `get_run_status` (`internal/mcp/server.go:33-143`).

### Persistence and domain logic

- **Verified:** PostgreSQL migrations define ten tables and indexes through migration `008_change_proposals.sql` (`internal/db/migrations/001_init.sql`–`008_change_proposals.sql`). See [`DATABASE.md`](DATABASE.md).
- **Verified:** Run, project, schedule, and settings PostgreSQL stores exist; run/project/schedule memory implementations exist (`internal/db/store.go`; `internal/project/store.go`; `internal/schedule/store.go`; `internal/db/settings_store.go`).
- **Verified:** Comparison, metrics, intelligence/risk, visual, report, event, workflow, and notification packages exist (`internal/compare`; `internal/metrics`; `internal/intelligence`; `internal/visual`; `internal/report`; `internal/events`; `internal/workflow`; `internal/notify`).

### Execution options

- **Verified:** Reusable `agent.Agent` state machine and LLM/Runner interfaces exist (`internal/agent/agent.go:110-175,223-345`).
- **Verified:** Direct Playwright, Docker, and Steel runner implementations exist (`internal/agent/playwright_runner.go`; `internal/runner/docker.go`; `internal/runner/steel.go`).
- **Verified:** Redis/Asynq queue code, sidecar client/graph, vision client, Braintrust logger, and JWT code exist. Their runtime wiring varies; see [`ARCHITECTURE.md`](ARCHITECTURE.md) and [`DEPENDENCIES.md`](DEPENDENCIES.md).

## Work In Progress / Partial

- **Verified:** Primary web execution directly creates LLM and Playwright clients rather than invoking the richer reusable Agent (`internal/api/server.go:1858-1953`; `internal/agent/agent.go:223-345`).
- **Verified:** API runner explicitly simulates every generated file as passed (`internal/agent/api_runner.go:14-33`).
- **Verified:** Default approved-case path sleeps through steps and synthesizes passing assertions unless `GOTEST_APPROVED_CASE_RUNNER=docker` is set (`internal/api/server.go:1203-1247`).
- **Verified:** Non-list schedule run-now and due-schedule branches create `idle` runs without launching execution (`internal/api/server.go:2671-2691,2721-2739`).
- **Verified:** Notification delivery code exists, but no production caller of `TriggerFailure` was found (`internal/notify/store.go:58-99`).
- **Verified:** Releases, notifications, reviews, suites, events, recordings, and visuals are process-memory stores even when PostgreSQL is available (`internal/api/server.go:71-85`).
- **Verified:** Queue, Steel, sidecar, vision, Braintrust, and JWT implementations are absent from the primary `cmd/server` composition path (`cmd/server/main.go:15-49`).
- **UNKNOWN:** Which execution path and browser backend are intended to become canonical. See `ADR-001` and `ADR-002` in [`DECISIONS.md`](DECISIONS.md).

## Blocked Work

### Critical

- [`TODO-001`](TODO.md#todo-001-restore-the-missing-planning-package) — restore or resolve the missing `internal/planning` package; blocks compilation and planning-related routes.
- [`TODO-002`](TODO.md#todo-002-make-production-api-authentication-fail-closed) — make production authentication fail closed.
- [`TODO-003`](TODO.md#todo-003-secure-or-internalize-the-sidecar) — secure or internalize the unauthenticated sidecar.

### High

- [`TODO-004`](TODO.md#todo-004-separate-the-github-webhook-secret) — separate GitHub webhook HMAC secret from API authentication.
- [`TODO-005`](TODO.md#todo-005-resolve-dashboard-authentication-transport) — align dashboard REST/SSE with backend authentication.
- [`TODO-006`](TODO.md#todo-006-bind-llm-credentials-to-approved-provider-origins) — prevent stored LLM credentials from being sent to caller-selected origins.
- [`TODO-007`](TODO.md#todo-007-consolidate-execution-paths) — choose and consolidate the canonical orchestrator.
- [`TODO-008`](TODO.md#todo-008-fix-non-list-schedule-execution) — execute non-list scheduled runs.
- [`TODO-009`](TODO.md#todo-009-implement-real-run-deletion) — implement run and artifact deletion.
- [`TODO-010`](TODO.md#todo-010-stop-reporting-synthetic-test-success) — eliminate synthetic pass results.

## Known Issues

| ID | Priority | Summary | Canonical task |
|---|---:|---|---|
| `RISK-001` | Critical | Missing planning implementation; expected build blocker | `TODO-001` |
| `RISK-002` | Critical | API auth bypasses when `API_KEY` is empty | `TODO-002` |
| `RISK-003` | Critical | Published sidecar has no inbound authentication and can present backend key | `TODO-003` |
| `RISK-004` | High | Configurable LLM base URL can reuse stored credential | `TODO-006` |
| `RISK-005` | High | Five divergent execution paths | `TODO-007` |
| `RISK-006` | High | Scheduled non-list runs remain idle | `TODO-008` |
| `RISK-007` | High | Run deletion endpoint is a no-op | `TODO-009` |
| `RISK-008` | High | Mock/default runners report synthetic success | `TODO-010` |
| `RISK-009` | Medium | Process-local operational state is lost on restart | `TODO-011` |
| `RISK-010` | Medium | Frontend and Python automated coverage absent | `TODO-012`, `TODO-013` |
| `RISK-011` | Medium | Dependency/build verification incomplete | `TODO-014`, `TODO-015` |
| `RISK-012` | Medium | Documentation conflicts with tracked implementation | `TODO-019`, `TODO-020` |

## Current Priorities

1. Restore buildability and capture fresh build/test evidence (`TODO-001`, `TODO-014`).
2. Close default authentication and credential-boundary failures (`TODO-002`–`TODO-006`).
3. Decide and converge execution architecture (`ADR-001`, `ADR-002`, `TODO-007`–`TODO-010`).
4. Decide persistence and multi-instance expectations (`ADR-003`, `ADR-004`, `TODO-011`).
5. Add missing automated verification and dependency reproducibility (`TODO-012`–`TODO-015`).
6. Reconcile user/operator docs with code (`TODO-019`, `TODO-020`).

## UNKNOWN Information

- `UNK-001` — Canonical execution orchestrator and browser backend.
- `UNK-002` — Single-administrator versus multi-user/tenant product model.
- `UNK-003` — Single-instance versus horizontally scaled deployment requirement.
- `UNK-004` — Required durability/retention for events, artifacts, releases, reviews, suites, and notifications.
- `UNK-005` — External production ingress, TLS, firewall, secret management, backups, telemetry, and egress controls.
- `UNK-006` — Actual dependency CVEs and current build/test results.
- `UNK-007` — Whether credential fields contain live secrets or secret-manager references.

See [`ROADMAP.md`](ROADMAP.md) for milestone sequencing and [`TODO.md`](TODO.md) for executable work.
