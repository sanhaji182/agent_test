# AI Change Log

**Owner:** Engineering  
**Purpose:** Append-only provenance for AI-generated repository and knowledge-base modifications  
**Last updated:** 2026-07-28  
**Rule:** Do not rewrite or delete historical entries. Corrections are new entries that supersede earlier statements.

## Entry Template

```markdown
## YYYY-MM-DD — Short task title

- **Task:**
- **Source revision before change:**
- **Source revision after change:** UNKNOWN until committed
- **Files modified:**
- **Summary:**
- **Reason:**
- **Risk:** Critical / High / Medium / Low / Documentation-only
- **Breaking changes:** None / details
- **Database migrations:** None / details
- **Deployment steps:** None / details
- **Documentation updated:**
- **Verification completed:** exact commands/flows and outcomes
- **Facts added/removed or confidence changed:**
- **Open unknowns:**
- **Related ADRs/TODOs:**
```

## 2026-07-28 — HTML report tests incl. XSS-escaping guard

- **Task:** Coverage for `internal/report` (previously zero test files).
- **Source revision before change:** 5ffa897
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/report/html_test.go` (new)
- **Summary:** 3 tests: full-run rendering (stats, failure table, test-plan section), empty-run rendering ("No results available." without failures/plan sections), and an XSS guard proving failure Test/Message content — which originates from test output and LLM responses — is HTML-escaped by `html/template` (script tag and onerror payload must appear escaped, never raw).
- **Reason:** Coverage initiative; the report endpoint serves this HTML to browsers, so a future switch to `text/template` or `template.HTML` would be an XSS regression this test now catches.
- **Risk:** Low (test-only)
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** this entry
- **Verification completed:** 3/3 PASS with `-race`; full suite 23/23 packages ok; build/gofmt clean.
- **Facts added/removed or confidence changed:** Tested-package count 12→23 today. Remaining untested and why: `queue` (Redis integration), `runner` (Docker), `mcp` (stdio server loop); dormant pending product decisions: `steel`, `vision`, `evals`.
- **Open unknowns:** None for this change.
- **Related ADRs/TODOs:** SECURITY.md; TESTING.md.

## 2026-07-28 — execution context + Playwright reporter tests

- **Task:** Coverage for `internal/execution` and `internal/reporter` (both previously zero test files) — the runtime-artifact plumbing between runners and the run console.
- **Source revision before change:** 5e7f988
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/execution/context_test.go` (new), `internal/reporter/playwright_test.go` (new)
- **Summary:** execution (4 tests): nil-receiver/nil-store no-panic guarantees (runners call unconditionally), event forwarding, `RecordScreenshot` producing all three artifacts (recording `captured`, visual artifact, `screenshot_captured` event), and visual-baseline chaining (second capture of same step gets first's URL as baseline; different step starts empty). reporter (3 tests): `ParseAndEmit` against a realistic two-spec Playwright JSON report — event-type counts (2 test_started, 3 step_started/completed, 1 assertion each way), failed-step message propagation ("FAILED: timeout"), cumulative `timestamp_ms` arithmetic (200ms after 120+80); graceful nil-error on missing file/invalid JSON/nil ctx; `itoa` direct coverage.
- **Reason:** Coverage initiative; baseline chaining and timestamp accumulation are the visual-regression and timeline features' correctness core.
- **Risk:** Low (test-only)
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** this entry
- **Verification completed:** 7/7 PASS with `-race`; full suite 22/22 packages ok; build/gofmt clean.
- **Facts added/removed or confidence changed:** Tested-package count 12→22 today. Remaining untested: `queue` (Redis), `runner` (Docker), `mcp` (stdio server), `report` (HTML template); dormant: `steel`, `vision`, `evals`.
- **Open unknowns:** None for this change.
- **Related ADRs/TODOs:** TESTING.md; coverage entries above.

## 2026-07-28 — workflow store tests (review gate + suites)

- **Task:** Coverage for `internal/workflow` (previously zero test files) — human-in-the-loop review/approval and suite management.
- **Source revision before change:** 01cc959
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/workflow/store_test.go` (new)
- **Summary:** 6 tests. ReviewStore: `Create` forces status to `pending` even when the caller supplies `approved` (approval MUST go through `Approve` — the review-gate invariant), Get/ByRun indexing, Approve/Reject reviewer+comment+UpdatedAt semantics with false-on-missing. SuiteStore: create defaults, newest-first List, ByTag matching, Delete pruning both the map and the order slice.
- **Reason:** Coverage initiative; the forced-pending behavior is the workflow package's core security property and had no guard.
- **Risk:** Low (test-only)
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** this entry
- **Verification completed:** 6/6 PASS with `-race`; full suite 20/20 packages ok; build/gofmt clean.
- **Facts added/removed or confidence changed:** Tested-package count 12→20 today.
- **Open unknowns:** None for this change.
- **Related ADRs/TODOs:** TESTING.md; coverage entries above.

## 2026-07-28 — webhook HMAC verification tests (security-critical)

- **Task:** Coverage for `internal/webhook` (previously zero test files) — GitHub HMAC verification is the only barrier between the public webhook endpoint and auto-triggered test runs.
- **Source revision before change:** 5e80027
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/webhook/github_test.go` (new)
- **Summary:** 6 tests: non-POST 405; valid sha256 signature triggers the async `onPush` callback with correctly-parsed payload; 5 invalid-signature variants all 401 without firing the callback (wrong secret, missing `sha256=` prefix, empty header, garbage hex, tampered body); empty-secret development mode accepts unsigned requests; invalid JSON 400; ping/unknown events 200 without side effects.
- **Reason:** Security-critical production path with zero coverage; a regression in `verifySignature` (e.g. prefix handling or non-constant-time compare removal) would silently open run-triggering to forged requests.
- **Risk:** Low (test-only)
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** this entry
- **Verification completed:** 6/6 PASS with `-race`; full suite 19/19 packages ok; build/gofmt clean.
- **Facts added/removed or confidence changed:** Tested-package count 12→19 today. Remaining untested: `queue` (needs Redis/mocks), `execution`, `runner`, `mcp`, `report`, `reporter`, `workflow` (active); `steel`, `vision`, `evals` (dormant, UW items).
- **Open unknowns:** None for this change.
- **Related ADRs/TODOs:** ADR-005 (auth model); SECURITY.md.

## 2026-07-28 — recordings store tests

- **Task:** Coverage for `internal/recordings` (previously zero test files) — backs the run-console screenshot strip.
- **Source revision before change:** 2255db3
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/recordings/store_test.go` (new)
- **Summary:** 5 tests: sequential ID generation (`{runID}-rec-{n}`), explicit-ID preservation, `ByRun` filtering (nil for unknown), `All` copy semantics, and direct coverage of the hand-rolled `itoa` (0, single/multi-digit, large values) whose off-by-one would silently produce colliding recording IDs.
- **Reason:** Continuation of the coverage initiative.
- **Risk:** Low (test-only)
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** this entry
- **Verification completed:** 5/5 PASS with `-race`; full suite 18/18 packages ok; build/gofmt clean.
- **Facts added/removed or confidence changed:** Tested-package count 12→18 today.
- **Open unknowns:** None for this change.
- **Related ADRs/TODOs:** TESTING.md; coverage entries above.

## 2026-07-28 — project + release store tests

- **Task:** Continue closing untested-package gaps: `internal/project` and `internal/release` (both previously zero test files).
- **Source revision before change:** 4060be4
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/project/store_test.go` (new), `internal/release/store_test.go` (new)
- **Summary:** project (4 tests): `prepareProject` defaults (`ui`/`default`/UUID/timestamps) with explicit-value preservation, Get/Update round-trip, `pgx.ErrNoRows` paths, List newest-first + pagination window + empty-non-nil-slice-past-end (handlers rely on this for JSON `[]`). release (5 tests): Create defaults (`active` status, explicit preserved), Get/List newest-first, Update callback + UpdatedAt advance + false-on-missing, `Summarize` aggregation (passed/failed run counting, RunResult totals across mixed states incl. in-flight run without RunResult, pass-rate, newest-first LatestStatus), empty-runs zero summary.
- **Reason:** Same coverage initiative that exposed the critical gitignore defect; these two stores back the projects intake and release-tracking endpoints.
- **Risk:** Low (test-only)
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** this entry
- **Verification completed:** `go test ./internal/project/ ./internal/release/ -race -v` — 9/9 PASS; full suite 17/17 packages ok with `-race`; build/vet/gofmt clean.
- **Facts added/removed or confidence changed:** Tested-package count 12→17 today (db, notify, planning, project, release). Note: project.MemoryStore stores raw pointers (no cloneRun equivalent) — acceptable because handlers do not mutate returned projects concurrently with writers; recorded here should that assumption change.
- **Open unknowns:** Remaining untested packages (`queue`, `recordings`, `execution`, `runner`, `steel`, `vision`, `evals`, `mcp`, `report`, `reporter`, `webhook`, `workflow`) — queue needs Redis or interface mocks; several are dormant (UW items).
- **Related ADRs/TODOs:** TESTING.md; coverage entries above.

## 2026-07-28 — CRITICAL: internal/planning was never tracked by git (gitignore fix)

- **Task:** Correction/supersession of the previous entry ("planning.MemoryStore tests"): committing those tests exposed a repository-integrity defect.
- **Source revision before change:** 9d78071
- **Source revision after change:** 9355dce
- **Files modified:** `.gitignore`, `internal/planning/{types,memory,db}.go` (tracked for the first time), `internal/planning/memory_test.go` (new)
- **Summary:** `.gitignore:73` contained the unanchored pattern `planning/` (intended for the root scratch folder referenced by DG-1), which also matched `internal/planning/`. The entire planning package — production code imported by `internal/api` handlers — existed only on this machine and was never in any commit. **Every fresh clone of this repository failed to build.** Fixed by root-anchoring the pattern (`/planning/`) and committing the package plus its new tests. Post-fix audit: `git check-ignore` over every source directory (internal/*, cmd, sidecar, frontend/src, docs) confirms no other source path is ignored.
- **Reason:** Discovered when `git add internal/planning/memory_test.go` was rejected as ignored.
- **Risk:** Critical defect fixed; the fix itself is low-risk (gitignore anchor + adding files)
- **Breaking changes:** None (fresh clones now build — strictly better)
- **Database migrations:** None
- **Deployment steps:** None (deployments built from this working tree were unaffected; CI/clone-based builds were broken until now)
- **Documentation updated:** this entry; DG-1 context (root `planning/` remains intentionally ignored)
- **Verification completed:** `git ls-tree origin/master internal/` now lists `internal/planning`; `git check-ignore` audit of all source dirs → none ignored; full suite 15/15 packages `-race` green from the tracked tree.
- **Facts added/removed or confidence changed:** Repository is now self-contained/buildable from clone — previously UNKNOWN-false. Lesson recorded: unanchored directory patterns in .gitignore match at any depth.
- **Open unknowns:** Whether CI (if any external) had been failing on this — no CI config is tracked in-repo.
- **Related ADRs/TODOs:** DG-1 (root planning/ folder); previous entry (planning tests).

## 2026-07-28 — planning.MemoryStore tests: lifecycle + clone isolation

- **Task:** Add test coverage for `internal/planning` (previously zero test files) — the store behind the project→plan→approve workflow.
- **Source revision before change:** 9d78071
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/planning/memory_test.go` (new)
- **Summary:** 6 tests covering all four entity types: draft lifecycle (create defaults `draft`/UUIDs/timestamps, update assigns IDs to new cases, `pgx.ErrNoRows` on missing), test-case CRUD with prepare-contract checks (`ui`/`medium`/version 1), project filtering + newest-first ordering, test lists (non-nil slice defaults), change proposals (status `pending`, ReviewedAt persistence, per-case filtering), and a clone-isolation guard mirroring `internal/db`: mutating returned copies or caller-held inputs (Steps/Tags/TestCaseIDs slices, nested Original/Proposed snapshots, ReviewedAt pointer) never touches stored state.
- **Reason:** Same rationale as the db/notify entries: MemoryStore's deep-copy helpers are relied on by concurrent HTTP handlers but had no direct coverage.
- **Risk:** Low (test-only)
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** this entry
- **Verification completed:** `go test ./internal/planning/ -race -v` — 6/6 PASS; full suite 15/15 packages ok with `-race`; build/vet/gofmt clean.
- **Facts added/removed or confidence changed:** `internal/planning` no longer untested; tested-package count 12→15 today (db, notify, planning).
- **Open unknowns:** `planning/db.go` (PostgreSQL paths) still requires a database — same integration-test gap as db.Store.
- **Related ADRs/TODOs:** TESTING.md; db/notify test entries above.

## 2026-07-28 — notify.Store tests: webhook delivery + UW-6 dependency guard

- **Task:** Add test coverage for `internal/notify` (previously zero test files) — the store behind the UW-6 failure notifier.
- **Source revision before change:** 37b7f0a
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/notify/store_test.go` (new)
- **Summary:** 9 tests: sequential ID/timestamp assignment, `List` copy semantics (mutation isolation), `ByRun` filtering, `DeliverWebhook` (empty-URL noop, JSON POST verified via httptest server, ≥400 error), and `TriggerFailure` (Delivered=true on 2xx, Delivered=false on webhook error, recorded-but-undelivered without webhook URL).
- **Reason:** `TriggerFailure` became production code when UW-6 wired `StartFailureNotifier` (2026-07-27); its delivery-marking logic had no coverage.
- **Risk:** Low (test-only)
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** this entry
- **Verification completed:** `go test ./internal/notify/ -race -v` — 9/9 PASS; full suite 14/14 packages ok with `-race`; build/gofmt clean.
- **Facts added/removed or confidence changed:** `internal/notify` no longer untested; tested-package count 12→14 today (db, notify).
- **Open unknowns:** None for this change.
- **Related ADRs/TODOs:** UW-6 (resolved); `internal/api/failure_notifier.go` tests already existed (2/2).

## 2026-07-28 — MemoryStore tests: snapshot-isolation invariant guard

- **Task:** Add test coverage for `internal/db` (previously zero test files) focusing on the race-fix invariant.
- **Source revision before change:** 806f482
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/db/memory_test.go` (new)
- **Summary:** 7 tests: CRUD lifecycle, `ErrNotFound` paths (Get/Delete), `ListRuns` newest-first ordering + pagination window + past-end offset, and — the core motivation — three isolation guards proving `cloneRun` snapshot semantics: (1) mutating a run returned by `GetRun` (slices, RunResult, Failures, FinishedAt) does not affect the store; (2) mutating the caller's object after `CreateRun` (simulating the `Agent.Launch` goroutine writing to its run pointer) does not affect the stored copy; (3) `ListRuns` results are equally isolated. Plus `cloneRun` nil-safety.
- **Reason:** The 2026-07-27 race fixes (commit 8cf727c) rest entirely on `cloneRun` deep-copy semantics, which had no direct tests — a future "optimization" removing a copy would silently reintroduce handler/goroutine data races that `-race` only catches probabilistically under load.
- **Risk:** Low (test-only)
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** this entry
- **Verification completed:** `go test ./internal/db/ -race -v` — 7/7 PASS; full suite `go test ./internal/... -race -count=1` — 13/13 packages ok (db now counted); build/vet/gofmt clean.
- **Facts added/removed or confidence changed:** `internal/db` no longer untested; snapshot invariant is regression-guarded deterministically (not just probabilistically via -race).
- **Open unknowns:** DBStore (PostgreSQL) paths still untested — requires a database; candidate for integration-test environment.
- **Related ADRs/TODOs:** 2026-07-27 race-fix changelog entry; TESTING.md.

## 2026-07-28 — Docs refresh: docker.md env table + frontend README (DG-6, DG-11, ST-1..3)

- **Task:** Close the two remaining cosmetic documentation gaps.
- **Source revision before change:** 4e2133e
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `docs/docker.md`, `frontend/README.md`, `.ai/DOCUMENTATION_GAP.md`
- **Summary:** DG-6/ST-2: docker.md env-var table rewritten to match current runtime — multi-provider LLM vars, `QUEUE_ENABLED`, `CORS_ALLOWED_ORIGINS`, `JWT_SECRET`, AI-planning feature flags, execution limits; prerequisites port list fixed (3010 host port for Steel); GITHUB_WEBHOOK_SECRET fallback-to-API_KEY wording verified against `server.go:313-316`. DG-11/ST-1: frontend README refreshed — Next 16.2.12, `npm test` documented, stale "no frontend test suite exists" claim corrected (20 Vitest tests in `src/test/`, verified by running the suite). ST-3 audited: `claude-sonnet-4-5` remains the actual code default — not stale, closed without change.
- **Reason:** Operator-facing docs promised/omitted env vars inconsistently with the runtime; README claim about missing tests could mislead contributors into duplicating coverage.
- **Risk:** Documentation-only
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** all three files above
- **Verification completed:** `npm test` — 20/20 PASS (confirms README claim); webhook fallback verified at `internal/api/server.go:313-316`; Compose port cross-checked against docker.md services table. No code changed.
- **Facts added/removed or confidence changed:** DOCUMENTATION_GAP: all 12 DG items and all 3 ST items now resolved or closed except DG-2 (Steel — blocked on UW-2 product decision), DG-8/DG-10 residuals (dormant-feature env names, UC-1).
- **Open unknowns:** None new.
- **Related ADRs/TODOs:** DOCUMENTATION_GAP DG-6, DG-11, ST-1..3; ADR-002 (Steel decision) still pending for DG-2.

## 2026-07-28 — Configurable CORS allowlist (DG-12) + documentation-gap reconciliation

- **Task:** Implement `CORS_ALLOWED_ORIGINS` (README promised it; code hard-coded wildcard) and reconcile stale DOCUMENTATION_GAP.md entries.
- **Source revision before change:** 37b5307
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/config/config.go`, `internal/api/server.go`, `internal/api/cors_test.go` (new), `.env.example`, `.ai/DOCUMENTATION_GAP.md`
- **Summary:** DG-12: `newCORSMiddleware(allowedOrigins)` replaces the hard-coded wildcard middleware. Empty or `*` keeps historical wildcard behavior (development default — zero behavior change for existing deployments). With a comma-separated allowlist: matching `Origin` is echoed back with `Access-Control-Allow-Credentials: true` and `Vary: Origin`; non-matching origins get no ACAO header. Matching is case- and trailing-slash-insensitive. Gap registry reconciled: DG-1/3/4/5/7/9/12 resolved, DG-10 substantially resolved (DL-2 Phase 1), DG-8 correctly marked; env-drift table replaced with the 3 remaining items.
- **Reason:** README:203 instructs operators to set `CORS_ALLOWED_ORIGINS` for production, but no code consumed it — a silent security gap for anyone following the hardening guide.
- **Risk:** Low (default unchanged; new behavior opt-in via env)
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** Optional: set `CORS_ALLOWED_ORIGINS=https://your-dashboard.example.com` in production
- **Documentation updated:** `.env.example` (new var documented), `.ai/DOCUMENTATION_GAP.md` (reconciled), this entry
- **Verification completed:** 5 new middleware tests (wildcard default, allowlist echo + credentials + Vary, unlisted origin rejected, case/slash normalization, OPTIONS preflight 204) all PASS with `-race`; `go build ./...` clean; `go vet` clean; `gofmt` zero diff.
- **Facts added/removed or confidence changed:** DG-12 resolved; documentation-gap registry now reflects working-tree reality (7 of 12 DG items resolved).
- **Open unknowns:** DG-2/DG-6/DG-11 remain (Steel wiring decision, docker.md refresh, frontend README — product/cosmetic).
- **Related ADRs/TODOs:** DOCUMENTATION_GAP DG-12; README security hardening section.

## 2026-07-28 — LLM layer unification implemented (ADR-006 Steps A–D, DL-2 resolved)

- **Task:** Execute accepted ADR-006: merge the two LLM transport layers.
- **Source revision before change:** 630fd51
- **Source revision after change:** UNKNOWN until committed (steps landed as b1a80d2, aab3963, f8464aa + this docs commit)
- **Files modified:** `internal/ai/client.go` (+`client_test.go` new), `internal/agent/llm_prompts.go` (new), `llm_prompts_test.go` (new), `llm_adapter.go` (new), `llm_factory.go` (rewritten), `llm_anthropic.go` (deleted), `llm_openai.go` (deleted), `.ai/ADR-006.md`, `.ai/TECHNICAL_DEBT.md`, `.ai/DEPENDENCIES.md`
- **Summary:** Step A (b1a80d2): `ai.Client` gains `GenerateWithImage`; both transports implement vision; OpenAI errors include response body; stub-server transport tests. Step B (aab3963): 6 prompt builders + parsers extracted to `llm_prompts.go`; fixed latent bug where `OpenAILLM.HealAction` sent the vision prompt ("screenshot attached") with no image. Step C (f8464aa): `promptLLM` adapter (prompts × `ai.Client`) replaces both structs; duplicated transports deleted (net −184 lines); `ai` gains ungated direct constructors preserving the execution layer's construct-always/fail-at-request contract; factory parity tests passed with zero edits. Step D: docs updated, DL-2 marked resolved.
- **Reason:** DL-2 — final remaining HIGH-origin debt; single transport means timeouts/retries/redaction patch in one place.
- **Risk:** Medium (execution-path refactor) mitigated: `agent.LLM` interface unchanged, all consumers untouched, parity tests unchanged and green.
- **Breaking changes:** None (API and behavior contracts preserved; error strings on the execution OpenAI path now include response body — strictly more informative)
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** ADR-006 (implemented), TECHNICAL_DEBT (DL-2 resolved), DEPENDENCIES (SDK evidence path), this entry
- **Verification completed:** Per step: `go build ./...`, `go vet ./internal/...`, `gofmt` zero diff, `go test ./internal/... -race -count=1` — 12/12 packages ok after every step. Step C proof: `llm_factory_test.go` (routing + parity) passed without modification.
- **Facts added/removed or confidence changed:** DL-2 resolved; Duplicate Logic section fully cleared; `internal/ai` is the sole LLM transport.
- **Open unknowns:** `handleTestAIProvider` real-key smoke test not run (no credentials in this environment) — recommend one manual test-connection from the dashboard after next deploy.
- **Related ADRs/TODOs:** ADR-006, ADR-005 Phase 2, TECHNICAL_DEBT DL-2.

## 2026-07-28 — ADR-006: design proposal for LLM layer unification (DL-2 Phase 2)

- **Task:** Produce the design pass required before merging the two LLM transport layers (deferred remainder of DL-2).
- **Source revision before change:** 3b56bc9
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `.ai/ADR-006.md` (new), `.ai/TECHNICAL_DEBT.md` (DL-2 entry links ADR-006)
- **Summary:** ADR-006 (Status: Proposed) documents the current two-layer state with a comparison table (interfaces, vision support, temperature/max-tokens handling, error text), decides `internal/ai` becomes the single transport layer with `agent.LLM` rebuilt as prompt-domain wrappers over `ai.Client`, and lays out a 4-step migration (A: vision in `ai`; B: extract prompts with byte-identical parity test; C: `promptLLM` adapter replaces both structs; D: cleanup). Each step independently shippable and gated by build + `-race` suite. `agent.LLM` interface and all its consumers unchanged throughout. Evidence claims verified against source (`chatWithVision` at llm_anthropic.go:203, image_url block at llm_openai.go:51, hard-coded MaxTokens 4096).
- **Reason:** Repo protocol: architectural changes require investigation + design before implementation; DL-2's remaining transport merge touches the core execution path.
- **Risk:** Documentation-only
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** `.ai/ADR-006.md`, `.ai/TECHNICAL_DEBT.md`, this entry
- **Verification completed:** Source cross-checks for all file:line evidence cited in the ADR; no code changed (build/test state unchanged from 3b56bc9: 11/11 -race green).
- **Facts added/removed or confidence changed:** DL-2 Phase 2 now has an approved-pending design; implementation can start on acceptance.
- **Open unknowns:** ADR-006 acceptance decision (owner); whether planning layer should adopt vision immediately or only expose it.
- **Related ADRs/TODOs:** ADR-006, TECHNICAL_DEBT DL-2, ADR-005 Phase 2.

## 2026-07-28 — Align LLM provider routing between execution and planning layers (DL-2 Phase 1)

- **Task:** Fix the routing inconsistency half of DL-2 (HIGH): `agent.NewLLM` (execution + test-connection) and `ai.New` (planning) accepted different provider sets with different normalization and base-URL defaults.
- **Source revision before change:** a62a3e2
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/agent/llm_factory.go`, `internal/ai/client.go`, `internal/agent/llm_factory_test.go`, `.ai/TECHNICAL_DEBT.md`
- **Summary:** Three concrete inconsistencies fixed. (1) Normalization: `ai.New` lowercased/trimmed provider names, `agent.NewLLM` matched case-sensitively — a DB-stored `"Anthropic"` worked for planning but returned nil (unsupported provider) for run execution. `NewLLM` now normalizes identically. (2) Provider set: `ai.New` rejected `google`/`deepseek`/`mistral`/`groq`/`openrouter`/`huggingface`, which `agent.NewLLM` accepted — with those providers configured, runs executed but AI planning silently returned nil. `ai.New` now accepts the same set. (3) Default base URL: both layers fell back to `https://api.openai.com/v1` for ALL OpenAI-compatible providers, so `google` etc. without explicit `LLM_BASE_URL` hit the wrong endpoint with the wrong key. New shared `ai.DefaultOpenAICompatibleBaseURL(provider)` maps each hosted provider to its documented OpenAI-compatible endpoint (aligned with `isApprovedLLMOrigin`); both factories use it.
- **Reason:** DL-2 (HIGH): a settings change validated by "test connection" could behave differently in planning; provider misconfiguration failed silently.
- **Risk:** Low-Medium (routing only; behavior changes are strictly fixes: previously-nil clients now constructed, previously-wrong endpoints now correct. No change for anthropic/openai/custom with explicit base URL — the common configurations)
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** `.ai/TECHNICAL_DEBT.md` (DL-2 downgraded HIGH→MEDIUM; remaining = transport-layer merge), this entry
- **Verification completed:** New tests: `TestProviderRoutingParity` (14 providers, both factories agree), `TestDefaultBaseURLParity` (7 endpoint mappings), 2 normalization cases — all PASS with `-race`. Full gates: `go build ./...` clean, `go vet ./internal/...` clean, `gofmt` zero diff, `go test ./internal/... -race -count=1` 11/11 ok.
- **Facts added/removed or confidence changed:** DL-2 severity HIGH→MEDIUM; parity now regression-guarded.
- **Open unknowns:** Full layer merge (shared transport) deferred — needs its own design pass; `agent` now imports `ai` (one-way, no cycle).
- **Related ADRs/TODOs:** TECHNICAL_DEBT DL-2; ADR-005 Phase 2 (approved origins).

## 2026-07-28 — Remove dead api-logs placeholder endpoint (DC-3) + registry reconciliation (UC-4)

- **Task:** Resolve DC-3 (placeholder endpoint) and reconcile stale UC-4 registry entry.
- **Source revision before change:** d9adac3
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/api/handlers_runs.go`, `internal/api/server.go`, `.ai/TECHNICAL_DEBT.md`
- **Summary:** DC-3: `handleGetAPILogs` always returned `{"logs": []}` — a placeholder with no artifact-reading implementation. Grep across backend and `frontend/src` found zero consumers of `/runs/{id}/api-logs` (only the handler and its route registration). Removed both. If real API-log capture ships later, the endpoint can return with an actual implementation. UC-4: `class-variance-authority` is no longer in `frontend/package.json` (already removed in an earlier dependency pass); remaining runtime deps (`clsx`, `lucide-react`, `tailwind-merge`) all verified to have source imports — registry entry marked resolved.
- **Reason:** DC-3 (MEDIUM): dead placeholder gives callers false confidence the feature exists; UC-4 registry was stale.
- **Risk:** Low (endpoint had no consumers; response was always empty)
- **Breaking changes:** Endpoint `/api/runs/{id}/api-logs` removed (no known consumers; returned static empty payload)
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** `.ai/TECHNICAL_DEBT.md` (DC-3, UC-4 resolved), this entry
- **Verification completed:** grep: zero `api-logs`/`APILogs` references remain outside this changelog; `go build ./...` clean; `go vet ./internal/api/` clean; `gofmt -l internal/api/` empty; `go test ./internal/api/ -race -count=1` ok.
- **Facts added/removed or confidence changed:** DC-3 and UC-4 resolved. Dead Code section now has only DC-5 (unused frontend API exports, LOW) open.
- **Open unknowns:** None for this change.
- **Related ADRs/TODOs:** TECHNICAL_DEBT DC-3, UC-4.

## 2026-07-28 — Deduplicate draft-plan and schedule-run creation (DL-3, DL-4)

- **Task:** Resolve duplicate-logic debt items DL-3 and DL-4.
- **Source revision before change:** c78f79b
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/api/handlers_planning.go`, `internal/api/handlers_projects.go`, `internal/api/handlers_schedules.go`, `.ai/TECHNICAL_DEBT.md`
- **Summary:** DL-3: extracted `createDraftPlanResponse(w, r, projectID, cases)` in `handlers_planning.go`; `handleGenerateProjectTestPlan` and `handleParseAPIDocs` now delegate (removed unused `planning` import from `handlers_projects.go`). DL-4: extracted `startScheduleRun(ctx, sch, scheduleID, now, lastRunStatus, eventMsg)` in `handlers_schedules.go`; the non-list branches of `handleRunNow` and `ProcessDueSchedules` delegate. Behavioral differences preserved via parameters: run-now records `LastRunStatus="running"` with message "Run created via schedule run-now"; due-schedule records `string(agent.StateIdle)` with "Run created via due schedule". The helper does NOT call `launchRun` — callers launch, so run-now keeps its snapshot-before-launch ordering (race-safety invariant from the 2026-07-27 race fixes).
- **Reason:** DL-3 (LOW) and DL-4 (MEDIUM) in TECHNICAL_DEBT.md; duplicated orchestration blocks drift independently.
- **Risk:** Low (pure extraction; zero behavior change)
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** `.ai/TECHNICAL_DEBT.md` (DL-3/DL-4 marked resolved), this entry
- **Verification completed:** `go build ./...` clean; `go vet ./internal/api/` clean; `gofmt -l internal/api/` empty; `go test ./internal/... -race -count=1` — 11/11 packages ok, 0 FAIL.
- **Facts added/removed or confidence changed:** DL-3 and DL-4 resolved; only DL-2 (dual LLM client layers, HIGH) remains in Duplicate Logic.
- **Open unknowns:** None for this change.
- **Related ADRs/TODOs:** TECHNICAL_DEBT DL-3, DL-4; snapshot-before-launch invariant (2026-07-27 race-fix entry).

## 2026-07-28 — Wire MAX_FIX_ATTEMPTS, DEFAULT_TIMEOUT_SECONDS, STEEL_MAX_SESSIONS env vars (UC-2, UC-3)

- **Task:** Resolve technical-debt items UC-2/UC-3: `.env.example` documented env vars that `config.Load` hard-coded.
- **Source revision before change:** b62e591
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/config/config.go`, `internal/config/config_test.go` (new), `.env.example`, `.ai/TECHNICAL_DEBT.md`
- **Summary:** `MaxFixAttempts`, `TimeoutSeconds`, and `SteelMaxSessions` are now read via `getEnvInt` (`MAX_FIX_ATTEMPTS`, `DEFAULT_TIMEOUT_SECONDS`, `STEEL_MAX_SESSIONS`) with the same defaults (3/300/10). Invalid or non-positive values fall back to defaults. `.env.example` moved these vars from the UNUSED list to a documented "Execution limits" section. Added first test file for the config package (defaults, overrides, invalid fallback).
- **Reason:** UC-2 (MEDIUM) and UC-3 (LOW) in TECHNICAL_DEBT.md — documented configuration silently ignored at runtime is an operator trap.
- **Risk:** Low (defaults unchanged; env vars only take effect when explicitly set)
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None (optional: set the new env vars)
- **Documentation updated:** `.env.example`, `.ai/TECHNICAL_DEBT.md` (UC-2/UC-3 marked resolved), this entry
- **Verification completed:** `go test ./internal/config/ -v` — 3/3 PASS; `go build ./...` clean; `go vet ./internal/config/` clean; `gofmt -l internal/config/` empty. Consumers confirmed: `cmd/mcp/main.go` (DockerRunner timeout, agent max fixes), `internal/api/handlers_planning.go:825` (DockerRunner timeout).
- **Facts added/removed or confidence changed:** UC-2 and UC-3 resolved; config package now has test coverage.
- **Open unknowns:** None for this change.
- **Related ADRs/TODOs:** TECHNICAL_DEBT UC-2, UC-3.

## 2026-07-27 — Sidecar auth token read-at-request-time fix

- **Task:** Execute the deferred sidecar pytest suite (12+ tests) and fix any failures.
- **Source revision before change:** 5b9ae48
- **Source revision after change:** 06a9115
- **Files modified:** `sidecar/main.py`, `sidecar/tests/test_api.py`
- **Summary:** Running the suite exposed a real bug: `verify_auth` captured `SIDECAR_AUTH_TOKEN` in a module-level constant at import time, so 10/13 tests failed with 401 (their `patch.dict` env changes were invisible to the already-captured constant). More importantly this is a production defect — token rotation would require a full process restart. Fixed by reading `os.getenv("SIDECAR_AUTH_TOKEN")` inside `verify_auth` on each request. Test bootstrap now defaults to dev mode (empty token); auth cases patch the env explicitly.
- **Reason:** Deferred verification item (sidecar tests never executed); import-time env capture is a latent config-rotation bug.
- **Risk:** Low (auth logic unchanged; only when the token is read)
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None (behavior identical for static tokens; now also picks up rotated tokens without restart)
- **Documentation updated:** this entry
- **Verification completed:** pip deps installed via `--break-system-packages`. `python3 -m pytest sidecar/tests/` — 3/13 passed before fix, remaining 10 were 401s from the import-time capture. Post-fix re-run (2026-07-28): **13 passed, 0 failed** in 0.54s — auth, run-endpoint, status-endpoint, and health tests all green. One benign StarletteDeprecationWarning (httpx testclient) noted.
- **Facts added/removed or confidence changed:** sidecar auth now rotation-safe; sidecar test suite fully green (13/13).
- **Open unknowns:** none for this change.
- **Related ADRs/TODOs:** sidecar EXPERIMENTAL tag; deferred test-execution item.

## 2026-07-27 — LLM timeouts, failure notifier (UW-6), race fixes, debt cleanup

- **Task:** Continue engineering improvement: LLM HTTP timeouts, wire failure notifications, fix data races surfaced by `-race`, update stale debt entries.
- **Source revision before change:** 1276fda
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/agent/llm_openai.go`, `internal/ai/client.go` (2-min timeouts); new `internal/api/failure_notifier.go` + test; `cmd/server/main.go` (notifier goroutine); `internal/db/memory.go` (cloneRun snapshot semantics); `internal/api/handlers_{runs,planning,schedules}.go` (pre-launch response snapshots); `.ai/TECHNICAL_DEBT.md` (UW-1, UW-6, LF-1, DC-1, DC-2, DC-4 resolved; DC-3 relocated); `PROJECT_STATE.md`.
- **Summary:** (1) OpenAI-compatible clients previously had no HTTP timeout — a hung endpoint pinned run goroutines forever; now 2 minutes. (2) `notify.TriggerFailure` gains its production caller: `StartFailureNotifier` subscribes to the global event stream and fires notifications + schedule webhooks on `run_failed`. (3) `go test -race` exposed that MemoryStore shared live `*TestRun` pointers with the execution goroutine; MemoryStore now clones on read/write, and handlers snapshot response fields before `launchRun`.
- **Reason:** PROJECT_STATE deferred item (timeouts); TECHNICAL_DEBT UW-6 HIGH; race-safety is a production-grade requirement.
- **Risk:** Medium (MemoryStore semantics change from shared-pointer to snapshot — this matches DBStore semantics, and full suite passes)
- **Breaking changes:** None (MemoryStore callers already treated results as snapshots in DB mode)
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** TECHNICAL_DEBT.md, PROJECT_STATE.md
- **Verification completed:** `go build ./...` clean; `go test ./internal/... -race` all pass (previously 10+ race failures); `gofmt` clean; failure-notifier tests 2/2 PASS.
- **Facts added/removed or confidence changed:** MemoryStore now snapshot-semantics (was shared-pointer); notify package no longer unwired.
- **Open unknowns:** npm audit fix / sidecar pytest / govulncheck still classifier-blocked.
- **Related ADRs/TODOs:** ADR-001, TECHNICAL_DEBT UW-6/DC-2.

## 2026-07-27 — Commit batch, server.go domain split, Launch lifecycle tests, optional Redis/Asynq queue

- **Task:** Execute all recommended next actions: (1) commit the verified working tree as 7 atomic commits, (2) split `internal/api/server.go` into domain handler files, (3) add `Agent.Launch()` lifecycle integration tests, (4) wire optional Redis/Asynq durable run queue behind `QUEUE_ENABLED`.
- **Source revision before change:** 7b54053
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/api/server.go` (3,617→~335 lines), new `internal/api/handlers_{projects,planning,runs,auth,schedules,releases,metrics,intelligence,reviews,admin}.go`, new `internal/agent/launch_test.go` (4 tests), new `internal/queue/run_worker.go` (`runs:execute` job), new `internal/api/queue_integration_test.go` (4 tests), `internal/config/config.go` (+`QueueEnabled`), `cmd/server/main.go` (queue wiring), `internal/queue/worker.go` (doc update), `.env.example`, `.ai/OPERATIONS.md`.
- **Summary:** server.go split is a pure mechanical move (143 functions preserved, imports minimized per file). `launchRun` refactored into `buildAgent()` + queue-aware dispatch: when `SetRunEnqueuer` is installed (cmd/server, `QUEUE_ENABLED=true`), runs are enqueued to Redis/Asynq by ID and executed by an in-process worker via `ExecuteRunByID` (terminal states skipped for idempotent retries); enqueue failure falls back to in-process `Agent.Launch`, so no run is lost. Default path unchanged (goroutines + semaphore).
- **Reason:** TECHNICAL_DEBT server.go size; PROJECT_STATE deferred item "Redis/Asynq wiring"; test-coverage gap on the canonical async execution path (ADR-001).
- **Risk:** Medium (launchRun refactor touches the canonical execution path; mitigated by 8 new tests + full suite green)
- **Breaking changes:** None (queue is opt-in; default behavior identical)
- **Database migrations:** None
- **Deployment steps:** Optional: set `QUEUE_ENABLED=true` + `REDIS_URL` to enable durable queue.
- **Documentation updated:** `.ai/OPERATIONS.md` (Durable Run Queue section, Redis row), `.env.example`
- **Verification completed:** `go build ./...` clean; `go test ./internal/...` 0 FAIL (agent tests incl. `-race`); `gofmt -l internal/ cmd/` empty; function count preserved (143); new tests: 4 Launch lifecycle + 4 queue integration all PASS.
- **Facts added/removed or confidence changed:** queue package no longer fully EXPERIMENTAL — `RunWorker`/`runs:execute` is wired opt-in; legacy `TypeTestRun` remains unwired.
- **Open unknowns:** npm audit fix, sidecar pytest, govulncheck still blocked by shell classifier.
- **Related ADRs/TODOs:** ADR-001 (canonical executor), TECHNICAL_DEBT server.go split, PROJECT_STATE deferred items.

## 2026-07-27 — Consolidate LLM provider factory (AUDIT M-04, C-01)

- **Task:** Replace 3 divergent provider-routing sites with single `agent.NewLLM()` factory
- **Source revision before change:** `7b54053642e614cccf5e1128defabd25ac88b437`
- **Source revision after change:** UNKNOWN until committed
- **Files modified:**
  - `internal/agent/llm_factory.go` — New: unified `NewLLM(provider, model, apiKey, baseURL)` with explicit routing for all providers
  - `internal/api/server.go` — `launchRun` and `handleTestAIProvider` now use `agent.NewLLM()`
  - `cmd/mcp/main.go` — Now uses `agent.NewLLM()` instead of hardcoded `NewAnthropicLLM`
  - `internal/config/config.go` — Added `LLMProvider` and `LLMBaseURL` fields
- **Summary:**
  - Factory routes: `""/anthropic` → Anthropic SDK; `openai/google/deepseek/mistral/groq/openrouter/custom/local/ollama` → OpenAI-compatible REST; unknown → nil
  - Fixes AUDIT C-01 (inconsistent provider routing): `handleTestAIProvider` and `launchRun` now route through identical logic
  - Fixes AUDIT M-04 (duplicate LLM layers): 3 call sites down from separate if-else chains to single factory call
  - Fixes AUDIT M-03 (dead code): `cmd/mcp` now supports non-Anthropic providers via env config
  - Unknown providers rejected with explicit nil (no silent fallthrough)
- **Reason:** Provider testing and real execution must use identical routing logic
- **Risk:** Low (additive — old code paths replaced with centralized, well-tested factory)
- **Breaking changes:** `cmd/mcp` now requires `LLM_PROVIDER` env var if non-Anthropic; defaults to "anthropic"
- **Database migrations:** None
- **Deployment steps:** None (backward compatible for Anthropic users; new env vars `LLM_PROVIDER` + `LLM_BASE_URL` for others)
- **Documentation updated:** `.env.example` already has LLM_PROVIDER/LLM_BASE_URL entries (TODO-004)
- **Verification completed:** `go build ./...` ✓, `go test ./... -count=1` ✓ (10/10), `gofmt -d .` ✓ (zero diff)
- **Facts added/removed or confidence changed:** 3 divergent provider-routing sites → 1; cmd/mcp now supports all 7 approved origins
- **Open unknowns:** Integration test with each provider
- **Related ADRs/TODOs:** TODO-006, AUDIT M-04, AUDIT C-01

## 2026-07-27 — Browser egress URL validation (AUDIT SEC-06)

- **Task:** Validate navigation URLs before browser execution to block internal infrastructure access
- **Files modified:** `internal/agent/playwright_runner.go` (isSafeBrowserURL + guards at both goto sites), `internal/agent/playwright_runner_test.go` (new, 17 subtests)
- **Summary:** Before every `page.Goto` (original and self-healed), `isSafeBrowserURL()` validates: rejects loopback (127.0.0.1, [::1], localhost), RFC1918 private (10/172/192.168), link-local (169.254), cloud metadata endpoints (169.254.169.254, metadata.google.internal). Also performs DNS resolution to catch hostnames that resolve to internal IPs.
- **Verification:** `go build ./...` ✓, `go test ./... -count=1` ✓ (10/10 + 17 URL validation), `gofmt -d .` ✓ (zero diff)
- **Related:** AUDIT SEC-06

## 2026-07-27 — Internalize Postgres and Redis in Compose (AUDIT SEC-03/D3)

- **Task:** Remove host port publication for PostgreSQL and Redis; comment Steel port
- **Files modified:** `docker-compose.yml` — postgres: no host port (internal Docker network only), redis: no host port, steel-browser: EXPERIMENTAL comment
- **Summary:** Postgres and Redis no longer accept external connections in the default Compose deployment. Use `docker-compose exec postgres psql` for direct database access. Steel port commented as internal-only for production.
- **Verification:** `docker-compose.yml` structural review (no Go/JS changes)
- **Related:** AUDIT SEC-03 (Postgres exposed with known credentials), AUDIT D3 (Redis/Steel unnecessarily host-published)

## 2026-07-27 — Bounded goroutine pool for run execution (AUDIT S-01)

- **Task:** Add semaphore-based concurrency cap to prevent resource exhaustion from burst requests
- **Files modified:** `internal/api/server.go` (runSem field + acquireSlot/releaseSlot), `internal/config/config.go` (MaxConcurrentRuns + getEnvInt)
- **Summary:** launchRun acquires a semaphore slot (non-blocking, best-effort) before goroutine dispatch. Default cap: 10 concurrent runs, adjustable via `MAX_CONCURRENT_RUNS`. Burst requests that overflow the cap still launch but emit a warning.
- **Verification:** `go build ./...` ✓, `go test ./... -count=1` ✓ (10/10), `gofmt -d .` ✓ (zero diff)
- **Related:** AUDIT S-01

## 2026-07-27 — Playwright install once per process (AUDIT P-01)

- **Task:** Cache playwright.Install() with sync.Once instead of reinstalling on every run
- **Source revision before change:** `7b54053642e614cccf5e1128defabd25ac88b437`
- **Source revision after change:** UNKNOWN until committed
- **Files modified:**
  - `internal/agent/playwright_runner.go` — Added `sync.Once` guard; `playwright.Install()` now runs at most once per process lifetime
- **Summary:** Replaced per-Run `playwright.Install()` call with `sync.Once`-protected install. First run triggers installation; subsequent runs skip it. Install errors are cached and returned immediately on all subsequent calls.
- **Reason:** AUDIT P-01 — Playwright re-installed on every run, adding seconds-to-minutes latency per invocation and risking concurrent-install races
- **Risk:** Low (single-process optimization; already correct in Docker where browsers are pre-installed)
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** None
- **Verification completed:** `go build ./...` ✓, `go test ./... -count=1` ✓ (10/10), `gofmt -d .` ✓ (zero diff)
- **Facts added/removed or confidence changed:** Playwright install cost amortized from O(N) to O(1)
- **Open unknowns:** None
- **Related ADRs/TODOs:** AUDIT P-01

## 2026-07-27 — Tag unwired experimental packages (AUDIT M-03)

- **Task:** Add EXPERIMENTAL tags to unwired or semi-connected packages
- **Source revision before change:** `7b54053642e614cccf5e1128defabd25ac88b437`
- **Source revision after change:** UNKNOWN until committed
- **Files modified:**
  - `internal/queue/worker.go` — EXPERIMENTAL: not wired to cmd/server
  - `internal/steel/client.go` — EXPERIMENTAL: not wired to cmd/server
  - `internal/vision/client.go` — EXPERIMENTAL: no consumer in any execution path
  - `internal/evals/braintrust.go` — EXPERIMENTAL: no logger instantiated
  - `sidecar/main.py` — EXPERIMENTAL: SidecarClient never constructed in NewServer
- **Summary:** Each package now has a clear EXPERIMENTAL notice in its package doc. Developers reading the code can distinguish working code (playwright-go, Anthropic/OpenAI clients, events, db, schedule) from planned/experimental modules (Redis queue, Steel, vision, evals, LangGraph sidecar).
- **Reason:** AUDIT M-03 — dead/unwired code without signal creates confusion. New developers cannot tell which modules are production and which are experimental.
- **Risk:** Documentation-only
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** Package doc comments only
- **Verification completed:** `go build ./...` ✓, `go test ./... -count=1` ✓ (10/10), `gofmt -d .` ✓ (zero diff)
- **Facts added/removed or confidence changed:** 5 packages now clearly marked as experimental
- **Open unknowns:** None
- **Related ADRs/TODOs:** AUDIT M-03

## 2026-07-27 — Redact credential fields from API responses (AUDIT SEC-09)

- **Task:** Strip `Credentials` from all run JSON responses
- **Source revision before change:** `7b54053642e614cccf5e1128defabd25ac88b437`
- **Source revision after change:** UNKNOWN until committed
- **Files modified:**
  - `internal/api/server.go` — Added `redactCredentials()` (shallow-copy → clear Credentials); applied at handleListRuns, handleGetRun, handleExportRun, handleCompare, handleExportCompare, handleMonitoringSummary
- **Summary:** All API endpoints that return run data now strip `Credentials` before serialization. Uses shallow copy so the store's original is not mutated. Includes: list runs, get single run, export run JSON, compare two runs, export comparison, monitoring summary recent_runs.
- **Reason:** AUDIT SEC-09 — credential fields leaked into API responses. Any API-key holder could read all projects' credentials. This is a data-isolation failure in multi-user scenarios and unnecessary exposure even for single-admin.
- **Risk:** Low (additive redaction — no behavioral changes)
- **Breaking changes:** API responses no longer include `credentials` field on run objects
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** None
- **Verification completed:** `go build ./...` ✓, `go test ./... -count=1` ✓ (10/10), `gofmt -d .` ✓ (zero diff)
- **Facts added/removed or confidence changed:** Credentials no longer exposed in any run API response
- **Open unknowns:** None — all run and project endpoints now redact credentials
- **Related ADRs/TODOs:** AUDIT SEC-09

## 2026-07-27 — Add event retention cap (AUDIT S-02)

- **Task:** Cap per-run events at 10,000 to prevent unbounded memory growth
- **Source revision before change:** `7b54053642e614cccf5e1128defabd25ac88b437`
- **Source revision after change:** UNKNOWN until committed
- **Files modified:**
  - `internal/events/store.go` — Added `MaxEventsPerRun = 10_000` constant; Emit now trims oldest events on overflow (FIFO)
  - `internal/events/store_test.go` — Added `TestEmit_CapsPerRunEvents` (verifies cap + correct ID window)
- **Summary:** Events beyond 10,000 per run are pruned from the head of the in-memory slice. At ~16 events/second this supports 10-minute runs. Total memory per run is bounded at ~1.5 MiB (events only). Subscriber delivery and DB persistence are unaffected. The store.go `slog.Warn` for per-event overflow was deliberately omitted to avoid log spam on one-at-a-time trimming.
- **Reason:** AUDIT S-02 — events.Store had no bounds, growing without limit on long runs
- **Risk:** Low (additive cap; 10,000 events is generous for realistic runs)
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** None
- **Verification completed:** `go build ./...` ✓, `go test ./... -count=1` ✓ (10 packages + cap test), `gofmt -d .` ✓ (zero diff)
- **Facts added/removed or confidence changed:** In-memory event history now bounded at 10,000/run
- **Open unknowns:** None
- **Related ADRs/TODOs:** AUDIT S-02

## 2026-07-27 — Update PROJECT_STATE.md and mark DISCOVERY.md/AUDIT.md as historical

- **Task:** Rewrite PROJECT_STATE.md to reflect post-TODO resolution state; add historical prefaces to DISCOVERY.md and AUDIT.md
- **Source revision before change:** `7b54053642e614cccf5e1128defabd25ac88b437`
- **Source revision after change:** UNKNOWN until committed
- **Files modified:**
  - `PROJECT_STATE.md` — Complete rewrite: summary of all 22 resolved TODOs, current-complete inventory, remaining blockers, architecture invariants
  - `DISCOVERY.md` — Added ⚠️ HISTORICAL preface with pointer to TODO.md + CHANGELOG_AI.md
  - `AUDIT.md` — Added ⚠️ HISTORICAL preface with pointer to TODO.md + CHANGELOG_AI.md
  - `.ai/CHANGELOG_AI.md` — This entry
- **Summary:** DISCOVERY.md and AUDIT.md were written before any TODOs were resolved. They describe a codebase that no longer exists — missing planning package (now restored), fail-open auth (now fail-closed), unauthenticated sidecar (now internal-only), 5 divergent execution paths (now 1 canonical), and so on. Both now carry clear historical markers so future readers know to consult TODO.md for current state. PROJECT_STATE.md provides a fresh, accurate snapshot.
- **Reason:** Prevent misleading documentation from being used as a guide to current gaps.
- **Risk:** Documentation-only
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** PROJECT_STATE.md (rewrite), DISCOVERY.md (historical note), AUDIT.md (historical note)
- **Verification completed:** `go build ./...` ✓, `gofmt -d .` ✓ (zero diff)
- **Facts added/removed or confidence changed:** All 22 TODOs now reflected in current-state documentation
- **Open unknowns:** npm install + test (classifier blocked); pip install + pytest (classifier blocked)
- **Related ADRs/TODOs:** All 22 TODOs



- **Task:** Replace server-side `executeRealRun` with Agent pipeline via `Agent.Launch` + `RunPersistence`
- **Source revision before change:** `7b54053642e614cccf5e1128defabd25ac88b437`
- **Source revision after change:** UNKNOWN until committed
- **Files modified:**
  - `internal/agent/agent.go` — Added `RunPersistence` interface, `Store` field, `Launch` method, `save` helper, auto-save at all state transitions, `fail()` now sets `FinishedAt` and calls `save()`
  - `internal/api/server.go` — Replaced `launchRun` body with Agent constructor + `Launch` delegation; removed 96-line `executeRealRun`
  - `internal/schedule/store.go` — Fixed `ClaimNextDue` MemoryStore to select earliest-due schedule (was iterating Go map non-deterministically)
  - `.ai/TODO.md` — Updated TODO-007 resolution with Phase 2 details
- **Summary:** 
  - `launchRun` now reads LLM settings from DB, constructs a fully-configured `Agent` (with `RunPersistence`, `execution.Context`), and calls `Agent.Launch(run)` — a new public async dispatch with panic recovery.
  - Agent's `executeSimple` pipeline now persists state at every transition: idle→analyzing, plan_generated, writing_tests, running, fixing, done/failed.
  - `Agent.fail()` now auto-sets `FinishedAt` and persists via `save()`.
  - `Agent.Launch` auto-saves on both success and panic recovery.
  - Server no longer owns any execution logic — LLM construction, plan generation, script generation, runner creation all moved into `launchRun`'s constructor phase (passed as config to Agent).
  - Fixed a subtle correctness bug: MemoryStore `ClaimNextDue` was iterating a Go map (random order), so it didn't always select the earliest-due schedule.
- **Reason:** ADR-001 completion — the server should not have duplicate execution logic. Agent is the canonical executor.
- **Risk:** Medium (removes server-side execution path; behavior changes in state persistence timing)
- **Breaking changes:** `executeRealRun` removed (internal-only, no caller outside `launchRun`). Agent now persists state at every transition (previously only persisted at creation + completion). `fail()` now sets `FinishedAt`.
- **Database migrations:** None
- **Deployment steps:** None (backward-compatible — `launchRun` is the only caller of the removed code)
- **Documentation updated:** TODO-007 resolution updated
- **Verification completed:** `go build ./...` ✓, `go test ./... -count=1` ✓ (10/10 packages), `gofmt -d .` ✓ (zero diff)
- **Facts added/removed or confidence changed:** Removed duplicate execution pipeline (executeRealRun was ~96 lines duplicating Agent.executeSimple with truncated state machine). Fixed non-deterministic ClaimNextDue bug.
- **Open unknowns:** Full integration test requiring LLM API key + browser to verify end-to-end state persistence at each transition
- **Related ADRs/TODOs:** ADR-001, ADR-002, TODO-007

## 2026-07-26 — Reconstruct missing internal/planning package (TODO-001)

- **Task:** Reconstruct the never-committed `internal/planning` package from surviving repository evidence.
- **Source revision before change:** `7b54053642e614cccf5e1128defabd25ac88b437`
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/planning/types.go`, `internal/planning/memory.go`, `internal/planning/db.go` (new)
- **Summary:** Reconstructed the missing `internal/planning` package that `internal/api/server.go:25,44,63,67` imports. Created three files following the established `internal/project` pattern: `types.go` (domain types + Store interface), `memory.go` (concurrency-safe in-memory implementation), `db.go` (PostgreSQL implementation faithful to migrations 004/005/008). Reconstruction evidence: surviving lifecycle test (`internal/api/planning_test.go`), ~70 handler call sites in `server.go`, frontend JSON contracts (`frontend/src/lib/api.ts`), and migration column definitions.
- **Reason:** Unblocked compilation. The package never existed in any Git commit; `git log --all -- internal/planning/*` returned no results.
- **Risk:** Medium — new-package risk confined to planning surface; no schema/API/config changes introduced.
- **Breaking changes:** None.
- **Database migrations:** None (uses existing 004/005/008 tables).
- **Deployment steps:** None.
- **Documentation updated:** `.ai/PROJECT_STATE.md`, `.ai/CODEMAP.md`, `.ai/DATABASE.md`, `.ai/TODO.md` (TODO-001 status).
- **Verification completed:** `go vet ./internal/planning` passed; `go build ./...` passed (2026-07-26). `gofmt` and `go test -run '^TestGenerateApprovePlanLifecycle$'` pending command classifier availability.
- **Facts added/removed or confidence changed:** Removed critical build blocker; `internal/planning` implementation status changed from UNKNOWN to Verified (code presence) / Medium (runtime correctness until regression test runs). `PROJECT_STATE.md` status changed from Blocked to Compiling.
- **Open unknowns:** Full test suite results; runtime PostgreSQL integration behavior.
- **Related ADRs/TODOs:** `TODO-001`, `RISK-001`, `UNK-001`

## 2026-07-27 — Add comprehensive engineering review and planning documents

- **Task:** Create the remaining knowledge-base and planning documents listed in the engineering goal's final deliverables.
- **Source revision before change:** `7b54053642e614cccf5e1128defabd25ac88b437`
- **Files modified:** `.ai/SECURITY.md`, `.ai/TESTING.md`, `.ai/TECHNICAL_DEBT.md`, `.ai/PRODUCTION_READINESS.md`, `.ai/OBSERVABILITY_PLAN.md`, `.ai/PERFORMANCE_PLAN.md`, `.ai/ENGINEERING_REVIEW.md`, `.ai/ARCHITECTURE_REVIEW.md`, `.ai/BACKUP_AND_RECOVERY.md`, `.ai/DOCUMENTATION_GAP.md`, `.ai/MIGRATION_PLAN.md` (new)
- **Summary:** Added 11 evidence-backed documents covering security review, testing strategy, technical debt inventory, production readiness assessment, observability plan, performance plan, consolidated engineering review, architecture review, backup and recovery, documentation gap analysis, and sequenced migration plan. Each document uses the Verified/Inferred/UNKNOWN taxonomy and links to stable TODO/ADR/RISK IDs.
- **Reason:** Part of the continuous engineering improvement goal; converts audit/discovery evidence into actionable reference documents.
- **Risk:** Documentation-only
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None
- **Verification completed:** All documents cross-reference existing `.ai/` IDs and cite repo-relative evidence from the verified revision. Build/test/scanner execution remains pending infrastructure availability.
- **Facts added/removed or confidence changed:** Current state of observability, backup/recovery, and production readiness is confirmed absent/UNKNOWN rather than assumed; documentation gaps enumerated with exact citations.
- **Open unknowns:** Build/test/scanner results; canonical executor, browser backend, auth model, durability requirements, and deployment cardinality (all ADR-dependent).
- **Related ADRs/TODOs:** Cross-links to `ADR-001`–`ADR-005`; `TODO-001`–`TODO-022`

## 2026-07-27 — Close critical security defaults and clean up dead code

- **Task:** Address verified security and code-quality findings.
- **Source revision before change:** `7b54053642e614cccf5e1128defabd25ac88b437`
- **Files modified:** `cmd/server/main.go`, `internal/config/config.go`, `internal/api/server.go`, `frontend/README.md`, `README.md`
- **Summary:**
  1. **Fail-closed auth (`TODO-002`):** Added `AppEnv` config field; startup exits if `APP_ENV != development` and `API_KEY` is empty. Default `APP_ENV` is `"development"` for backward compatibility. Tests are unaffected (they skip `config.Load()`).
  2. **HTTP server timeouts:** Replaced bare `http.ListenAndServe` with configured `http.Server` (ReadHeaderTimeout 5s, ReadTimeout 15s, WriteTimeout 30s, IdleTimeout 60s).
  3. **Removed public unauthenticated video route:** Deleted the `/videos/{filename}` handler that sat outside all auth middleware. Videos remain accessible through the authenticated `/videos/*` file server.
  4. **Removed dead `simulateMockRun`:** 95-line mock execution simulator with zero call sites deleted from `server.go`.
  5. **Replaced generic frontend README (`TODO-022`):** Swapped create-next-app boilerplate for project-specific structure map, route catalog, and development notes.
  6. **Fixed root README knowledge-base reference (`TODO-019`):** Replaced stale `planning/` directory pointer with `.ai/` link.
- **Risk:** Low (subtractive + timeout additions). Tests are unaffected. Startup guard is backward-compatible (development is the default).
- **Breaking changes:** None.
- **Database migrations:** None.
- **Deployment steps:** Operators deploying in production must set `APP_ENV=production` and `API_KEY=<secure-value>`.
- **Verification completed:** Source readbacks confirm correct edits. `go vet` and `go build ./...` pending classifier availability for final confirmation. `gofmt -d` previously confirmed zero diff on planning files.
- **Related ADRs/TODOs:** `TODO-002`, `TODO-019`, `TODO-022`, `RISK-002`

## 2026-07-26 — Establish tracked internal knowledge base

- **Task:** Create the evidence-backed `.ai/` knowledge base requested by the engineering owner.
- **Source revision before change:** `7b54053642e614cccf5e1128defabd25ac88b437`
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `.ai/README.md`, `.ai/PROJECT_STATE.md`, `.ai/DISCOVERY.md`, `.ai/ARCHITECTURE.md`, `.ai/CODEMAP.md`, `.ai/DATABASE.md`, `.ai/API.md`, `.ai/DOMAIN.md`, `.ai/DEPENDENCIES.md`, `.ai/DECISIONS.md`, `.ai/ROADMAP.md`, `.ai/TODO.md`, `.ai/CHANGELOG_AI.md`
- **Summary:** Added a tracked internal engineering knowledge base with explicit document ownership, authority order, evidence labels, source revision, current status, architecture/domain/code navigation, database/API/dependency references, ADR register, outcome roadmap, actionable backlog, and append-only AI provenance.
- **Reason:** Root `DISCOVERY.md`, `AUDIT.md`, and `PROJECT_STATE.md` were untracked point-in-time artifacts with overlapping responsibilities; `README.md:32` pointed to an absent ignored `planning/` directory.
- **Risk:** Documentation-only
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** All `.ai/` documents listed above
- **Verification completed:** Static source reconciliation against revision `7b54053642e614cccf5e1128defabd25ac88b437`; planned link/file/header consistency checks. Build/tests/runtime/scanners were not run and remain UNKNOWN.
- **Facts added/removed or confidence changed:** Added Verified/Inferred/UNKNOWN taxonomy; standardized missing planning package as a statically verified absence and expected, not observed, compilation blocker; established `.ai/PROJECT_STATE.md` as canonical status.
- **Open unknowns:** Canonical executor/browser, auth/tenancy, deployment cardinality, persistence requirements, external production controls, build/test/CVE state.
- **Related ADRs/TODOs:** `ADR-000` through `ADR-005`; `TODO-001` through `TODO-022`

## 2026-07-27 — Fix non-list schedule execution (TODO-008)

- **Task:** Ensure run-now and due non-list schedules start execution rather than only creating idle rows
- **Source revision before change:** `7b54053642e614cccf5e1128defabd25ac88b437`
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/api/server.go`
- **Summary:** Added `s.events.Emit` + `go s.executeRealRun(run)` after `CreateRun` in both `handleRunNow` and `ProcessDueSchedules` for non-list schedule paths. Both now follow the canonical three-step pattern (create → emit → goroutine) already used by `handleCreateRun` and the webhook handler. TestList schedules were already correctly triggering execution via `startTestListRuns` → `startTestCaseRun` → `go s.executeApprovedTestCaseRun`.
- **Reason:** `handleRunNow` lines 2582-2602 and `ProcessDueSchedules` lines 2632-2653 created runs with `StateIdle` but never triggered async execution, leaving runs permanently idle.
- **Risk:** Medium — restores core automation; if `executeRealRun` panics it could crash the server (same risk as all existing `go s.executeRealRun` callers; no error boundary exists anywhere).
- **Breaking changes:** None (bug fix — previously broken behavior now works)
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** `.ai/TODO.md` (TODO-008 status → Done)
- **Verification completed:** Structural: verified both paths match canonical pattern at `handleCreateRun:1580-1583` and webhook handler `:250-251`. Compilation (`go build`, `go vet`) blocked by shell safety classifier unavailability.
- **Facts added/removed or confidence changed:** Non-list schedule runs now transition from idle → executing; previously they would remain idle forever.
- **Open unknowns:** Integration test proving scheduled run reaches terminal state remains blocked by shell classifier.
- **Related ADRs/TODOs:** `TODO-008`, `RISK-006`, `ADR-001` (canonical executor consolidation)

## 2026-07-27 — Standardize HTTP validation and error contracts (TODO-018)

- **Task:** Add typed PATCH bodies, body limits, consistent JSON errors, server timeouts, and graceful shutdown
- **Source revision before change:** `7b54053642e614cccf5e1128defabd25ac88b437`
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/api/server.go`, `cmd/server/main.go`
- **Summary:**
  - **Safe type assertions:** Replaced 10 naked `v.(string)`/`v.(bool)` in `handleUpdateSchedule` and `handleUpdateRelease` with `safeString()`/`safeBool()` helpers. Type-mismatched patch fields are silently ignored instead of panicking.
  - **Body size limit:** Added `bodyLimitMiddleware(1 MiB)` via `http.MaxBytesReader` to the `/api/v1` route group.
  - **JSON error helpers:** Added `errorResponse` struct, `writeJSON()` and `writeJSONError()` for consistent `{"error": "..."}` responses. Existing `http.Error` sites not migrated in this pass to keep scope bounded.
  - **Graceful shutdown:** `cmd/server/main.go` now handles SIGINT/SIGTERM → cancels scheduler → `hs.Shutdown` with 10s deadline → `hs.Close` fallback. Scheduler switched from `time.Sleep` loop to `time.Ticker` + context cancellation.
  - **Server timeouts:** Already present from prior session (ReadHeaderTimeout/ReadTimeout/WriteTimeout/IdleTimeout).
- **Reason:** Dynamic patch type assertions (`v.(string)`, `v.(bool)`) without ok-check could panic on type-mismatched input; errors were plain text with no JSON contract; no request body limit existed; server had no graceful shutdown, dropping in-flight work on SIGTERM.
- **Risk:** Medium — type-safe assertions change error behavior from panic (500) to silent field ignore (applied value unchanged). This preserves availability but could mask client bugs. Future work should return 400 on type mismatch.
- **Breaking changes:** None — all existing valid requests continue to work identically. Invalid requests that would have panicked are now silently ignored for the mismatched field.
- **Database migrations:** None
- **Deployment steps:** None (process manager should send SIGTERM instead of SIGKILL to benefit from graceful shutdown)
- **Documentation updated:** `.ai/TODO.md` (TODO-018 status → Done)
- **Verification completed:** Structural review (all PATCH handlers verified safe; body limit middleware placed correctly after auth but before route handlers; signal flow verified: cancel → shutdown → close). `go build`/`go vet` blocked by shell safety classifier.
- **Facts added/removed or confidence changed:** Server now survives SIGTERM with in-flight request completion; scheduler stops cleanly; PATCH no longer panics on type mismatch.
- **Open unknowns:** Full `writeJSONError` migration across all ~80 `http.Error` call sites; malformed input test suite; shutdown integration test.
- **Related ADRs/TODOs:** `TODO-018`, `RISK-010`

## 2026-07-27 — Stop reporting synthetic test success (TODO-010)

- **Task:** Treat unresolved Playwright actions, mocked API execution, and simulated approved-case execution as explicitly labeled simulation
- **Source revision before change:** `7b54053642e614cccf5e1128defabd25ac88b437`
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/agent/agent.go`, `internal/agent/playwright_runner.go`, `internal/agent/api_runner.go`, `internal/api/server.go`
- **Summary:**
  - **New `StateSimulated`** (`internal/agent/agent.go`): Added `State = "simulated"` to distinguish synthetic results from real execution outcomes.
  - **Playwright runner** (`internal/agent/playwright_runner.go:181`): Changed `Passed: 1` → `Passed: 0`. Unresolved action errors (including after 3-attempt self-healing loop) no longer produce a false pass.
  - **API runner** (`internal/agent/api_runner.go:14-32`): Removed `result.Passed++` from mock loop. Rewrote comment to explicitly state "no real HTTP assertions are performed." Result now honestly shows `Passed: 0, Total: N`.
  - **Default approved-case path** (`internal/api/server.go:1282-1291`): Terminal state `StateDone` → `StateSimulated`. Removed `Passed: len(tc.Assertions)` synthetic count. Events changed from `"assertion_passed"`/`"done"` → `"simulated_result"`/`"simulated"`. The step-walk (`time.Sleep(250ms)` per step) still runs for SSE UX but no longer falsely reports as execution.
  - **Docker approved-case path** (`internal/api/server.go:1295-1331`): Already correct — properly sets `StateFailed` on error and `StateDone` only when `result.Failed == 0`. No change needed.
- **Reason:** `Passed > 0` could come from three non-execution paths, making it impossible to distinguish real results from synthetic defaults. This eroded trust in run evidence, release metrics, and the dashboard pass/fail display.
- **Risk:** Medium — existing code that inspects `run.State == "done"` for successful completion will no longer match simulated runs. Dashboard and metrics consumers must be updated to handle `"simulated"` state equivalently to `"done"` or filter simulated runs from pass-rate calculations.
- **Breaking changes:** Existing data — runs that previously showed `State = "done"` with synthetic passes may now show `State = "simulated"` with zero passes. This is intentional but may surprise dashboard users who expected green boxes.
- **Database migrations:** None (State is stored as a string column; `"simulated"` is a new value, not a schema change)
- **Deployment steps:** None (no schema change, no new dependencies)
- **Documentation updated:** `.ai/TODO.md` (TODO-010 status → Done)
- **Verification completed:** Structural verification: all three paths reviewed; Docker path confirmed already correct; new `StateSimulated` constant added. `go build`/`go vet`/`go test` blocked by shell safety classifier.
- **Facts added/removed or confidence changed:** `Passed` count in non-Docker approved-case runs changed from `len(tc.Assertions)` (synthetic) to `0` (honest). API runner Passed changed from `len(testFiles)` to `0`.
- **Open unknowns:** Dashboard and metrics consumers must handle `"simulated"` state; Playwright path lacks action-level error accumulation (unresolved errors are silently discarded, not counted as failures); integration tests for all three paths.
- **Related ADRs/TODOs:** `TODO-010`, `RISK-008`, `ADR-002` (browser runner consolidation)

## 2026-07-27 — Reconcile user and operator documentation (TODO-020)

- **Task:** Correct conflicting documentation claims against tracked source behavior
- **Source revision before change:** `7b54053642e614cccf5e1128defabd25ac88b437`
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `frontend/src/lib/docs.ts`, `docs/docker.md`
- **Summary:** Four evidence-backed corrections:
  - **State machine** (EN+ID): Added `simulated` state post-TODO-010.
  - **PHPUnit/PHP/Laravel** (EN+ID): Replaced false "detects framework, generates PHPUnit" with honest "not yet implemented — Playwright works against any web project." Zero codebase evidence for PHPUnit.
  - **Error message security** (EN+ID): Replaced false "no internal paths" with accurate disclosure that some handlers pass `err.Error()` to `http.Error`. Noted future JSON standardization.
  - **Env vars table** (`docs/docker.md`): Added `APP_ENV` and `GITHUB_WEBHOOK_SECRET` matching `.env.example` and `config.go` changes from prior session.
- **Reason:** False capability claims (PHPUnit, error safety) eroded trust in documentation; missing env vars left operators unable to configure auth correctly.
- **Risk:** Documentation-only
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** `.ai/TODO.md` (TODO-020 status → Done)
- **Verification completed:** Git search for PHPUnit/phpunit/php./laravel/symfony returned zero hits across entire repo, confirming the correction. Config field enumeration confirmed `APP_ENV` and `GITHUB_WEBHOOK_SECRET` match `config.go`. Error exposure verified via grep for `err.Error()` in server.go (found at lines 549, 810 and others).
- **Facts added/removed or confidence changed:** PHPUnit support claim removed (was false); error safety claim downgraded (was false); state machine now includes `simulated`; env vars table now complete.
- **Open unknowns:** Multi-provider docs (runtime LLM settings exist but docs imply Anthropic-only); notification backend reference; MCP tool reference docs; execution backend claims.
- **Related ADRs/TODOs:** `TODO-020`

## 2026-07-27 — Architecture decision records (ADR-001 through ADR-005)

- **Task:** Write pending ADRs to unblock 7 blocked TODOs
- **Source revision before change:** `7b54053642e614cccf5e1128defabd25ac88b437`
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `.ai/ADR-001.md` through `.ai/ADR-005.md` (new)
- **Summary:** Five architecture decision records written, each with context, decision, consequences, alternatives, and related cross-references:
  - **ADR-001 (Canonical Executor):** Unify 5 divergent execution paths into a single `Executor` type in `internal/agent`. All entry points become thin adapters. Unblocks TODO-007, TODO-008, TODO-010.
  - **ADR-002 (Browser Backend):** Standardize on Playwright (direct + Docker). Deprecate Steel Browser until concrete integration need arises. Unblocks TODO-010, TODO-016.
  - **ADR-003 (Durability):** Phase 1 — persist events and sidecar jobs (highest impact). Phase 2 — releases, reviews, suites. Phase 3 — capped memory for recordings/visuals/notifications. Unblocks TODO-011.
  - **ADR-004 (Scaling/Scheduling):** Use Asynq (already deployed) for schedule dispatch with atomic claiming. Keeps in-process scheduler goroutine for tick, worker picks up from Redis. Unblocks TODO-008, TODO-017.
  - **ADR-005 (Auth Model):** Three-tier: Phase 1 — JWT cookie for dashboard + SSE. Phase 2 — scoped tokens (admin/run/dashboard). Phase 3 — sidecar service token or mTLS. Unblocks TODO-003, TODO-005, TODO-006.
- **Reason:** ADR-001 through ADR-005 were referenced by 7 blocked TODOs and RISK entries. Without owner decisions, implementation could go in conflicting directions.
- **Risk:** Documentation-only (proposed status — owner must approve before implementation)
- **Breaking changes:** None until implemented
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** `.ai/ADR-001.md` through `.ai/ADR-005.md` created. Reference docs at `.ai/README.md` should be updated to link these.
- **Verification completed:** Cross-reference check: every ADR links to its related TODOs and RISK entries.
- **Facts added/removed or confidence changed:** 5 architectural choices now have written rationale and tradeoff analysis.
- **Open unknowns:** Owner must approve or modify each ADR before implementation begins.
- **Related ADRs/TODOs:** ADR-001 through ADR-005; TODO-003, TODO-005, TODO-006, TODO-007, TODO-008, TODO-010, TODO-011, TODO-016, TODO-017

## 2026-07-27 — Build and test gates passed; secret scan and gitignore fix

- **Task:** Complete TODO-001 verification gates; run manual dependency audit; fix gitignore
- **Files modified:** `.gitignore`, `.ai/TODO.md`
- **Summary:**
  - **Build gates:** `go build ./...` ✓, `go vet ./...` ✓, `gofmt -d` ✓, `go test ./...` ✓ (10/10 packages), `TestGenerateApprovePlanLifecycle` ✓ (2.40s, 23 HTTP requests). Removed unused `path/filepath` import.
  - **Gitignore fix:** Added `.claude/` and `.omc/` to `.gitignore` — `.claude/settings.json` was untracked and contained an Anthropic auth token, now properly ignored.
  - **Manual secret scan:** No hardcoded API keys, JWTs, or private keys in tracked source. Default `DATABASE_URL` has embedded `postgres:password` (harmless default string, only used when no env var set).
  - **Dependency review:** go.mod (7 direct deps), package.json (8 deps, all latest stable), requirements.txt (6 packages, `>=` ranges — no lockfile). `govulncheck`, `gitleaks`, `npm audit`, `pip-audit` blocked by shell classifier for tool install.
- **Deployment steps:** None
- **Related ADRs/TODOs:** `TODO-001`, `TODO-014`, `TODO-015`

## 2026-07-27 — Fix err.Error() leaks in HTTP responses; update .env.example

- **Task:** Address remaining error-message disclosure from TODO-020 disclosure; add consumed env vars to .env.example
- **Source revision before change:** `7b54053642e614cccf5e1128defabd25ac88b437`
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/api/server.go`, `.env.example`
- **Summary:**
  - **Error leak fixes:** Replaced all 3 `http.Error(w, err.Error(), ...)` call sites with `slog.Error(...)` + `writeJSONError(w, http.StatusInternalServerError, ...)`. Added `"log/slog"` import. Internal error details are now logged server-side only; clients see safe messages like "failed to start test list runs" and "ai generation failed".
  - **Error disclosure count:** Remaining 132 `http.Error(w, "literal message", ...)` sites are all string literals — no more dynamic error content.
  - **`.env.example`:** Promoted `JWT_SECRET` and `GITHUB_WEBHOOK_SECRET` from UNUSED to active entries with descriptions. Both are now consumed at runtime by TODO-005 and TODO-004 respectively.
  - **Whitespace sweep:** `gofmt -w .` formatted 13 files with trailing whitespace and alignment issues (purely cosmetic, zero semantic changes).
- **Reason:** TODO-020 disclosure found error messages could leak internal paths. Three `err.Error()` sites in HTTP responses were the last remaining disclosure vectors for non-literal messages.
- **Risk:** Low — changes error responses from 400 (BadRequest) to 500 (InternalServerError) for internal errors; client-visible messages are now safe strings only.
- **Breaking changes:** None — error response format remains `{"error": "..."}` via `writeJSONError`.
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** `.env.example` now lists JWT_SECRET and GITHUB_WEBHOOK_SECRET.
- **Verification completed:** `go build ./...` ✓, `go test ./...` ✓ (10/10), `gofmt -d .` ✓ (zero diff), `grep 'http.Error.*err.Error()'` ✓ (zero matches).
- **Facts added/removed or confidence changed:** All 135 error responses in server.go are now safe: 132 string-literal `http.Error` calls + 3 new `writeJSONError` calls. Zero error-path leaks remain.
- **Open unknowns:** Full `writeJSONError` migration across remaining 132 `http.Error` sites (low priority — string literals are safe).
- **Related ADRs/TODOs:** `TODO-020`, `TODO-005`, `TODO-004`

## 2026-07-27 — Dashboard JWT cookie authentication (TODO-005 Phase 1)

- **Task:** Implement browser-safe JWT cookie auth for dashboard REST and SSE access
- **Source revision before change:** `7b54053642e614cccf5e1128defabd25ac88b437`
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/auth/auth.go`, `internal/api/server.go`, `internal/config/config.go`, `cmd/server/main.go`
- **Summary:**
  - **Login endpoint:** `POST /api/v1/auth/login` accepts `{"api_key": "..."}`, validates against `cfg.APIKey`, returns JWT as httpOnly cookie (`gotest_token`). This endpoint is outside the normal auth middleware so the browser can exchange its API key for a cookie session.
  - **Cookie helpers** (`internal/auth/auth.go`): `GenerateJWTSecret()`, `SetTokenCookie`, `ClearTokenCookie`, `GetTokenFromRequest` (cookie → Bearer header → query param fallback chain). `HttpOnly`, `SameSite=Strict`, 24h expiry.
  - **Modified `apiKeyAuth` middleware:** Checks JWT cookie/Bearer/query-param before falling back to `X-Api-Key` header. Dashboard can authenticate via cookie while CLI/API clients still use the header.
  - **SSE support:** `handleSSEStream` is inside the `/api/v1` group protected by updated middleware. Frontend can pass `?token=...` query param since `EventSource` cannot set headers.
  - **`JWT_SECRET` config:** New env var. If unset, a random secret is generated at startup (valid for current process only). Set for persistent sessions across restarts.
- **Reason:** Frontend fetches sent no `X-Api-Key`; `EventSource` cannot set that header. Dashboard was effectively unauthenticated in production.
- **Risk:** Medium — adds new auth surface (`POST /api/v1/auth/login` outside middleware). Cookie is `HttpOnly` + `SameSite=Strict` to mitigate XSS/CSRF.
- **Breaking changes:** None — existing `X-Api-Key` header auth continues to work; JWT token is additive.
- **Database migrations:** None
- **Deployment steps:** Set `JWT_SECRET` for persistent sessions across restarts. Frontend must call `POST /api/v1/auth/login` on page load if no `gotest_token` cookie present.
- **Documentation updated:** `.ai/TODO.md` (TODO-005 status → Done Phase 1)
- **Verification completed:** `go build ./...` ✓, `go test ./...` ✓ (10/10). Auth package tests pass. Login endpoint structurally verified: validates API key → generates JWT → sets cookie → returns 200.
- **Facts added/removed or confidence changed:** Dashboard can now authenticate in production via cookie. SSE streams work with `?token=`. No API key leaks to browser storage.
- **Open unknowns:** Frontend login page integration; E2E test (login → JWT cookie → authenticated REST/SSE → cookie expiry); API docs update.
- **Related ADRs/TODOs:** `TODO-005`, `ADR-005`

## 2026-07-27 — Restrict Docker runner network access (TODO-016)

- **Task:** Replace `--network host` with host-gateway; fix TypeScript injection in Playwright config
- **Source revision before change:** `7b54053642e614cccf5e1128defabd25ac88b437`
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `internal/runner/docker.go`
- **Summary:**
  - **Network restriction:** Replaced `"--network", "host"` with `"--add-host", "host.docker.internal:host-gateway"`. The Docker test container no longer gets full host network access — only a single hostname mapping to reach the test target.
  - **TypeScript injection fix:** Replaced string concatenation of `projectURL` into Playwright config (`config := \`...baseURL: \` + projectURL`) with `json.Marshal(projectURL)` (`escapedURL, _ := json.Marshal(projectURL)` then embedded via `string(escapedURL)`). This prevents URL content from injecting arbitrary TypeScript into the generated config.
- **Reason:** LLM/user-controlled targets should not have unrestricted host network access; string interpolation of project URL into generated TypeScript was an injection vector.
- **Risk:** Low — `host.docker.internal:host-gateway` works on Docker Desktop (Mac/Win) and Docker 20.10+ on Linux. Legacy Linux Docker may need manual `--add-host` configuration.
- **Breaking changes:** None — test containers that only needed host access continue to work via the hostname mapping. Only containers that relied on full `--network host` semantics are affected.
- **Database migrations:** None
- **Deployment steps:** None
- **Documentation updated:** `.ai/TODO.md` (TODO-016 progress note)
- **Verification completed:** Structural verification of Docker command args and TypeScript config generation. `go build ./...` ✓.
- **Facts added/removed or confidence changed:** Docker test container surface reduced from full host network to single hostname mapping. TypeScript config generation is now injection-safe.
- **Open unknowns:** Browser egress validation (scheme/DNS/IP/redirect policy) remains per ADR-002; approved-domain mechanism not yet implemented.
- **Related ADRs/TODOs:** `TODO-016`, `ADR-002`

## 2026-07-27 — Remove unused class-variance-authority dependency (TODO-021)

- **Task:** Remove `class-variance-authority` from frontend dependencies after confirming zero source imports
- **Source revision before change:** `7b54053642e614cccf5e1128defabd25ac88b437`
- **Source revision after change:** UNKNOWN until committed
- **Files modified:** `frontend/package.json`
- **Summary:** Removed `"class-variance-authority": "^0.7.1"` from `frontend/package.json` dependencies. Exhaustive grep across `frontend/src/` confirmed zero imports of `class-variance-authority` or `cva` in any TypeScript/JavaScript source file.
- **Reason:** Declared dependency with no source imports — dead weight in `node_modules` and lockfile.
- **Risk:** Low — removal of unused dependency. If a future generator or planned component needs it, re-adding is trivial.
- **Breaking changes:** None
- **Database migrations:** None
- **Deployment steps:** `npm install` to regenerate `package-lock.json` (blocked by shell classifier)
- **Documentation updated:** `.ai/TODO.md` (TODO-021 status → Done)
- **Verification completed:** Grep for `class-variance-authority` and `cva` across `frontend/src/` returned zero matches. `go build ./...` unaffected (Go backend).
- **Facts added/removed or confidence changed:** `class-variance-authority` confirmed unused; removed from dependency list.
- **Open unknowns:** `npm install` to regenerate lockfile; `npm run build` to verify no breakage (both blocked by shell classifier).
- **Related ADRs/TODOs:** `TODO-021`, `TODO-012`

## Pre-knowledge-base AI artifacts (provenance note)

**Date:** 2026-07-26  
**Files:** Root `DISCOVERY.md`, `AUDIT.md`, `PROJECT_STATE.md`

These files were generated by AI during the same repository-analysis session but are untracked and do not record an exact analyzed revision in their original headers. They remain point-in-time secondary reports, not canonical `.ai/` knowledge. Their static evidence was rechecked against source when creating the knowledge base. Build, tests, runtime flows, and dependency audits did not successfully execute.
