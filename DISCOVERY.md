# GoTest Agent — Repository Discovery

> **⚠️ HISTORICAL — This document was written 2026-07-26 before resolving all 22 TODOs.**
> The findings below describe the codebase as discovered. For the current resolved state,
> see [`TODO.md`](.ai/TODO.md) (all 22 items Done) and [`CHANGELOG_AI.md`](.ai/CHANGELOG_AI.md)
> (provenance for each change). Do not use this document as a guide to current gaps — most
> have been fixed.

**Date:** 2026-07-26  
**Method:** Static read-only inspection of tracked files, directory structure, imports, dependency manifests, Docker/Compose config, migrations, and route registrations. Targeted text searches across the full repository.  
**Overall Confidence:** **High** for tracked source structure, schema, route topology, and dependency manifests. **Medium** for runtime behavior and execution correctness because no build, test, lint, browser, or dependency audit could be executed (Bash safety classifier was unavailable during the discovery window).  
**Verification gaps:** `go build ./cmd/server`, `go test ./...`, `npm audit`, `govulncheck`, `pip-audit`, and native `gitleaks` were attempted but blocked. Dependency CVE status is **unknown, not clean**. No build/test pass is claimed.

---

## Executive Summary

GoTest Agent is a self-hosted AI testing platform. It accepts a codebase target, generates Playwright browser tests through LLM calls, executes them, attempts limited self-healing, and exposes auditable evidence through a Next.js dashboard, REST API, and optional MCP toolset.

The repository is ambitious in feature breadth—projects, test plans, approved cases, lists, schedules, monitoring, reviews, releases, alerts, exports, and settings all have registered REST routes and dashboard screens. However, several core execution paths are partial, disconnected, or synthetic. The web run pipeline, MCP, reusable agent, queue worker, sidecar, and approved-case executor each implement different state transitions and execution semantics. Redis, Steel Browser, and LangGraph appear in Compose but are not wired into the primary web-run path. The tree currently imports a missing `internal/planning` package, making it an expected build blocker.

The Compose deployment defaults are explicitly marked as local/demo and list several required production hardening steps. Several defaults (empty API key, known database credentials, all-interface port publication, wildcard CORS) would create material security risk if deployed to a reachable network without change.

---

## 1. Business Overview

**Problem solved:** Manual test writing, maintenance, and auditing for web applications. The product promises automated test generation from specs, execution, self-repair, and evidence collection.

**Target users:** Developers, QA teams, tech leads, and teams preferring self-hosting over SaaS.  
*Evidence:* `frontend/src/lib/docs.ts:35-46` (📌 High).

**Primary workflows:**

| # | Flow | Frontend entry | Backend handler | Confidence |
|---|---|---|---|---|
| 1 | Guided one-off run: spec → generate → execute → audit | `frontend/src/app/create/page.tsx:63-85,104-291` | `internal/api/server.go:1543-1953` | High |
| 2 | Reusable project/plan/approve pathway | `frontend/src/app/projects/page.tsx:71-419` | `internal/api/server.go:273-598` | High |
| 3 | Governed refinement with change proposals | `frontend/src/app/tests/page.tsx:55-267` | `internal/api/server.go:671-935` | High |
| 4 | Test lists and recurring schedules | `frontend/src/app/suites/page.tsx:33-205` | `internal/api/server.go:1045-1200,2637-2741` | High |
| 5 | Run audit: events, video, screenshots, comparison, analysis, report | `frontend/src/app/runs/[id]/page.tsx:38-639` | `internal/api/server.go:2150-2505` | High |
| 6 | Monitoring, risk, releases, reviews, alerts, exports, settings | `frontend/src/components/sidebar.tsx:11-24` | `internal/api/server.go:181-234,2064-3154,3227-3556` | High |
| 7 | IDE/MCP full pipeline | N/A (stdio client) | `internal/mcp/server.go:33-143` | High |

**Inferred product goal (High):** Move QA from ad-hoc generation to governed lifecycle: spec intake → approved reusable cases → lists/schedules → evidence/maintenance → release decisions. The connected project, plan, library, schedule, monitoring, review, and release surfaces support this inference (`frontend/src/components/sidebar.tsx:11-24`; `internal/api/server.go:135-234`).

**Production readiness:** The README explicitly describes the current deployment as optimized for local development/demo (`README.md:182-205`). Rate limiting is not implemented. Several advertised capabilities are partial (see §11).

**Unknowns:** No tracked artifact defines business model, pricing, adoption metrics, SLOs, supported browsers, or formal production-readiness criteria. No `planning/` directory exists despite the README referencing it (`README.md:32`; this path is gitignored per `.gitignore:72-73`).

---

## 2. Technology Stack

| Layer | Technology | Evidence | Confidence |
|---|---|---|---|
| Backend | Go 1.26.4, Go Modules | `go.mod:1-3` | High |
| Frontend | TypeScript, React 19.2.7, Next.js 16.2.7 (App Router), Tailwind CSS 4 | `frontend/package.json:5-29` | High |
| Sidecar | Python 3.12, FastAPI, LangGraph, Uvicorn | `sidecar/Dockerfile:1-7`; `sidecar/requirements.txt:1-6` | High |
| Database | PostgreSQL 16.14, pgx/v5 | `README.md:22-29`; `go.mod:5-14` | High |
| Cache/queue | Redis 7 (configured, unwired at runtime) | `docker-compose.yml:76-87`; `internal/queue/worker.go:27-92` | High |
| Browser engine | Playwright-Go (direct), optional Docker via Playwright container, optional Steel Browser | `internal/agent/playwright_runner.go:32-67`; `internal/runner/docker.go:40-107`; `internal/runner/steel.go:14-69` | High |
| Frontend PM | npm, lockfile v3 | `frontend/package-lock.json:1-29` | High |
| Build | Make, Go toolchain, npm, Docker | `Makefile:3-18` | High |
| Deployment | Docker Compose (six services), no tracked CI/CD/Kubernetes/IaC/cloud manifest found | `docker-compose.yml:1-136` | High |
| MCP tool | Separate Go stdio binary, mcp-go | `cmd/mcp/main.go:13-24` | High |

---

## 3. Repository Structure

### Root
- `Makefile` — Builds both Go binaries, local and Compose lifecycle, smoke test (`Makefile:3-48`)
- `Dockerfile` — Multi-stage Go 1.26 backend image; builds `server`+`mcp`, installs Playwright `@latest` with `|| true` (`Dockerfile:1-18`)
- `docker-compose.yml` — Six-service local stack: backend, frontend, Postgres, Redis, Steel, sidecar (`docker-compose.yml:1-136`)
- `README.md` — Product pitch, stack, API/MCP usage, architecture, security caveats (`README.md:1-206`)
- `.env.example` — Environment reference; contains values not consumed by runtime (`JWT_SECRET`, `GITHUB_WEBHOOK_SECRET`, `BRAINTRUST_API_KEY`)
- `.gitleaks.toml` — Secret scanning config with default rules and allowlist patterns
- `test_playwright.go` — Compile probe only; not a Go test

### `cmd/server/` — HTTP application entry point
`cmd/server/main.go` loads config, connects PostgreSQL or falls back to memory, runs migrations, composes the API server, starts scheduler goroutine, calls `http.ListenAndServe` without configured timeouts (`cmd/server/main.go:15-48`).

### `cmd/mcp/` — Independent stdio MCP process
`cmd/mcp/main.go` constructs Anthropic LLM, Docker runner, agent, and MCP server (`cmd/mcp/main.go:13-24`). Does not share the web server's database or SSE state.

### `internal/`
| Package | Responsibility | Key files |
|---|---|---|
| `api` | Chi router, middleware, monolithic handlers/store composition | `server.go:38-3556` |
| `agent` | `TestRun` model, LLM/Runner/Screenshotter interfaces, reusable state machine, sidecar client | `agent.go:74-345`; `playwright_runner.go:32-186`; `api_runner.go:14-33`; `sidecar.go:14-100` |
| `ai` | Provider-neutral LLM planning client (duplicates partial agent LLM layer) | `client.go:17-146` |
| `auth` | JWT generation/validation (unwired outside tests), API key generation/hashing | `auth.go:21-114` |
| `compare` | Stateless run comparison domain logic | `compare.go:21-103` |
| `config` | Environment-backed configuration struct | `config.go:7-41` |
| `db` | `RunStore` interface, PostgreSQL and memory implementations, embedded migrations | `memory.go:10-17`; `store.go:19-249`; `migrate.go:14-83` |
| `db/migrations` | Eight ordered SQL migration files (`001_init.sql` through `008_change_proposals.sql`) | `001_init.sql:1-76` |
| `evals` | Braintrust evaluation logger (unwired) | `braintrust.go:16-115` |
| `events` | In-memory pub/sub event bus for SSE | `store.go:42-145` |
| `execution` | Runtime artifact context joining events/recordings/visuals | `context.go:13-78` |
| `gitdiff` | Changed-file impact analysis | `diff.go` |
| `intelligence` | Risk scoring, recommendations, release confidence | `engine.go:26-227` |
| `mcp` | MCP server with four tools, local in-memory run map | `server.go:17-143` |
| `metrics` | Pass-rate, hotspot, flakiness, trend computation | `metrics.go:40-140` |
| `notify` | In-memory notification store, webhook delivery (unwired failure trigger) | `store.go:22-99` |
| `project` | Project memory and PostgreSQL repositories | `store.go:33-210` |
| `queue` | Redis/Asynq worker and enqueue (unwired from server) | `worker.go:27-92` |
| `recordings` | In-memory execution recording metadata store | `store.go:21-61` |
| `release` | In-memory release store | `store.go:34-79` |
| `report` | HTML report generation | `html.go:4-81` |
| `reporter` | Playwright report parsing | `reporter.go` |
| `runner` | Docker and Steel browser execution | `docker.go:18-107`; `steel.go:14-69` |
| `schedule` | Memory and PostgreSQL schedule repositories, next-run calculation | `store.go:47-340` |
| `steel` | Steel Browser HTTP client | `client.go:15-125` |
| `vision` | OpenAI vision/analysis client (unwired) | `client.go:17-139` |
| `visual` | In-memory visual artifact/diff store | `store.go:24-69` |
| `webhook` | GitHub push webhook handler | `github.go:46-100` |
| `workflow` | In-memory review and suite stores | `store.go:32-181` |

### `frontend/` — Next.js dashboard
- `src/app/` — App Router pages: `/`, `/create`, `/projects`, `/tests`, `/runs`, `/runs/[id]`, `/runs/[id]/compare`, `/suites`, `/monitoring`, `/risk`, `/releases`, `/reviews`, `/alerts`, `/exports`, `/docs`, `/settings`
- `src/lib/api.ts` — Centralized typed REST client; API URL defaults `http://localhost:8080`; shared fetch sends JSON but no auth headers; per-run SSE (no reconnect), global SSE (reconnect + polling fallback) (`api.ts:1,222-234,386-415,428-691`)
- `src/lib/docs.ts` — Bilingual EN/ID product manual rendered by `/docs`
- `src/lib/utils.ts` — `cn()` helper using `clsx`+`tailwind-merge`
- `src/components/ui/` — Badge, card, chart, section (skeleton/empty state)
- `src/components/console/` — Inspector, timeline, tabs, screenshot strip
- `frontend/package.json` — No test scripts, no test dependencies, no form/state/query/auth libraries
- `frontend/README.md` — Generic create-next-app template (not product-specific)
- `frontend/AGENTS.md`, `frontend/CLAUDE.md` — Contributor notes about Next version

### `sidecar/` — Optional Python multi-agent service
- `main.py` — FastAPI app, in-memory job store, `/agent/run` and `/agent/{job_id}` (unauthenticated)
- `graph.py` — LangGraph planner→writer→critic/rewriter→executor/fixer workflow
- `agents/` — Role-per-module: planner, writer, critic, executor, fixer
- `state.py` — Shared `TypedDict` state contract
- `requirements.txt` — Lower-bounded ranges, no lockfile
- No Python test files found

### `docs/` — Operator documentation
`docs/docker.md` — Docker prerequisites, service topology, commands, health checks, troubleshooting, persistence locations, destructive reset (`docs/docker.md:1-101`).

### Absent
- `planning/` — Referenced by README but gitignored and absent (`.gitignore:72-73`; `README.md:32`)
- `internal/planning/` — Imported by `internal/api/server.go:25` but no tracked source exists
- No ADRs, PRDs, roadmaps, architecture diagrams, or CI/CD workflow files found
- No frontend, Python, or E2E test files found

---

## 4. Architecture

**Primary classification: Modular monolith (High).** One Go process composes all domain modules and serves HTTP, SSE, and scheduler work behind a single Chi router (`internal/api/server.go:38-88`; `cmd/server/main.go:15-48`).

**Supporting classifications:**

| Classification | Fit | Evidence |
|---|---|---|
| Client/server application | Strong | Browser directly calls REST/SSE; no Next.js BFF layer | `frontend/src/lib/api.ts:1,222-234` |
| In-process event-driven | Medium | Mutex-protected event bus broadcasts to SSE subscribers; events are transient, not durable | `internal/events/store.go:42-145` |
| Partial ports-and-adapters | Medium | LLM/Runner/RunStore interfaces exist but web path bypasses the reusable Agent and constructs concrete clients directly | `internal/agent/agent.go:110-134`; `internal/api/server.go:1858-1929` |
| Distributed deployment shell | Low | Compose defines six services but Redis/Steel/LangGraph are not in the primary request execution chain | `docker-compose.yml:1-136` |
| Microservices | Not supported | Only the optional Python sidecar is independently deployed; no domain partitioning | `sidecar/main.py:8-69` |
| CQRS/event sourcing | Not supported | Commands and reads use the same stores; events are transient notifications | `internal/db/store.go:46-249` |

**Execution-path divergence (High):** The main architectural tension is replicated orchestration. The web API directly calls LLM and Playwright in `internal/api/server.go:1858-1953`, while a richer `agent.Agent` with full state machine and fix loop lives at `internal/agent/agent.go:223-345` and is only used by MCP. Redis/Asynq queue exists (`internal/queue/worker.go:27-92`), Steel runner exists (`internal/runner/steel.go:20-70`), sidecar graph exists (`sidecar/graph.py:18-42`), but none are wired into the primary dashboard run path.

---

## 5. Application Flow

### Primary web run execution (the active path)

1. Browser POSTs to `/api/v1/runs` with requirements, target URL, auth notes, mode.  
   `frontend/src/lib/api.ts:222-260`; `frontend/src/app/create/page.tsx:63-85`

2. Chi middleware logs, assigns request ID, optionally checks `X-Api-Key`.  
   `internal/api/server.go:55-60,133-135,2512-2524`

3. `handleCreateRun` derivation of feature map (AI-assisted when `GOTEST_AI_PLANNING=1`), creates `idle` row, emits event, launches goroutine, returns `202`.  
   `internal/api/server.go:1543-1590,1638-1717`

4. Goroutine detaches via `context.Background`, sets `analyzing` state, reads provider/model/key/base URL **only from DB settings** (not `cfg` fallback), generates plan then Playwright action files via LLM. Passes literal `"Web Application"` as analysis context.  
   `internal/api/server.go:1858-1914`

5. Transitions to `running`, constructs local `PlaywrightRunner`, installs Playwright, executes generated JSON actions (`goto`, `fill`, `click`, `scroll`, `wait`). Self-heals individual actions up to 3 retries via LLM.  
   `internal/api/server.go:1916-1929`; `internal/agent/playwright_runner.go:32-186`

6. Unresolved action errors are not converted to failures. Runner returns 1 passed/0 failed except when Playwright setup itself errors.  
   `internal/agent/playwright_runner.go:104-163,180-186`

7. Marked `done`, emits events. Storage update errors are ignored.  
   `internal/api/server.go:1931-1953`

8. Frontend run detail page conditionally opens SSE only if initial fetch shows non-`idle` state; SSE has no reconnect or polling fallback.  
   `frontend/src/app/runs/[id]/page.tsx:38-65`; `frontend/src/lib/api.ts:386-415`

### Project plan → approved case → test list flow

Project creation → feature extraction → draft plan generation → case approval → runs via simulated (default) or Docker (`GOTEST_APPROVED_CASE_RUNNER=docker`). Cases feed lists; lists feed schedules.  
`internal/api/server.go:273-598,1045-1200,1250-1286`

### MCP flow (separate)

`cmd/mcp` constructs `agent.Agent` with full state machine (`idle → analyzing → plan_generated → writing_tests → running ↔ fixing → done|failed`), runs synchronously, returns JSON result. Uses in-memory-only run map.  
`cmd/mcp/main.go:13-24`; `internal/mcp/server.go:74-143`; `internal/agent/agent.go:223-345`

### Schedule flow (partially disconnected)

60-second scheduler loop. Test-list schedules call `startTestListRuns` and execute cases. Non-list schedules and `run-now` only create `idle` rows — execution is never launched. Due selection has no atomic claim.  
`cmd/server/main.go:41-67`; `internal/api/server.go:2637-2741`; `internal/schedule/store.go:245-269`

### Sidecar (experimental/unwired from web)

Advanced Agent can call sidecar, which runs planner→writer→critic→executor/fixer graph. But sidecar executor posts a new REST run and treats the `202` acknowledgement as completion — it does not wait for or consume the created run's result. Web API does not construct `SidecarClient`.  
`internal/agent/agent.go:178-220`; `sidecar/agents/executor.py:9-38`; `cmd/server/main.go:15-46`

### State transitions per path

| Flow | Transitions | Evidence | Confidence |
|---|---|---|---|
| Web create/rerun/webhook | `idle → analyzing → running → done \| failed` | `internal/api/server.go:1543-1590,1858-1953` | High |
| Reusable Agent (MCP only) | `idle → analyzing → plan_generated → writing_tests → running ↔ fixing → done \| failed` | `internal/agent/agent.go:17-26,223-345` | High |
| Approved case, default | `idle → writing_tests → running → done` (simulated, always passes) | `internal/api/server.go:1203-1248` | High |
| Approved case, Docker | `idle → writing_tests → running → done \| failed` | `internal/api/server.go:1250-1286` | High |
| Non-list schedule | `idle` (stuck, no execution) | `internal/api/server.go:2671-2691,2721-2740` | High |
| Sidecar job | `running → completed \| failed` | `sidecar/main.py:33-69` | High |

---

## 6. Database

**Technology:** PostgreSQL 16.14 via pgx/v5 (`go.mod:5-14`). Default Compose DSN: `postgres://postgres:password@postgres:5432/gotest_agent?sslmode=disable`.  
**Storage:** Named volume `pgdata` at `/var/lib/postgresql/data`; app data at `app_data:/data`.  
**Backups:** None found. Only documented operation is destructive volume reset (`docs/docker.md:92-101`).

### Schema (all tables, High confidence)

| Table | Key columns | FKs | Indexes | Evidence |
|---|---|---|---|---|
| `schema_migrations` | `version PK`, `applied_at` | — | — | `internal/db/migrate.go:21-26` |
| `projects` | `id uuid PK`, `name NOT NULL`, `path`, `language`, `framework`, `config jsonb`, `test_type`, `base_url`, `environment`, `spec`, `api_docs`, `auth_type`, `credentials text`, `focus_hints`, `skip_hints`, `feature_map jsonb`, timestamps | — | `created_at DESC`, `test_type` | `001_init.sql:3-11`; `003_projects_intake.sql:1-14` |
| `test_runs` | `id uuid PK`, `project_id uuid`, `state NOT NULL`, `mode`, `requirements`, `code_analysis`, `test_plan jsonb`, `test_files jsonb`, `run_result jsonb`, `screenshots jsonb`, `error_msg`, `duration_ms`, `llm_tokens_used`, `video_*`, `fix_attempts`, `test_case_id uuid`, `test_list_id uuid`, timestamps | `project_id → projects(id)` | `state`, `project_id`, `created_at DESC`, `test_case_id`, `test_list_id` | `001_init.sql:13-36,64-66,68-76`; `002_testsprite_run_metadata.sql:1-9`; `006_run_test_links.sql:1-5` |
| `api_keys` | `id uuid PK`, `key_hash UNIQUE NOT NULL`, `name`, timestamps | — | — | `001_init.sql:38-44` |
| `settings` | `key PK`, `value NOT NULL`, `updated_at` | — | — | `001_init.sql:46-50` |
| `test_plan_drafts` | `id uuid PK`, `project_id uuid`, `status NOT NULL`, `cases jsonb NOT NULL`, timestamps | `project_id → projects(id)` | `project_id` | `004_test_planning_review.sql:1-8,26` |
| `test_cases` | `id uuid PK`, `project_id uuid`, `plan_id uuid`, `title NOT NULL`, `type NOT NULL`, `feature`, `priority NOT NULL`, `steps jsonb`, `assertions jsonb`, `tags jsonb`, `version NOT NULL`, timestamps | `project_id → projects(id)`, `plan_id → test_plan_drafts(id)` | `project_id`, `type` | `004_test_planning_review.sql:10-28` |
| `test_lists` | `id uuid PK`, `name NOT NULL`, `project_id uuid`, `tags jsonb`, `test_case_ids jsonb`, `pinned NOT NULL`, timestamps | `project_id → projects(id)` | `project_id`, `pinned` | `005_test_lists.sql:1-13` |
| `schedules` | `id uuid PK`, `project_id uuid`, `test_list_id uuid`, `name NOT NULL`, `frequency NOT NULL`, `timezone NOT NULL`, `enabled NOT NULL`, `next_run_at NOT NULL`, `last_run_id uuid`, `notify_on_fail`, `cron_expr`, `webhook_url`, timestamps | `project_id → projects(id) ON DELETE SET NULL`, `test_list_id → test_lists(id) ON DELETE CASCADE`, `last_run_id → test_runs(id) ON DELETE SET NULL` | `(enabled,next_run_at)`, `test_list_id`, `project_id` | `007_schedules.sql:1-27` |
| `change_proposals` | `id uuid PK`, `test_case_id uuid NOT NULL`, `status NOT NULL`, `prompt`, `rationale`, `original jsonb`, `proposed jsonb`, `reviewed_at`, `reviewer`, `review_comment`, timestamps | `test_case_id → test_cases(id) ON DELETE CASCADE` | `(test_case_id,created_at DESC)`, `status` | `008_change_proposals.sql:1-17` |

**Application-only (non-FK) relationships:**
- `test_runs.test_case_id` / `test_runs.test_list_id` are indexed but have no foreign keys (`006_run_test_links.sql:1-5`)
- `test_lists.test_case_ids` is a JSON array of UUIDs with no relational constraint to `test_cases` (`005_test_lists.sql:5-6`)

### Migration behavior (High)

- Embedded SQL files, lexically sorted, applied via `pool.Exec` per file
- Each migration and its `schema_migrations` version insert are **separate operations** without a transaction or advisory lock
- Startup logs migration failure but **continues serving** against potentially mismatched schema
- No down migrations, rollback tooling, or checksum validation
- `uuid-ossp` extension required (`001_init.sql:1`)

*Evidence:* `internal/db/migrate.go:14-83`; `cmd/server/main.go:22-34`

### Repository & persistence behavior (High)

| Entity | PostgreSQL path | Memory fallback | Issues |
|---|---|---|---|
| Runs | `internal/db/store.go:46-249` (parameterized) | `internal/db/memory.go:19-80` | Ignores JSON errors, no RowsAffected check on UPDATE, no `rows.Err()` on List; declared columns missing from projections; `project_id` always inserted NULL |
| Projects | `internal/project/store.go:107-172` | `internal/project/store.go:40-97` | No delete; credentials plaintext |
| Settings | `internal/db/settings_store.go:19-72` | N/A (unavailable in memory mode) | `GetAll` misses `rows.Err()`; only `SetMany` uses a transaction |
| Schedules | `internal/schedule/store.go:135-269` | `internal/schedule/store.go:56-133` | `context.Background()`, errors collapsed to nil/false; no update lock; timezone stored but ignored in next-run calc |
| Events | Always memory | N/A | Lost on restart, unbounded growth, nonblocking sends lose events on full channel |
| Recordings | Always memory | N/A | Lost on restart |
| Visuals | Always memory | N/A | Lost on restart |
| Releases | Always memory | N/A | Lost on restart |
| Notifications | Always memory | N/A | Lost on restart; `TriggerFailure` has no production callers |
| Reviews | Always memory | N/A | Lost on restart |
| Suites | Always memory | N/A | Lost on restart |
| Sidecar jobs | Always memory (`sidecar/main.py:11-12`) | N/A | Lost on sidecar restart |

**Planning repository:** Unknown. `internal/api/server.go:25,44,63,67` imports `internal/planning`, but no tracked `internal/planning/**` exists. Confirmed absent source; expected build blocker. Compiler command unavailable. Planning table schemas and tests exist, indicating incomplete extraction (`004_test_planning_review.sql:1-28`; `internal/api/planning_test.go:12-245`).

### Seeds (High)

- Eight default settings inserted with `ON CONFLICT DO NOTHING` (`001_init.sql:52-62`)
- Demo endpoint creates 5 runs, events, 1 schedule, 1 release — ignores insert errors, no idempotency (`internal/api/server.go:3225-3334`)
- No standalone seed files for projects, cases, lists, or API keys

### Prioritized database issues

1. **HIGH/High** — Missing `internal/planning` source: expected build blocker
2. **HIGH/High** — Non-atomic migrations, non-fatal migration failure
3. **HIGH/High** — No-op DELETE run; no retention/artifact cleanup
4. **HIGH/High** — Plaintext credentials and LLM API key in database columns
5. **HIGH/High** — Due schedules not atomically claimed; multi-write flows non-transactional
6. **HIGH/High** — Application-only run-to-case/list and list-member FKs absent
7. **MEDIUM/High** — Volatile stores lost on restart, unbounded growth
8. **MEDIUM/High** — No PostgreSQL integration tests; current tests use memory stores
9. **MEDIUM/High** — Free-text state/type/priority/status fields without CHECK constraints

---

## 7. API

### Route inventory

All registrations at `internal/api/server.go:124-261`.  
**Auth:** `P` = public, `K` = optional `X-Api-Key` (bypassed when `API_KEY` empty), `H` = conditional GitHub HMAC.  
**Validation:** `ID` = alphanumeric/underscore/hyphen 1–64 chars, `J` = JSON decode only, `R` = additional checks, `—` = no meaningful check.

#### Public and special routes

| Method | Path | Auth | Handler | Notes |
|---|---|---|---|---|
| `GET` | `/health` | P | `handleHealth` | Process-only, no DB/Redis probe |
| `GET` | `/videos/{filename}` | P | inline | Temporary video; bypasses authenticated `/videos/*` |
| `POST` | `/api/v1/webhooks/github` | H | `GitHubHandler` | HMAC when secret non-empty; reuses `API_KEY` |
| `ANY` | `/videos/*` | K | file server | Authenticated persistent video serving |

#### `/api/v1` routes (all conditional `X-Api-Key` auth)

**Projects:**
| Method | Path | Handler | Validation |
|---|---|---|---|
| `POST` | `/api/v1/projects` | `handleCreateProject` | J; R: nonblank name |
| `GET` | `/api/v1/projects` | `handleListProjects` | — |
| `GET` | `/api/v1/projects/{id}` | `handleGetProject` | ID; existence |
| `PATCH` | `/api/v1/projects/{id}` | `handleUpdateProject` | ID; J; existence |
| `POST` | `/api/v1/projects/{id}/api-docs` | `handleUploadAPIDocs` | ID; J; existence |
| `POST` | `/api/v1/projects/{id}/parse-api` | `handleParseAPIDocs` | ID; R |
| `POST` | `/api/v1/projects/{id}/extract-features` | `handleExtractProjectFeatures` | ID; existence |
| `POST` | `/api/v1/projects/{id}/test-plan` | `handleGenerateProjectTestPlan` | ID; existence |

**Test plans and cases:**
| Method | Path | Handler | Validation |
|---|---|---|---|
| `GET` | `/api/v1/test-plans/{id}` | `handleGetTestPlan` | ID; existence |
| `PATCH` | `/api/v1/test-plans/{id}/cases/{caseId}` | `handleUpdateTestPlanCase` | J; membership |
| `POST` | `/api/v1/test-plans/{id}/regenerate` | `handleRegenerateTestPlan` | existence |
| `POST` | `/api/v1/test-plans/{id}/approve` | `handleApproveTestPlan` | existence |
| `GET` | `/api/v1/test-cases` | `handleListTestCases` | optional `project_id` |
| `GET` | `/api/v1/test-cases/maintenance` | `handleTestCaseMaintenance` | optional `project_id` |
| `GET` | `/api/v1/test-cases/{id}` | `handleGetTestCase` | ID; existence |
| `PATCH` | `/api/v1/test-cases/{id}` | `handleUpdateTestCase` | J; existence |
| `POST` | `/api/v1/test-cases/{id}/run` | `handleRunTestCase` | ID; existence |
| `POST` | `/api/v1/test-cases/{id}/refine` | `handleRefineTestCase` | ID; J; R: prompt |

**Change proposals:**
| Method | Path | Handler | Validation |
|---|---|---|---|
| `GET` | `/api/v1/test-cases/{id}/proposals` | `handleListTestCaseProposals` | ID |
| `GET` | `/api/v1/change-proposals` | `handleListChangeProposals` | optional `test_case_id` |
| `POST` | `/api/v1/change-proposals/{id}/approve` | `handleApproveChangeProposal` | ID; pending |
| `POST` | `/api/v1/change-proposals/{id}/reject` | `handleRejectChangeProposal` | ID; pending |

**Test lists:**
| Method | Path | Handler | Validation |
|---|---|---|---|
| `POST` | `/api/v1/test-lists` | `handleCreateTestList` | J; R: name + nonempty IDs |
| `GET` | `/api/v1/test-lists` | `handleListTestLists` | optional `project_id` |
| `GET` | `/api/v1/test-lists/{id}` | `handleGetTestList` | ID; existence |
| `GET` | `/api/v1/test-lists/{id}/history` | `handleTestListHistory` | ID; existence |
| `POST` | `/api/v1/test-lists/{id}/run` | `handleRunTestList` | ID; existence |

**Runs (core):**
| Method | Path | Handler | Validation |
|---|---|---|---|
| `POST` | `/api/v1/runs` | `handleCreateRun` | J only |
| `GET` | `/api/v1/runs` | `handleListRuns` | — (hard-coded 50) |
| `GET` | `/api/v1/runs/{id}` | `handleGetRun` | ID; existence |
| `POST` | `/api/v1/runs/{id}/rerun` | `handleRerun` | existence |
| `GET` | `/api/v1/runs/{id}/stream` | `handleSSEStream` | — |
| `GET` | `/api/v1/runs/{id}/events` | `handleGetEvents` | — |
| `GET` | `/api/v1/runs/{id}/api-logs` | `handleGetAPILogs` | ID; existence (placeholder) |
| `GET` | `/api/v1/runs/{id}/report` | `handleReport` | existence |
| `POST` | `/api/v1/runs/{id}/analyze-failure` | `handleAnalyzeFailure` | existence |
| `GET` | `/api/v1/runs/{id}/compare/{otherId}` | `handleCompare` | both exist |
| `GET` | `/api/v1/runs/{id}/recordings` | `handleGetRecordings` | — |
| `GET` | `/api/v1/runs/{id}/visual` | `handleGetVisualArtifacts` | — |
| `GET` | `/api/v1/runs/{id}/video` | `handleGetVideoMetadata` | existence |
| `DELETE` | `/api/v1/runs/{id}` | `handleDeleteRun` | — (returns 204, deletes nothing) |

**Streaming:**
| Method | Path | Handler | Auth |
|---|---|---|---|
| `GET` | `/api/v1/stream` | `handleGlobalStream` | K |

**Schedules:**
| Method | Path | Handler | Validation |
|---|---|---|---|
| `POST` | `/api/v1/schedules` | `handleCreateSchedule` | J |
| `GET` | `/api/v1/schedules` | `handleListSchedules` | — |
| `GET` | `/api/v1/schedules/{id}` | `handleGetSchedule` | existence |
| `PATCH` | `/api/v1/schedules/{id}` | `handleUpdateSchedule` | J map; type assertion panics on wrong types |
| `DELETE` | `/api/v1/schedules/{id}` | `handleDeleteSchedule` | existence |
| `POST` | `/api/v1/schedules/{id}/run-now` | `handleRunNow` | existence (creates idle only) |

**Releases (in-memory):**
| Method | Path | Handler | Validation |
|---|---|---|---|
| `POST` | `/api/v1/releases` | `handleCreateRelease` | J only |
| `GET` | `/api/v1/releases` | `handleListReleases` | — |
| `GET` | `/api/v1/releases/{id}` | `handleGetRelease` | existence |
| `PATCH` | `/api/v1/releases/{id}` | `handleUpdateRelease` | J map; type assertion panics |
| `GET` | `/api/v1/releases/{id}/summary` | `handleReleaseSummary` | existence |
| `GET` | `/api/v1/releases/{id}/confidence` | `handleReleaseConfidence` | existence |
| `GET` | `/api/v1/releases/{id}/risk` | `handleReleaseRisk` | existence |
| `GET` | `/api/v1/releases/{id}/explanation` | `handleReleaseExplanation` | existence |

**Metrics & monitoring:**
| Method | Path | Handler |
|---|---|---|
| `GET` | `/api/v1/monitoring/summary` | `handleMonitoringSummary` |
| `GET` | `/api/v1/metrics/summary` | `handleMetricsSummary` |
| `GET` | `/api/v1/metrics/hotspots` | `handleMetricsHotspots` |
| `GET` | `/api/v1/metrics/flaky` | `handleMetricsFlaky` |
| `GET` | `/api/v1/metrics/trend` | `handleMetricsTrend` |
| `GET` | `/api/v1/metrics/risk` | `handleMetricsRisk` |
| `GET` | `/api/v1/metrics/recommendations` | `handleMetricsRecommendations` |

**Reviews, suites, notifications, exports:**
| Method | Path | Handler |
|---|---|---|
| `POST` | `/api/v1/suite-selection` | `handleSuiteSelection` |
| `POST` | `/api/v1/reviews` | `handleCreateReview` |
| `GET` | `/api/v1/runs/{id}/reviews` | `handleGetRunReviews` |
| `POST` | `/api/v1/reviews/{id}/approve` | `handleApproveReview` |
| `POST` | `/api/v1/reviews/{id}/reject` | `handleRejectReview` |
| `POST` | `/api/v1/reviews/{id}/request-changes` | `handleRequestChangesReview` |
| `GET` | `/api/v1/reviews` | `handleListAllReviews` |
| `POST` | `/api/v1/suites` | `handleCreateSuite` |
| `GET` | `/api/v1/suites` | `handleListSuites` |
| `GET` | `/api/v1/suites/{id}` | `handleGetSuite` |
| `DELETE` | `/api/v1/suites/{id}` | `handleDeleteSuite` |
| `GET` | `/api/v1/notifications` | `handleListNotifications` |
| `POST` | `/api/v1/alert-rules/evaluate` | `handleEvaluateAlertRules` (returns rules, does not persist/deliver) |

**Settings and AI:**
| Method | Path | Handler | Notes |
|---|---|---|---|
| `GET` | `/api/v1/settings` | `handleGetSettings` | Masks `llm_api_key` (presentation only) |
| `PUT` | `/api/v1/settings` | `handleUpdateSettings` | J map; key allowlist; no URL validation |
| `GET` | `/api/v1/ai/providers` | `handleListAIProviders` | — |
| `POST` | `/api/v1/ai/test-provider` | `handleTestAIProvider` | J; 200 on failure; provider routing mismatch with real execution |

**Demo and exports:**
| Method | Path | Handler |
|---|---|---|
| `POST` | `/api/v1/demo/seed` | `handleDemoSeed` |
| `GET` | `/api/v1/runs/{id}/export` | `handleExportRun` |
| `GET` | `/api/v1/runs/{id}/compare/{otherId}/export` | `handleExportCompare` |
| `GET` | `/api/v1/metrics/risk/export` | `handleExportRisk` |
| `GET` | `/api/v1/releases/{id}/confidence/export` | `handleExportConfidence` |

### Authentication, validation, and response conventions (High)

- **Auth:** One shared `X-Api-Key` for all `/api/v1`; completely disabled when empty (`internal/api/server.go:2512-2524`). No user identity, roles, ownership checks, or per-endpoint authorization.
- **JWT:** Implemented but not wired into any production route (`internal/auth/auth.go:21-95`).
- **Frontend incompatibility:** Centralized API client sends no `X-Api-Key` and `EventSource` cannot attach custom headers (`frontend/src/lib/api.ts:222-229,395`). Dashboard breaks if `API_KEY` is configured.
- **CORS:** Hard-coded wildcard `*` (`internal/api/server.go:98-108`). No `CORS_ALLOWED_ORIGINS` config despite README mention (`README.md:201-204`).
- **Webhook secret reuse:** GitHub webhook uses `cfg.APIKey`, not the separately documented `GITHUB_WEBHOOK_SECRET` (`internal/api/server.go:237`).
- **Validation:** Inconsistent. JSON bodies decode without unknown-field rejection, content-type checks, or size limits. Schedule and release PATCH handlers use unchecked `v.(bool)`/`v.(string)` type assertions that panic on wrong types (`internal/api/server.go:2588-2617,2785-2792`).
- **Response formats:** Ad-hoc JSON via `json.NewEncoder`; errors use plain `http.Error`. No common error envelope, status codes, or DTO layer.
- **Notable transport gaps:** `http.ListenAndServe` without timeouts or graceful shutdown (`cmd/server/main.go:44-46`). No body limits. Webhook reads with unbounded `io.ReadAll` (`internal/webhook/github.go:53`).

---

## 8. Frontend

**Framework:** Next.js 16.2.7 App Router, React 19.2.7, TypeScript strict, Tailwind CSS 4, npm.  
**Build:** Standalone Node 22 output, non-root user (`frontend/Dockerfile:1-25`).  
**No frontend test scripts or test dependencies** (`frontend/package.json:5-9,20-29`).

**Route catalog (all confirmed):** `/`, `/create`, `/projects`, `/tests`, `/runs`, `/runs/[id]`, `/runs/[id]/compare`, `/monitoring`, `/risk`, `/releases`, `/reviews`, `/alerts`, `/exports`, `/docs`, `/settings`, `/suites`. Sidebar navigation omits `/projects` (`frontend/src/components/sidebar.tsx:11-24`).

**State:** Local React hooks only. No state management, query cache, form, or validation libraries.

**API client:** Centralized at `frontend/src/lib/api.ts`, defaults to `http://localhost:8080`. JSON fetch wrapper has no auth headers, timeout, retry, or response validation. Several exported functions have no frontend call sites: `updateProject`, `uploadApiDocs`, `updateTestCase`, `getTestListHistory`, `updateSchedule`, `deleteSchedule`, `getReleaseSummary`, `getReleaseConfidence`, `getReleaseExplanation`, `getSuites` (`api.ts:271-280,308-315,355-357,513-553,631-687`).

**Streaming:** Global SSE reconnects with polling fallback (`frontend/src/app/page.tsx:32-66`). Per-run SSE has no reconnect or polling fallback (`frontend/src/lib/api.ts:391-415`). Per-run SSE only opens when initial state ≠ `idle`, creating a race with the backend goroutine.

**Error handling:** Several pages silently catch network errors and render empty state (alerts, monitoring, risk, releases, overview). Mutations frequently lack error feedback (projects, suites, tests, reviews). Guided creation redirects to `/runs` on failure rather than showing the error.

**Accessibility:** Custom tabs lack ARIA roles/keyboard behavior (`components/console/tabs.tsx:24-55`). Screenshot lightbox has no dialog semantics or Escape handling (`components/console/screenshot-strip.tsx:49-60`). Sidebar is fixed 220px with no responsive collapse.

**Unused dependency:** `class-variance-authority` in `frontend/package.json:12` has no source imports. Styling uses `clsx`+`tailwind-merge` via `lib/utils.ts:1-6`.

**Settings concerns:** GitHub webhook URL uses frontend `window.location.origin` instead of `NEXT_PUBLIC_API_URL` (`settings/page.tsx:319-329`). Settings saves optimistically without checking `res.ok`.

---

## 9. Infrastructure

### Deployment (High)

**Only concrete deployment:** Docker Compose, six services on default network, all host-published. README classifies arrangement as local/demo (`README.md:182-205`).  
Quick start: copy `.env.example`, `make up`, `make smoke-test` — but documented quick-start copy enables auth that breaks the unauthenticated dashboard and smoke test.

| Service | Image | Published port | Volume | Health | Notes |
|---|---|---|---|---|---|
| Backend | `golang:1.26-bookworm` (full runtime) | `8080:8080` | — | curl `/health` | `http.ListenAndServe`, no timeouts |
| Frontend | Node 22 Alpine, standalone | `3001:3001` | — | curl | Non-root UID 1001 |
| PostgreSQL | `postgres:16.14-alpine` | `5432:5432` | `pgdata` | `pg_isready` | postgres/password, sslmode=disable |
| Redis | `redis:7-alpine` | `6379:6379` | — | ping | No repo auth/TLS; unwired from server |
| Steel | `steel-browser:latest` | `3010:3000` | — | HTTP | SYS_ADMIN, mutable tag, auth unknown |
| Sidecar | Python 3.12, Uvicorn | `8000:8000` | — | HTTP | No inbound auth; binds 0.0.0.0 |

### Build and supply chain

- **Go:** Two binaries via `go build`; module-locked with `go.sum`
- **Frontend:** `npm ci` with lockfile v3
- **Python:** `requirements.txt` with `>=` ranges, no lockfile
- **Backend image:** Playwright installed at `@latest`, failure suppressed with `|| true` (`Dockerfile:13-14`)
- **Steel:** Mutable `:latest` tag; Redis: `:7-alpine` tag (not digest-pinned)

### Environment variables (selected critical)

| Variable | Status | Evidence |
|---|---|---|
| `API_KEY` | Optional; empty = no auth | `internal/config/config.go:28`; `internal/api/server.go:2512-2524` |
| `DATABASE_URL` | Optional; connection failure → memory fallback | `cmd/server/main.go:22-37` |
| `REDIS_URL` | Optional; loaded but no server wiring found | `internal/config/config.go:30` |
| `ANTHROPIC_API_KEY` | Feature-required for MCP; NOT used by web run (reads DB settings only) | `internal/api/server.go:1868-1894` |
| `LLM_PROVIDER/MODEL/API_KEY/BASE_URL` | Code-consumed but omitted from `.env.example`/Compose | `internal/ai/client.go:57-73` |
| `GITHUB_WEBHOOK_SECRET` | `.env.example` only; runtime uses `API_KEY` | `internal/api/server.go:237` |
| `GOTEST_APPROVED_CASE_RUNNER` | `docker` enables real Docker execution; default simulates | `internal/api/server.go:1203-1247` |
| `JWT_SECRET`, `BRAINTRUST_API_KEY`, `VISION_MODEL`, `ENABLE_VISUAL_REGRESSION`, `ENABLE_ADVANCED_AGENT` | `.env.example` only; no code consumers | `.env.example:4-5,17-19,29-33,39-43` |
| `MAX_FIX_ATTEMPTS`, `DEFAULT_TIMEOUT_SECONDS` | `.env.example` names; runtime hard-codes 3/300 | `internal/config/config.go:37-38` |
| `GOOGLE_API_KEY`, `DEEPSEEK_API_KEY` | README only; no code consumers | `README.md:80-84` |

### Not found (High confidence from full repository sweep)

No tracked CI/CD workflows (`.github/workflows`, GitLab CI, Jenkins), Kubernetes/Helm/Kustomize, Terraform/Pulumi/Ansible, Vercel/Fly/Render/Railway manifests, backup/restore procedures, Prometheus/OpenTelemetry/Grafana integration, log shipping, incident runbooks, or SLO definitions. External systems not tracked in this repository may exist.

### Logging and monitoring

Request ID/access logging/panic recovery via Chi middleware (`internal/api/server.go:55-60`). Lifecycle/scheduler/queue/webhook use `slog`. `/health` returns `{"status":"ok"}` only — no DB/Redis/downstream probes. Sidecar health is likewise process-only. `/monitoring/summary` and `/metrics/*` are product QA analytics, not infrastructure telemetry exporters.

---

## 10. Dependencies

### Go (all 9 direct modules imported; High confidence)

| Module | Purpose | Evidence |
|---|---|---|
| `anthropic-sdk-go` | Anthropic LLM client | `internal/agent/llm_anthropic.go:9-10`; `internal/ai/client.go:13-14` |
| `chi/v5` | HTTP router and middleware | `internal/api/server.go:13-14,55-60` |
| `jwt/v5` | JWT generation/validation | `internal/auth/auth.go:15-16,38-67` (dormant: no production routes wired) |
| `google/uuid` | UUID generation | `internal/api/server.go:35`; `internal/mcp/server.go:12` |
| `asynq` | Redis task queue | `internal/queue/worker.go:14,33-46,81-92` (dormant: no server wiring) |
| `pgx/v5` | PostgreSQL driver | `internal/db/store.go:13,24-44` |
| `mcp-go` | MCP stdio server | `internal/mcp/server.go:13-14,26-71` |
| `playwright-go` | Direct browser automation | `internal/agent/playwright_runner.go:11` |
| `robfig/cron/v3` | Cron expression parsing | `internal/schedule/store.go:11,321-339` |

### Frontend

| Dependency | Status | Evidence |
|---|---|---|
| `next`, `react`, `react-dom`, `lucide-react` | Used | Imported broadly |
| `clsx`, `tailwind-merge` | Used | `frontend/src/lib/utils.ts:1-6` |
| `class-variance-authority` | **Likely unused** | No source imports found; declared at `package.json:12` |
| Dev deps | Explained | Tailwind/PostCSS/TypeScript/ESLint configs |

### Python

All six declared dependencies (`langgraph`, `langchain-anthropic`, `fastapi`, `pydantic`, `uvicorn`, `httpx`) have source imports. `langchain_core` and `typing_extensions` are directly imported but only transitively declared (`sidecar/requirements.txt:1-6`).

### Vulnerability audit status

**All unavailable.** Commands blocked by Bash classifier outage:
- `npm audit` — not run
- `govulncheck` — not run
- `pip-audit` — not run
- Native `gitleaks` — not run

---

## 11. Documentation

| Artifact | Status | Scope |
|---|---|---|
| `README.md` | Complete | Product pitch, stack, API/MCP/schedule/risk/compare workflow, ASCII architecture, production caveats |
| `docs/docker.md` | Complete | Docker prerequisites, services, topology, commands, health, troubleshooting, persistence |
| `frontend/src/lib/docs.ts` | Complete | Embedded bilingual EN/ID product manual, taxonomy, onboarding, concepts, FAQ |
| `frontend/src/app/docs/page.tsx` | Present | Renders embedded docs with language switcher |
| `.env.example` | Partial | Several values have no runtime consumer; missing LLM_PROVIDER/LLM_API_KEY/LLM_BASE_URL |
| `docker-compose.yml` | Executable ref | Six-service topology |
| `Makefile` | Executable ref | Build/start/smoke-test targets |
| `frontend/README.md` | Stale | Generic create-next-app template |
| `frontend/AGENTS.md` / `CLAUDE.md` | Present | Contributor Next version notes |
| ADRs, PRDs, roadmaps, diagrams | **Absent** | Referenced `planning/` is gitignored and absent |

**Known documentation contradictions (High):**
- README says blueprints in `planning/` — gitignored, absent (`README.md:32`; `.gitignore:72-73`)
- Product docs claim Steel Browser execution — web path uses local Playwright (`frontend/src/lib/docs.ts:31-33`; `internal/api/server.go:1922-1929`)
- FAQ claims PHPUnit for Laravel — both generators request Playwright (`frontend/src/lib/docs.ts:614-616`; `internal/agent/llm_anthropic.go:60-85`)
- Embedded docs say Slack/Telegram alerts — notification trigger has no production caller (`frontend/src/lib/docs.ts:393-399`)
- Docker docs cite old model default and port mismatch for Steel (3000 vs 3010 in Compose) (`docs/docker.md:26-33`; `docker-compose.yml:92-93`)

---

## 12. Current Progress

### Substantive/complete (High)

- Broad REST route surface: projects, plans, cases, proposals, lists, runs, streaming, schedules, releases, metrics, intelligence, reviews, suites, settings, exports
- PostgreSQL-backed core entities (runs, projects, settings, schedules) with memory fallback
- Reusable `agent.Agent` state machine with full fix loop (MCP only)
- Dashboard with 15+ routes, SSE control room, embedded docs
- Embedded sorted SQL migrations
- MCP toolset (run_tests, analyze_project, generate_test_plan, get_run_status)
- Health checks and restart policies for all Compose services

### Partial

| Module | Status | Evidence |
|---|---|---|
| Web run pipeline | No codebase analysis, literal `"Web Application"` context, no outer fix loop | `internal/api/server.go:1858-1953` |
| Playwright runner | Unresolved action errors false-pass; always returns 1 passed/0 failed | `internal/agent/playwright_runner.go:104-186` |
| Approved cases, default | Sleeps through steps, synthesizes assertions as passing | `internal/api/server.go:1203-1247` |
| Non-list schedules | Create idle runs without execution; `run-now` same defect | `internal/api/server.go:2671-2741` |
| Alerts | Notification delivery code exists; `TriggerFailure` has no production caller | `internal/notify/store.go:58-99` |
| Releases/visuals/recordings | Always memory; lost on restart | `internal/api/server.go:78-85` |
| Plan/scripts persistence | Generated plan and files are local variables, not stored on run | `internal/api/server.go:1896-1914` |

### Experimental / unwired

| Module | Evidence |
|---|---|
| `APIRunner` | Explicitly calls itself mock; marks everything passed | `internal/agent/api_runner.go:14-33` |
| Advanced mode (dashboard) | UI option present; `executeRealRun` never branches on `run.Mode` | `frontend/src/app/create/page.tsx:283-288`; `internal/api/server.go:1858-1953` |
| Sidecar + LangGraph | Graph exists; sidecar executor treats `202` as completion; server never constructs `SidecarClient` | `sidecar/graph.py:18-42`; `sidecar/agents/executor.py:9-38`; `cmd/server/main.go:15-46` |
| Redis/Asynq queue | Full implementation; no enqueue/worker startup in either command | `internal/queue/worker.go:27-92` |
| Steel Browser | Client and runner exist; no construction in server | `internal/steel/client.go:30-49`; `internal/runner/steel.go:14-69` |
| Braintrust evals | Logger exists; no instantiation | `internal/evals/braintrust.go:35-43,86-115` |
| Vision client | Exists; no wiring | `internal/vision/client.go:17-139` |
| JWT auth | Full implementation; no production route wired | `internal/auth/auth.go:21-95` |
| `api_keys` table | Schema exists; no repository consumer | `001_init.sql:38-44` |

### Dead / stub

- Run DELETE returns 204, deletes nothing (`internal/api/server.go:2508-2510`)
- `simulateMockRun` — rich mock with no caller (`internal/api/server.go:1955-2049`)
- API logs endpoint — returns empty placeholder (`internal/api/server.go:2275-2295`)

---

## 13. Risks

### A. Security findings

**All confirmed with exact citations. Prerequisites stated for each.**

#### Confirmed vulnerabilities in default Compose configuration

**A1. HIGH — Backend API and webhook fail open when `API_KEY` is unset.**  
Default Compose publishes 8080 with empty `API_KEY`. All `/api/v1` routes are publicly accessible. GitHub webhook is unsigned. Wildcard CORS.  
`docker-compose.yml:7-18`; `internal/config/config.go:24-30`; `internal/api/server.go:98-108,133-235,237-255,2512-2524`; `internal/webhook/github.go:59-66`.  
Mitigation: configured non-empty `API_KEY`.  
Confidence: 10/10.

**A2. HIGH — Unauthenticated published sidecar as confused deputy.**  
Sidecar published on 8000, binds 0.0.0.0, accepts unauthenticated `/agent/run` and `/agent/{job_id}`. Forwards `GOTEST_API_KEY` to backend, creating a confused-deputy path.  
`docker-compose.yml:112-122`; `sidecar/main.py:28-46`; `sidecar/agents/executor.py:5-21`.  
Mitigation: remove host mapping or private network; authenticate inbound sidecar.  
Confidence: 10/10.

**A3. HIGH — Host-published PostgreSQL with known password and disabled TLS.**  
`postgres/password`, port 5432 published, `sslmode=disable`. Database holds LLM keys, project credentials, and all application data.  
`docker-compose.yml:10-12,57-67`.  
Mitigation: remove host port; unique generated password; TLS; least-privilege roles.  
Confidence: 10/10.

**A4. HIGH — Stored LLM API key exfiltration via configurable base URL.**  
Settings endpoint permits `llm_base_url` change; stored `llm_api_key` is sent to new URL via `Authorization: Bearer`. No admin role separation.  
`internal/api/server.go:3423-3447,1868-1896`; `internal/agent/llm_openai.go:21-38,74-84`.  
Prerequisite: DB settings, stored key, API access (unauthenticated in default; restricted to shared-key holders when `API_KEY` is set).  
Mitigation: bind credentials to immutable approved origins; admin-only settings; validate destinations.  
Confidence: 10/10.

**A5. HIGH — Docker runner code injection via unescaped project URL.**  
Approved cases with `GOTEST_APPROVED_CASE_RUNNER=docker` interpolate `projectURL` into single-quoted TypeScript without escaping. `--network host` provides full host access. MCP always uses `DockerRunner`.  
`internal/api/server.go:1150-1169,1214-1259`; `internal/runner/docker.go:49-65,70-79`; `cmd/mcp/main.go:13-20`.  
Prerequisite: Docker runner explicitly enabled (`GOTEST_APPROVED_CASE_RUNNER=docker`), or MCP access.  
Confidence: 9/10.

**A6. HIGH — Unbounded resource consumption in default deployment.**  
No request timeouts, body limits, concurrency cap, or per-user quotas. Each run launches an unguarded goroutine.  
`cmd/server/main.go:44-46`; `internal/api/server.go:1543-1585`; `internal/webhook/github.go:53`.  
Confidence: 10/10.

**A7. MEDIUM — Public temporary video route bypasses API authentication.**  
`/videos/{filename}` registered before auth middleware. Separate authenticated `/videos/*` file server protects a different path.  
`internal/api/server.go:127-131,257-261`.  
Prerequisite: knowledge of generated video filename.  
Confidence: 9/10.

#### Design risks (exploitability tied to deployment context)

**D1. HIGH risk — Unrestricted browser egress from generated Playwright actions.**  
LLM-generated `goto` actions execute arbitrary URLs without network policy validation. Approved-case path uses project `BaseURL` directly.  
`internal/agent/playwright_runner.go:77-102`; `internal/api/server.go:1295-1368`.  
Confidence: 8.5/10.

**D2. MEDIUM — Shared security domain with plaintext credential fields.**  
All projects/runs visible to any API-key holder. Credential fields returned by JSON endpoints. No tenant model.  
`internal/project/store.go:139-155`; `internal/db/store.go:94-192`.  
Severity depends on single-admin vs multi-user intent.

**D3. MEDIUM — Redis and Steel unnecessarily host-published; Steel receives SYS_ADMIN.**  
Redis auth/TLS not configured in repository. Steel auth defaults unknown, mutable `:latest` tag, `SYS_ADMIN` capability. Backend and sidecar run as image-default root.  
`docker-compose.yml:76-104`; `Dockerfile:9-18`; `sidecar/Dockerfile:1-7`.

**D4. MEDIUM — Non-reproducible build and supply-chain drift.**  
Python unbounded versions, Steel `:latest`, Playwright `@latest` + `|| true`, Redis version-only tag, no image digests.  
`sidecar/requirements.txt:1-6`; `docker-compose.yml:91`; `Dockerfile:13-14`.

#### Verified security controls

- Parameterized pgx queries (SQL reviewed in `internal/db/store.go`, `settings_store.go`, `project/store.go`, `schedule/store.go`)
- HTML reports use `html/template` (`internal/report/html.go:4-13,72-81`)
- No `dangerouslySetInnerHTML` except static theme script (`frontend/src/app/layout.tsx:15-20`)
- Route ID allowlist (`internal/api/server.go:111-122`)
- HMAC-SHA256 constant-time webhook verification when secret configured (`internal/webhook/github.go:59-66,91-100`)
- `crypto/rand` API key generation (`internal/auth/auth.go:99-113`)
- Settings API key masking and key-name allowlist (`internal/api/server.go:3405-3447`)
- `.env`/key/log/data gitignored (`.gitignore:8-12,37-45,63-65`)
- Gitleaks defaults enabled (`.gitleaks.toml:6-8`)
- Frontend non-root user

### B. Non-security risks

| Risk | Severity | Confidence | Evidence |
|---|---|---|---|
| Missing `internal/planning` — expected build blocker | HIGH | High | `internal/api/server.go:25,44,63,67` |
| Non-atomic migrations, non-fatal failure | HIGH | High | `internal/db/migrate.go:45-69`; `cmd/server/main.go:28-34` |
| Execute path divergence (5+ orchestration implementations) | HIGH | High | `internal/api/server.go:1858-1953` vs `internal/agent/agent.go:223-345` vs `internal/queue/worker.go:27-92` vs `sidecar/graph.py:18-42` |
| Volatile stores lost on restart (events, artifacts, releases, notifications, reviews, suites) | MEDIUM | High | `internal/api/server.go:71-85` |
| Schedule duplication (no atomic claim) | HIGH | High | `internal/schedule/store.go:245-269`; `internal/api/server.go:2694-2741` |
| Non-list schedules stall at idle | HIGH | High | `internal/api/server.go:2671-2741` |
| Playwright false-pass on unresolved errors | HIGH | High | `internal/agent/playwright_runner.go:104-186` |
| Flaky SSE (per-run no reconnect, race with idle state) | MEDIUM | High | `frontend/src/lib/api.ts:391-415`; `frontend/src/app/runs/[id]/page.tsx:48-64` |
| Frontend silent error states (network failure → empty) | MEDIUM | High | `frontend/src/app/alerts/page.tsx:12-16`; `frontend/src/app/monitoring/page.tsx:14-22` |
| No frontend, Python, or E2E tests | MEDIUM | High | `frontend/package.json:5-9,20-29`; sidecar directory sweep |
| No CI/CD, IaC, backup, production telemetry found | MEDIUM | High | Full repository sweep |

---

## 14. Unknown Areas

1. Is `internal/planning` accidentally omitted from the `a0eae33` commit, generated elsewhere, or expected from a private module?
2. Which execution path is canonical: web API, reusable Agent, Asynq worker, or sidecar?
3. Is the supported browser backend: local Playwright, Docker, or Steel?
4. Should non-list schedules execute or only create idle rows?
5. Is production always behind firewall, TLS reverse proxy, and authenticated ingress not tracked here?
6. Is `API_KEY` enforced by deployment automation despite the repository default being empty?
7. Is the application strictly single-administrator or multi-user? This determines whether shared authorization domain and visible credentials are acceptable.
8. Are credential fields guaranteed to contain secret-manager references, or can they hold live credentials?
9. What are the upstream Steel authentication defaults when no auth variable is passed?
10. Does deployment infrastructure provide backup, encryption, and rotation not represented in the repository?
11. What are the actual CVE results from `npm audit`, `govulncheck`, and `pip-audit` once command execution is restored?
12. Should releases, reviews, suites, recordings, visuals, and events survive server restart?
13. Is `class-variance-authority` intentionally kept for future component work, or is it removable dead weight?
14. Is advanced mode (sidecar/LangGraph) intended for the dashboard, MCP only, or experimental only?
15. Should the playground-style `GOTEST_APPROVED_CASE_RUNNER=docker` be the default, or is simulated success an intentional demo mode?

---

## Questions for Owners

1. Is `internal/planning` a missing tracked file or a planned package from an unmerged branch?
2. What is the intended production topology beyond local `docker-compose up`?
3. Should `API_KEY` be mandatory outside development, and how should the dashboard authenticate?
4. Are `credentials` and `llm_api_key` expected to contain production secrets? If yes, encryption at rest is required.
5. Which browser backend should the web path use, and should it match the MCP path?
6. Should Redis/Asynq queue replace in-process goroutines for durability?
7. What is the required backup RPO/RTO for PostgreSQL and `/data`?
8. Should multiple backend replicas be supported? Current scheduling assumes singleton.

---

## Overall Confidence

| Area | Confidence |
|---|---|
| Repository structure, directory responsibilities | High |
| Technology stack and versions | High |
| Registered route topology | High |
| Database schema, migrations, seeds | High |
| Dependency manifests and usage | High |
| Deployment topology (Compose) | High |
| Architecture style (modular monolith) | High |
| Implementation status per module | High (code), Medium (runtime) |
| Security configuration state | High (static), Medium (deployment unknowns) |
| Current build and test status | **Low** — no commands could be executed |
| Dependency vulnerability status | **Unknown** — scanners unavailable |
| Production intent and scale | Low — no tracked artifacts define it |

Build, test, and dependency-audit commands could not be executed because the Bash safety classifier was unavailable throughout the discovery window. The structurally confirmed missing `internal/planning` package should be treated as a blocker until a clean build is observed. All code-behavior conclusions are static and should be verified with runtime exercise before operational decisions.
