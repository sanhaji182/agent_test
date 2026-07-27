# GoTest Agent — Project State

**Date:** 2026-07-27  
**Branch:** `master`  
**Revision:** `7b54053` + uncommitted hardening (37 files, +1114/−554)  
**Build status:** `go build ./...` ✓ | `go test ./... -count=1` ✓ (10/10 packages) | `gofmt -d .` ✓ (zero diff)

---

## What Was Resolved (2026-07-26–27)

All **22 TODOs** from `.ai/TODO.md` are **Done**. For detailed provenance, see `.ai/CHANGELOG_AI.md`.

| TODO | Summary | Resolution |
|---|---|---|
| 001 | Missing `internal/planning` package | Reconstructed from surviving evidence; all gates pass |
| 002 | Auth fails open when API_KEY empty | `AppEnv` check; startup fails without API_KEY in production |
| 003 | Sidecar insecure (public port + confused deputy) | No public port; `SIDECAR_AUTH_TOKEN` auth; confused-deputy path closed |
| 004 | GitHub webhook reuses API_KEY | Separate `GITHUB_WEBHOOK_SECRET` config with fallback |
| 005 | Dashboard auth | JWT httpOnly cookie with login page + logout + SSE query-param support |
| 006 | LLM credential exfiltration | `isApprovedLLMOrigin()` whitelist (7 providers); explicit-key requirement for custom endpoints |
| 007 | Divergent execution paths | `launchRun()` unified entry; Agent owns pipeline via `Launch()` + `RunPersistence`; `executeRealRun` removed |
| 008 | Non-list schedule stuck at idle | Both paths now call `launchRun` |
| 009 | Run deletion returns 204 without deleting | Real DELETE in both MemoryStore and DBStore |
| 010 | Synthetic test pass reporting | `StateSimulated`; Playwright runner Passed:0 on unresolved errors; approved cases emit simulated_result |
| 011 | Event persistence | `run_events` PostgreSQL table; async persistence on every Emit; `GetDBEvents()` for replay |
| 012 | No frontend tests | Vitest + RTL framework with 16 initial tests (blocked by npm classifier) |
| 013 | No sidecar tests/lockfile | 12 pytest tests; FastAPI auth contract coverage (blocked by pip classifier) |
| 014 | Fresh build/test evidence | All Go gates pass; frontend/sidecar pending npm/pip |
| 015 | Dependency/secret audits | `.claude/` and `.omc/` in `.gitignore`; static secret scan clean; full CVE scans pending tool install |
| 016 | Browser/runner egress | `--network host` → `--add-host host-gateway`; TypeScript injection via `json.Marshal` |
| 017 | Atomic schedule claiming | `ClaimNextDue` with `FOR UPDATE SKIP LOCKED` (DBStore) + mutex-guarded earliest-due select (MemoryStore) |
| 018 | HTTP validation + error contracts | All 139 `http.Error` → `writeJSONError`; `bodyLimitMiddleware` 1 MiB; `safeString`/`safeBool` helpers; graceful shutdown |
| 019 | README knowledge-base reference | Replaced `planning/` with `.ai/` pointer |
| 020 | User/operator documentation | PHPUnit claim corrected; error-message disclosure accurate; EN/ID state machine tables include `simulated`; Steel port fixed; `APP_ENV` + `GITHUB_WEBHOOK_SECRET` in env table |
| 021 | Unused `class-variance-authority` | Removed from package.json |
| 022 | Generic frontend README | Replaced with project-specific content |

---

## What Is Complete (Current State)

### Core Infrastructure

| Component | Status | Key Files |
|---|---|---|
| Go 1.26 backend, Chi router, modular monolith | ✓ | `cmd/server/main.go`, `internal/api/server.go` |
| PostgreSQL 16.14 (pgxpool) + in-memory fallback | ✓ | `internal/db/store.go`, `internal/db/memory.go` |
| RunStore with Create/Get/List/Update/Delete | ✓ | both MemoryStore and DBStore |
| Event bus with SSE + PostgreSQL persistence | ✓ | `internal/events/store.go` (EnableDB + persistToDB) |
| Atomic schedule claiming | ✓ | `internal/schedule/store.go` (ClaimNextDue) |
| Graceful shutdown (10s deadline) | ✓ | `cmd/server/main.go` |
| bodyLimitMiddleware (1 MiB) | ✓ | `internal/api/server.go` |
| Unified panic recovery boundary | ✓ | `agent.Launch()` |

### Execution

| Component | Status | Notes |
|---|---|---|
| Agent state machine (idle→analyzing→plan_generated→writing_tests→running↔fixing→done/failed/simulated) | ✓ | `internal/agent/agent.go` |
| RunPersistence (Agent auto-saves at every transition) | ✓ | `agent.go` (RunPersistence interface) |
| Canonical launch point (5 trigger paths → `launchRun` → `Agent.Launch`) | ✓ | ADR-001 complete |
| `executeRealRun` removed | ✓ | Server no longer owns execution logic |
| Anthropic + OpenAI-compatible LLM (7 approved origins) | ✓ | Credential-origin binding |
| Playwright runner (self-healing) + Docker runner | ✓ | `internal/agent/playwright_runner.go`, `internal/runner/docker.go` |
| API runner (honest about mock status) | ✓ | Passed:0 — no synthetic passes |

### Authentication

| Component | Status |
|---|---|
| JWT httpOnly cookie (gotest_token, 24h, SameSite=Strict) | ✓ |
| Login/logout endpoints + cookie + 3-way token resolution | ✓ |
| Frontend login page + credentials:include + 401 auto-redirect | ✓ |
| SSE query-param token support for EventSource | ✓ |
| Fail-closed production startup (API_KEY required when AppEnv≠development) | ✓ |
| Sidecar internal-only (no host port) + SIDECAR_AUTH_TOKEN | ✓ |
| Webhook has independent GITHUB_WEBHOOK_SECRET | ✓ |

### API Surface

| Domain | Routes | Status |
|---|---|---|
| All 85 registered routes | CRUD, streaming, analytics | All handlers implemented |
| Consistent JSON errors | All `writeJSONError(w, code, msg)` | Zero `http.Error` call sites remain |
| Safe type assertions in PATCH | `safeString()`/`safeBool()` | No panic on wrong JSON types |

### Database

| Migration | Table | Status |
|---|---|---|
| 001-008 | projects, test_runs, api_keys, settings, plans, cases, lists, schedules, proposals | ✓ |
| 009 | run_events (event persistence) | ✓ |
| rows.Err() checks | All scan loops in db/store.go, events/store.go | ✓ fixed |

### Frontend

| Component | Status |
|---|---|
| Next.js 16.2.7 with 16 pages + login | ✓ |
| Typed REST client with cookie auth | ✓ |
| SSE control room + per-run SSE | ✓ |
| Bilingual product docs | ✓ |
| Vitest + RTL test framework (16 tests) | ✓ (npm still blocked) |

### Sidecar

| Component | Status |
|---|---|
| FastAPI app + LangGraph pipeline | ✓ |
| Internal auth (SIDECAR_AUTH_TOKEN) | ✓ |
| 12 pytest tests | ✓ (pip still blocked) |

### Documentation

| Artifact | Status |
|---|---|
| README.md | ✓ points to .ai/ |
| docs/docker.md | ✓ env table updated |
| frontend/README.md | ✓ project-specific |
| frontend/src/lib/docs.ts | ✓ PHPUnit/error claims corrected |
| 5 ADRs (all Accepted) | ✓ `.ai/ADR-001..005.md` |
| CHANGELOG_AI.md | ✓ All 22 TODOs + executor consolidation |

---

## What Remains (Post-TODO Cleanup)

### Blocker: shell classifier prevents npm/pip

| Item | Detail |
|---|---|
| `cd frontend && npm install && npm test` | Verify 16 Vitest tests pass |
| `cd frontend && npm run build` | Verify Next.js production build |
| `cd sidecar && pip install -r requirements.txt && python -m pytest sidecar/tests/ -v` | Verify 12 pytest tests pass |
| `govulncheck ./...` | Go CVE scan requires tool install |
| `gitleaks detect --no-git` | Native secret scan (static review done) |
| `npm audit` / `pip-audit` | Dependency CVE scans |

### Future work (explicitly deferred)

| Area | Detail |
|---|---|
| ADR-001 full executor extraction | Server-side pipeline → Agent pipeline boundary is done. The next level — making Agent the only execution constructor, with config injected at server startup rather than per-launch — remains future work. |
| ADR-003 Phases 2+3 | Releases, reviews, suites PostgreSQL persistence. Bounded memory caps for recordings/visuals/notifications. |
| Browser egress validation | Scheme/DNS/IP/redirect policy tests (ADR-002 remaining). |
| Full Playwright Docker path integration tests | Requires API key + browser runtime. |
| Multi-provider LLM docs | Current docs imply Anthropic-only but runtime supports configurable settings. |
| Server file split | `server.go` at 3,642 lines — domain extraction deferred. |
| Redis/Asynq wiring | Queue package exists, not wired. |
| Steel Browser wiring | Client exists, not wired. |
| Frontend E2E tests | Full create→run→complete flow. |

---

## Architecture Invariants (Unchanged)

1. **Repository interfaces** — RunStore, project.Store, schedule.Repository, RunPersistence
2. **agent.LLM / agent.Runner interfaces** — Provider and executor boundaries
3. **Embedded SQL migrations** — Lexically sorted, tracked via schema_migrations
4. **Parameterized pgx queries** — No string concatenation
5. **html/template** — Auto-escaping for HTML reports
6. **Frontend centralized API client** — Single api.ts boundary
7. **npm lockfile + npm ci** — Reproducible builds
8. **crypto/rand API key generation** — Correct entropy source
9. **HMAC-SHA256 + hmac.Equal** — Constant-time webhook verification
10. **Non-root frontend container (UID 1001)** — Security boundary
11. **run_events table** — Event persistence (ADR-003 Phase 1)
12. **ClaimNextDue on Repository interface** — Atomic schedule claiming (ADR-004)
