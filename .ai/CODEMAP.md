# Code Map

**Owner:** Engineering (individual module owners UNKNOWN)  
**Authoritative sources:** Tracked directory structure and source symbols  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-26  
**Verification performed:** Static file/symbol inspection  
**Confidence:** High for paths and symbols; runtime ownership and team ownership are UNKNOWN

## Entry Points

| Path | Entry point | Responsibility |
|---|---|---|
| `cmd/server/main.go` | `main()` | Load config, choose PG/memory run store, run migrations, compose API, start scheduler, listen HTTP (`:15-49`) |
| `cmd/mcp/main.go` | `main()` | Compose Anthropic LLM, Docker runner, reusable Agent, and stdio MCP server (`:13-25`) |
| `frontend/src/app/layout.tsx` | `RootLayout` | Dashboard shell, global CSS/theme, sidebar/header (`:10-33`) |
| `sidecar/main.py` | FastAPI `app` | Health, job start, job polling, background graph execution (`:8-69`) |

## Root Files

| Path | Purpose / caution |
|---|---|
| `README.md` | Product/operator overview; not implementation authority; `:32` has stale `planning/` reference |
| `Makefile` | Go build/dev and Compose commands (`:3-48`) |
| `Dockerfile` | Builds server/MCP; runtime installs Playwright (`:1-18`) |
| `docker-compose.yml` | Six-service local/demo topology (`:1-136`) |
| `.env.example` | Environment template; contains documented/runtime mismatches |
| `.gitleaks.toml` | Secret-scanning policy |
| `.ai/` | Internal engineering knowledge base |
| `test_playwright.go` | Playwright option compile probe; not a `_test.go` test |

## `internal/` Packages

| Package | Key symbols/files | Responsibility and status |
|---|---|---|
| `api` | `Server`, `NewServer`, routes, handlers, `launchRun` | HTTP composition/transport; delegates execution to `agent.Agent` via `launchRun` |
| `agent` | `TestRun`, `TestPlan`, `LLM`, `Runner`, `Agent`, `Execute` | Core execution models/interfaces, reusable Agent, LLM clients, direct Playwright, sidecar client |
| `ai` | `Config`, `New`, provider client | Separate provider-neutral planning client; overlaps agent LLM clients |
| `auth` | `Auth`, JWT middleware, `GenerateAPIKey` | JWT code dormant in production; API-key generation used as utility |
| `compare` | `Compare` | Run comparison |
| `config` | `Config`, `Load` | Environment values/defaults |
| `db` | `RunStore`, `Store`, `MemoryStore`, `SettingsStore`, `RunMigrations` | Core persistence and migration runner |
| `db/migrations` | `001_init.sql`–`008_change_proposals.sql` | Physical PostgreSQL schema authority |
| `evals` | `BraintrustLogger` | Evaluation logger; primary server wiring absent |
| `events` | `Store`, `Emit`, `Subscribe`, `SubscribeAll` | Memory event history/pub-sub for SSE |
| `execution` | `Context` | Joins event/recording/visual evidence production |
| `gitdiff` | Diff/impact helpers | Changed-file analysis |
| `intelligence` | Engine/risk/confidence helpers | Risk, suite selection, release confidence/explanation |
| `mcp` | `Server`, tool handlers | Four stdio tools and local run status map |
| `metrics` | Summary/hotspot/flaky/trend functions | QA analytics over runs |
| `notify` | `Store`, `TriggerFailure` | Memory notifications and webhook delivery; trigger wiring absent |
| `project` | `Project`, `Store`, `MemoryStore`, `DBStore` | Project domain/persistence |
| `queue` | `Worker`, `EnqueueRun` | Asynq path; no executable wiring found |
| `recordings` | `Store` | Memory recording metadata |
| `release` | `Release`, `Store` | Memory releases |
| `report` | `GenerateHTML` | HTML report generation using `html/template` |
| `reporter` | Playwright report parser | Emits step-level execution evidence |
| `runner` | `DockerRunner`, `SteelRunner` | External/sandbox browser execution alternatives |
| `schedule` | `Schedule`, `Repository`, `Store`, `DBStore`, `CalcNextRun` | Schedule storage/calculation |
| `steel` | `Client` | Steel service HTTP API |
| `vision` | `Client` | OpenAI-compatible visual analysis; primary wiring absent |
| `visual` | `Store`, comparison helpers | Memory visual artifacts/diffs |
| `webhook` | `GitHubHandler` | HMAC validation and async push callback |
| `workflow` | `ReviewStore`, `SuiteStore` | Memory review/suite workflows |

### `internal/planning/` — Reusable test-design lifecycle

| Key symbols/files | Responsibility and status |
|---|---|
| `Store` interface, `DraftPlan`, `DraftCase`, `TestCase`, `TestList`, `ChangeProposal` | Domain types and persistence boundary in `types.go` |
| `MemoryStore` (`memory.go`) | Concurrency-safe in-memory implementation; used by tests |
| `DBStore` (`db.go`) | PostgreSQL-backed implementation for migrations 004/005/008 |

**Reconstructed 2026-07-26** from surviving lifecycle test, server call sites, frontend contracts, and migration definitions. `go vet` and `go build ./...` passed; formatting and regression test pending infrastructure availability. See `TODO-001` (in progress).

## Frontend Map

### Routes/pages

| Route | File | Responsibility |
|---|---|---|
| `/` | `frontend/src/app/page.tsx` | Control room/onboarding/global SSE |
| `/create` | `frontend/src/app/create/page.tsx` | Guided run intake |
| `/projects` | `frontend/src/app/projects/page.tsx` | Project/feature/plan approval workflow |
| `/tests` | `frontend/src/app/tests/page.tsx` | Approved-case library/refinement |
| `/runs` | `frontend/src/app/runs/page.tsx` | Run list/filter/inspector |
| `/runs/[id]` | `frontend/src/app/runs/[id]/page.tsx` | Run audit/live events/artifacts/actions |
| `/runs/[id]/compare` | `frontend/src/app/runs/[id]/compare/page.tsx` | Run comparison |
| `/suites` | `frontend/src/app/suites/page.tsx` | Test lists and schedules |
| `/monitoring` | `frontend/src/app/monitoring/page.tsx` | Health/list summary |
| `/risk` | `frontend/src/app/risk/page.tsx` | Risk/recommendations |
| `/releases` | `frontend/src/app/releases/page.tsx` | Release list (creation API-only) |
| `/reviews` | `frontend/src/app/reviews/page.tsx` | Human review and proposal queues |
| `/alerts` | `frontend/src/app/alerts/page.tsx` | Notification history |
| `/exports` | `frontend/src/app/exports/page.tsx` | JSON exports |
| `/docs` | `frontend/src/app/docs/page.tsx` | Bilingual embedded docs |
| `/settings` | `frontend/src/app/settings/page.tsx` | LLM/provider/execution settings |

### Shared frontend code

| Path | Responsibility |
|---|---|
| `frontend/src/lib/api.ts` | Domain/client types, REST functions, SSE subscriptions; canonical browser transport boundary |
| `frontend/src/lib/docs.ts` | Embedded bilingual user docs; product language, not implementation authority |
| `frontend/src/lib/utils.ts` | `cn` using `clsx`/`tailwind-merge` |
| `frontend/src/components/sidebar.tsx` | Navigation shell (does not list `/projects`) |
| `frontend/src/components/ui/` | Badge/card/chart/section/skeleton/empty-state primitives |
| `frontend/src/components/console/` | Run inspector, timeline, tabs, screenshots |

## Sidecar Map

| Path | Responsibility |
|---|---|
| `sidecar/main.py` | FastAPI endpoints and memory job registry |
| `sidecar/graph.py` | LangGraph topology |
| `sidecar/state.py` | Shared graph TypedDict |
| `sidecar/agents/planner.py` | Test plan generation |
| `sidecar/agents/writer.py` | Test file generation |
| `sidecar/agents/critic.py` | Critique/rewrite decision |
| `sidecar/agents/executor.py` | Calls backend run endpoint; currently does not await terminal result |
| `sidecar/agents/fixer.py` | Fix proposal |

## Test Map

**Verified:** Colocated Go `_test.go` coverage exists for agent, API, auth, compare, events, gitdiff, intelligence, metrics, schedule, and visual packages. API/schedule tests use memory stores (`internal/api/api_test.go:22-28,265`; `internal/schedule/store_test.go:59`).

**Verified absence:** No frontend test script/dependency (`frontend/package.json:5-9,20-29`) and no Python test files found. PostgreSQL integration-test usage was not found. See `TODO-012`–`TODO-014`.

## Critical Files and Hotspots

| Path | Why high-risk |
|---|---|
| `internal/api/server.go` | Route registry, composition, handlers, orchestration, settings, scheduler, exports; broad blast radius |
| `internal/agent/agent.go` | Core state/model/interface semantics for MCP/reusable flow |
| `internal/agent/playwright_runner.go` | Browser action interpretation, healing, result trust, egress |
| `internal/runner/docker.go` | Generated files/config, Docker invocation, host networking, artifacts |
| `cmd/server/main.go` | Store fallback, migration policy, scheduler/listener lifecycle |
| `internal/db/migrations/*.sql` | Physical schema and data integrity |
| `frontend/src/lib/api.ts` | Entire browser API contract and SSE behavior |
| `docker-compose.yml` | Network exposure, credentials, service topology |
| `sidecar/main.py`, `sidecar/agents/executor.py` | Unauthenticated job boundary and backend credential forwarding |

## Common Development Paths

### Add/change an HTTP endpoint

1. Register route and middleware scope in `internal/api/server.go:124-261`.
2. Implement/adjust handler; prefer existing `isValidID` and store/domain functions.
3. Add Go `httptest` coverage in `internal/api/*_test.go`.
4. Update frontend `api.ts` only if consumed by dashboard.
5. Update [`API.md`](API.md), affected [`DOMAIN.md`](DOMAIN.md), TODO/ADR, and `CHANGELOG_AI.md`.
6. Preserve `/api/v1` compatibility or document breaking change/migration.

### Add/change persisted data

1. Add a new forward migration; never edit applied migration semantics without a recovery plan.
2. Update PostgreSQL and memory store/domain mapping.
3. Add PostgreSQL integration tests and migration test.
4. Update [`DATABASE.md`](DATABASE.md), [`DOMAIN.md`](DOMAIN.md), API DTOs if exposed, and changelog.

### Add a frontend page/flow

1. Add `frontend/src/app/<route>/page.tsx` using existing local-hook/API patterns.
2. Reuse `components/ui`, `components/console`, `lib/api.ts`, and `lib/utils.ts`.
3. Add navigation intentionally (not every route is currently in sidebar).
4. Add tests and update API/domain docs.

### Change execution behavior

1. Read all five execution paths and accepted `ADR-001`/`ADR-002` status.
2. Lock current behavior with focused tests before consolidation.
3. Keep state/result/event semantics consistent across entry points.
4. Update architecture/domain/API/state/TODO/changelog documents.

### Add/change a dependency

1. Prove existing dependencies/utilities cannot satisfy the need.
2. Update the correct manifest and lockfile.
3. Document purpose, version, wiring, replacement considerations in [`DEPENDENCIES.md`](DEPENDENCIES.md).
4. Run relevant audits/build/tests and record evidence.
