# Technical Debt

**Owner:** Engineering  
**Authoritative sources:** Static source inspection against verified audit/discovery evidence  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-29  
**Confidence:** High for source-verified items; runtime impact partly UNKNOWN

## Debt Severity

| Severity | Definition |
|---|---|
| **CRITICAL** | Blocks safe operation or compilation |
| **HIGH** | Actively degrades correctness, security, or reliability |
| **MEDIUM** | Impedes maintenance, testing, or future changes |
| **LOW** | Cosmetic, stylistic, or speculative cleanup |

## Dead Code

| ID | Location | Description | Severity | Evidence |
|---|---|---|---|---|
| DC-1 | `internal/api/server.go` | ~~`simulateMockRun`~~ ✅ Resolved: function removed from codebase (grep 2026-07-27 finds no definition or caller) | ✅ Resolved | — |
| DC-2 | `internal/api/handlers_runs.go` | ~~`handleDeleteRun` returns 204 without deleting~~ ✅ Resolved: now calls `s.store.DeleteRun` (implemented in both MemoryStore and DBStore) | ✅ Resolved | `internal/db/memory.go:84`, `internal/db/store.go:47` |
| DC-3 | ~~`handleGetAPILogs` returns an explicit empty placeholder~~ ✅ Resolved (2026-07-28): dead endpoint removed — handler and route `/runs/{id}/api-logs` deleted; grep confirmed zero frontend or backend consumers | ✅ Resolved | — |
| DC-4 | `internal/auth/auth.go` | ~~JWT with zero production usage~~ ✅ Resolved: `api.NewServer` constructs `auth.New`; JWT cookie auth used by `handleLogin`/`apiKeyAuth` | ✅ Resolved | `internal/api/server.go` (jwtAuth field), `internal/api/handlers_auth.go` |
| DC-5 | `frontend/src/lib/api.ts` | ~10 exported functions with no frontend call sites (updateProject, uploadApiDocs, updateTestCase, etc.) | LOW | See `DEPENDENCIES.md`; removal requires confirm no planned usage |

## Unwired Implementations

These packages have implementation code but no connection from the running server process:

| ID | Package | Wiring gap | Severity | Evidence |
|---|---|---|---|---|
| UW-1 | `internal/queue` | ✅ Resolved (2026-07-27): `RunWorker`/`runs:execute` wired opt-in via `QUEUE_ENABLED=true` in `cmd/server`; legacy `TypeTestRun` job remains unwired | ✅ Resolved | `internal/queue/run_worker.go`; `cmd/server/main.go` |
| UW-2 | `internal/runner/steel.go`, `internal/steel/` | ✅ Resolved (2026-07-29): `SteelRunner` in `internal/agent/steel_runner.go` connects Playwright via CDP to Steel sessions; API endpoints wired | ✅ Resolved | `internal/agent/steel_runner.go` |
| UW-3 | `internal/agent/sidecar.go` | `SidecarClient` exists; no server construction | MEDIUM | `cmd/server/main.go` does not construct `SidecarClient` |
| UW-4 | `internal/vision/client.go` | Vision analysis client exists; no wiring | LOW | No construction in either executable |
| UW-5 | `internal/evals/braintrust.go` | Evaluation logger exists; no instantiation | LOW | `internal/evals/braintrust.go:35-43,86-115` |
| UW-6 | `internal/notify/store.go` | ✅ Resolved (2026-07-27): `Server.StartFailureNotifier` (global event subscription) calls `TriggerFailure` on `run_failed`; wired in `cmd/server` | ✅ Resolved | `internal/api/failure_notifier.go` |

## Large Files and Functions

| ID | Location | Size | Severity | Notes |
|---|---|---|---|---|
| LF-1 | `internal/api/server.go` | ✅ Resolved (2026-07-27): split into 10 domain handler files (`handlers_*.go`); server.go now ~335 lines (Server struct, NewServer, routes, middleware, helpers) | ✅ Resolved | 143 functions preserved; pure mechanical move |
| LF-2 | `internal/api/server.go` `launchRun` | ~35 lines | LOW | Agent constructor in server: reads LLM settings, constructs Agent, delegates to `Agent.Launch` |
| LF-3 | `internal/api/server.go` `handleApproveTestPlan` | ~30 lines | LOW | Multi-step approval (case creation then draft update) without a transaction or service boundary |

## Duplicate Logic

| ID | Location | Description | Severity |
|---|---|---|---|
| DL-1 | `internal/api/server.go` vs `internal/agent/agent.go` | `launchRun` now delegates to Agent; `executeRealRun` removed | ✅ Resolved (TODO-007 Phase 2) |
| DL-2 | `internal/agent/llm_*.go` vs `internal/ai/client.go` | ~~Two independent LLM client layers~~ ✅ Resolved (2026-07-28, ADR-006 Steps A–D): `internal/ai` is the single transport (text + vision); `agent.LLM` implemented by `promptLLM` adapter composing shared prompt builders (`llm_prompts.go`) with `ai.Client`; duplicated SDK/HTTP transports deleted (~250 lines); routing parity + prompt-contract + transport tests guard regressions | ✅ Resolved |
| DL-3 | ~~Two near-identical draft-plan creation blocks~~ ✅ Resolved (2026-07-28): shared `createDraftPlanResponse` helper in `handlers_planning.go`; both `handleGenerateProjectTestPlan` and `handleParseAPIDocs` delegate | ✅ Resolved | — |
| DL-4 | ~~Non-list `run-now` and due-schedule branches duplicate run creation~~ ✅ Resolved (2026-07-28): shared `startScheduleRun` helper in `handlers_schedules.go`; behavioral differences (last-run status, event message) parameterized; caller launches so run-now can snapshot response first | ✅ Resolved | — |

## Poor Naming and Consistency

| ID | Issue | Location | Severity |
|---|---|---|---|
| NC-1 | Repository interface naming inconsistency: `Store` (planning), `Store` (project), `Repository` (schedule), `RunStore` (db) | `internal/project/store.go`, `internal/schedule/store.go`, `internal/db/memory.go`, `internal/planning/types.go` | LOW |
| NC-2 | Schema column `path` vs code field `BaseURL` on project | `internal/db/migrations/003_projects_intake.sql:3`; `internal/project/store.go:19-20` | LOW |
| NC-3 | Indonesian-language comments and messages throughout Go source (event messages, variable names) are a product-level bilingual choice, not a technical-debt finding | Multiple locations | N/A |

## Unused Configuration and Dependencies

| ID | Issue | Severity | Evidence |
|---|---|---|---|
| UC-1 | `.env.example` declares `BRAINTRUST_API_KEY`, `VISION_MODEL`, `GOOGLE_API_KEY`, `DEEPSEEK_API_KEY` with no runtime consumer (updated 2026-07-28: `JWT_SECRET` and `GITHUB_WEBHOOK_SECRET` removed from this item — both now consumed: `internal/api/server.go:69,288`) | LOW | grep 2026-07-28: zero Go references to the four remaining vars |
| UC-2 | ~~`.env.example` declares `MAX_FIX_ATTEMPTS=3` and `DEFAULT_TIMEOUT_SECONDS=300` but runtime values are hard-coded~~ ✅ Resolved (2026-07-28): `config.Load` now reads both via `getEnvInt` with unchanged defaults; covered by `internal/config/config_test.go` | ✅ Resolved | `internal/config/config.go`; consumers: `cmd/mcp/main.go`, `internal/api/handlers_planning.go` |
| UC-3 | ~~`STEEL_MAX_SESSIONS` in `.env.example` is unused; config hard-codes `10`~~ ✅ Resolved (2026-07-28): read via `getEnvInt("STEEL_MAX_SESSIONS", 10)` | ✅ Resolved | `internal/config/config.go` |
| UC-4 | ~~Frontend `class-variance-authority` declared but no source import found~~ ✅ Resolved (verified 2026-07-28): dependency no longer present in `frontend/package.json`; all remaining runtime deps (`clsx`, `lucide-react`, `tailwind-merge`) have source imports | ✅ Resolved | grep 2026-07-28: zero references in `frontend/` |
| UC-5 | ~~Redis defined in Compose, Asynq code exists, but the server does not enqueue or consume jobs~~ ✅ Resolved (2026-07-27, with UW-1): `QUEUE_ENABLED=true` wires `RunWorker` + enqueuer in `cmd/server`; Redis remains optional | ✅ Resolved | `cmd/server/main.go`; `internal/queue/run_worker.go` |

## Obsolete Abstractions

None identified at this point. The established patterns (`Store` + `MemoryStore` + `DBStore`, embedded migrations, interface boundaries at external edges) are fit for purpose. The codebase's duplication is in orchestration, not in abstraction layering.

## Complex Logic (high cognitive load)

| ID | Location | Issue | Severity |
|---|---|---|---|
| CL-1 | `internal/api/handlers_*.go` | Handlers mix transport decode, validation, business logic, persistence calls, event emission, and response encoding in one function (mitigated by 2026-07-27 domain split; per-handler structure unchanged) | MEDIUM |
| CL-2 | `internal/api/handlers_schedules.go` (`handleUpdateSchedule`) | ✅ Resolved: PATCH now uses checked `safeBool`/`safeString` helpers (verified 2026-07-27) | ✅ Resolved |
| CL-3 | `internal/api/handlers_releases.go` (`handleUpdateRelease`) | ✅ Resolved: same checked-helper pattern (verified 2026-07-27) | ✅ Resolved |
| CL-4 | `internal/agent/playwright_runner.go:104-163` | Nested try/heal/retry loop with DOM snapshot, screenshot, and LLM healing — correct behavior depends on error handling at each nesting level | MEDIUM |
| CL-5 | `internal/schedule/store.go` | ~~`CalcNextRun` ignores the stored timezone~~ ✅ Resolved (2026-07-27, commit 5b9ae48): `CalcNextRunInTZ` evaluates cron in the schedule's IANA timezone; all call sites pass `Timezone`; 3 TZ tests added | ✅ Resolved |

## Prioritized Cleanup Order

1. **CRITICAL:** None remaining after `TODO-001` resolution (planning package restored).
2. **HIGH:** Remove `handleDeleteRun` dead body, consolidate execution paths, unify provider routing, fix non-list schedule execution, fix PATCH type assertions.
3. **MEDIUM:** Wire or remove dormant integrations (after ADR decisions), add missing tests, add server timeouts, fix synthetic test outcomes.
4. **LOW:** Remove `simulateMockRun` dead code, prune unused API exports, remove unused frontend dependency, rename inconsistent interfaces only during a dedicated readability pass.
