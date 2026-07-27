# Architecture Decisions

**Owner:** Engineering  
**Authoritative sources:** Accepted ADR text plus linked tracked-source evidence  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-26  
**Verification performed:** Static inspection and Git-history review  
**Confidence:** High that no formal ADR set existed; rationale for historical implementation choices is mostly UNKNOWN

## ADR Policy

Each durable decision must include:

- ID and title
- Date
- Status: `Proposed`, `Accepted`, `Superseded`, `Rejected`, or `Decision needed`
- Context
- Decision
- Alternatives considered
- Consequences
- Evidence
- Owner
- Related TODOs/risks
- Supersedes / superseded by

Do not convert an inferred implementation pattern or commit message into an accepted decision without owner confirmation. Never delete accepted or superseded ADR history.

## Decision Register

### ADR-000 — Use `.ai/` as the tracked internal engineering knowledge base

**Date:** 2026-07-26  
**Status:** Accepted  
**Owner:** Engineering

**Context**

The root README points to an absent `planning/` directory, which is ignored by Git (`README.md:32`; `.gitignore:72-73`). Historical `AGENT.md` and `gotest-agent-master-blueprint.md` were deleted in commit `a0eae33` without tracked replacements. The user requested a continuously maintained, evidence-backed internal knowledge base.

**Decision**

Use tracked `.ai/` Markdown documents as the primary internal engineering knowledge base. Source code, manifests, migrations, lockfiles, and observed verification results remain more authoritative than the knowledge base.

**Alternatives considered**

1. Restore `planning/` — rejected because it is ignored and historical content is stale/aspirational.
2. Keep root `DISCOVERY.md`, `AUDIT.md`, and `PROJECT_STATE.md` as parallel authorities — rejected because they overlap and were untracked point-in-time reports.
3. Use external wiki only — rejected because it would not evolve atomically with repository changes.

**Consequences**

- Internal engineering facts have a trackable home.
- Every meaningful implementation must update affected `.ai/` documents and append to `CHANGELOG_AI.md`.
- The root README must be corrected to reference `.ai/` (`TODO-019`).
- Route/schema/version inventories still require periodic source reconciliation.

**Evidence:** `README.md:32`; `.gitignore:72-73`; Git history at `a0eae33`, `4401efe`; current `.ai/README.md`  
**Related:** `TODO-019`, `RISK-012`

---

### ADR-001 — Choose the canonical execution orchestrator

**Date:** 2026-07-26  
**Status:** Decision needed  
**Owner:** UNKNOWN

**Context**

Five execution paths have materially different behavior:

- Web direct orchestration (`internal/api/server.go:1858-1953`)
- Reusable Agent/MCP (`internal/agent/agent.go:223-345`; `internal/mcp/server.go:74-143`)
- Approved-case path (`internal/api/server.go:1203-1286`)
- Asynq worker (`internal/queue/worker.go:27-92`)
- LangGraph sidecar (`sidecar/graph.py:18-42`)

**Decision:** UNKNOWN. No owner-confirmed canonical path exists.

**Alternatives requiring evaluation**

1. Make reusable `agent.Agent` the canonical orchestrator.
2. Extract a new application service around existing Agent interfaces.
3. Make Asynq the mandatory durable execution boundary.
4. Make sidecar/LangGraph canonical for advanced and simple modes.
5. Preserve separate workflows intentionally and document their contracts.

**Consequences of deferral**

Behavior, state machines, retries, provider selection, artifacts, and persistence continue to diverge.

**Evidence:** paths above  
**Related:** `TODO-007`, `TODO-010`, `RISK-005`, `UNK-001`

---

### ADR-002 — Choose the canonical browser/test execution backend

**Date:** 2026-07-26  
**Status:** Decision needed  
**Owner:** UNKNOWN

**Context**

Direct Playwright-Go, Docker Playwright, Steel Browser, and a mocked API runner coexist (`internal/agent/playwright_runner.go`; `internal/runner/docker.go`; `internal/runner/steel.go`; `internal/agent/api_runner.go`). The active web path hard-codes direct Playwright, while MCP uses Docker and product docs claim Steel.

**Decision:** UNKNOWN.

**Alternatives requiring evaluation**

- Direct Playwright in the API process
- Isolated Docker runner
- Steel Browser service
- Per-test-type backend selection behind one Runner factory

**Consequences of deferral**

Resource isolation, network policy, video paths, runtime dependencies, and operational scaling remain inconsistent.

**Evidence:** `internal/api/server.go:1922-1929`; runner paths above; `frontend/src/lib/docs.ts:31`  
**Related:** `ADR-001`, `TODO-007`, `TODO-016`, `UNK-001`

---

### ADR-003 — Define durable versus ephemeral product state

**Date:** 2026-07-26  
**Status:** Decision needed  
**Owner:** UNKNOWN

**Context**

Runs, projects, planning entities, schedules, and settings have intended PostgreSQL paths. Events, recordings, visuals, releases, notifications, reviews, and suites are process-memory stores (`internal/api/server.go:62-85`). Sidecar jobs are also memory-only (`sidecar/main.py:11-12`).

**Decision:** UNKNOWN which entities must survive restarts and what retention applies.

**Alternatives requiring evaluation**

1. Persist all user-visible entities.
2. Persist only audit-critical entities; explicitly label transient projections.
3. Keep memory stores but document demo-only semantics.

**Consequences of deferral**

Restarts silently erase visible state; horizontal scaling is inconsistent; memory may grow without retention.

**Related:** `TODO-011`, `RISK-009`, `UNK-004`

---

### ADR-004 — Define deployment cardinality and background-job ownership

**Date:** 2026-07-26  
**Status:** Decision needed  
**Owner:** UNKNOWN

**Context**

The server starts an in-process scheduler (`cmd/server/main.go:41-67`). Due schedules are selected without an atomic claim (`internal/schedule/store.go:245-269`). Events and operational stores are local to one process. Compose runs one backend instance.

**Decision:** UNKNOWN whether production must support multiple backend replicas.

**Alternatives requiring evaluation**

- Explicit single-instance application
- Multi-instance API with leader-elected scheduler
- Dedicated worker/scheduler process using Asynq or database leases

**Consequences of deferral**

Multiple replicas can duplicate schedules and partition SSE/state visibility.

**Related:** `TODO-017`, `ADR-001`, `UNK-003`

---

### ADR-005 — Define the authentication and authorization model

**Date:** 2026-07-26  
**Status:** Decision needed  
**Owner:** UNKNOWN

**Context**

Active REST authentication is one optional shared `X-Api-Key`; empty configuration bypasses authentication (`internal/api/server.go:2512-2524`). JWT code exists but is not wired (`internal/auth/auth.go:21-95`). The frontend sends no API-key header (`frontend/src/lib/api.ts:222-229`) and native EventSource cannot attach it. No ownership or role checks exist.

**Decision:** UNKNOWN whether the product is single-admin, shared-key, or multi-user/tenant.

**Alternatives requiring evaluation**

- Explicit single-admin shared-secret mode
- Cookie/session authentication via same-origin/BFF
- JWT/OIDC with roles and tenant ownership
- Separate script API keys and browser sessions

**Consequences of deferral**

Production auth remains incompatible with the dashboard, and resource ownership requirements cannot be designed confidently.

**Related:** `TODO-002`, `TODO-005`, `TODO-006`, `RISK-002`, `UNK-002`

---

## Historical Context (Not Decisions)

- Commit `33a9e3a` introduced the project foundation and historical planning documents.
- Commit `a0eae33` deleted `AGENT.md` and `gotest-agent-master-blueprint.md` while describing a move into `planning/`; no replacements were tracked.
- Commit `4401efe` added the README claim about `planning/`.
- Commit `7b54053` fixed rerun execution.

These entries establish chronology only. They do not confirm current architectural intent.
