# Documentation Gap Analysis

**Owner:** Engineering  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437` (original audit); reconciled 2026-07-28 against working tree  
**Last updated:** 2026-07-28  
**Verification performed:** Static inspection of all tracked documentation against tracked source  
**Confidence:** High for identified conflicts; UNKNOWN for external documentation not in this repository

## Primary Source/Documentation Conflicts

These are cases where tracked documentation makes claims that conflict with tracked implementation.

| ID | Document | Claim | Reality | Evidence |
|---|---|---|---|---|
| DG-1 | `README.md:32` | ~~`planning/` reference~~ ✅ Resolved: README no longer references `planning/` (grep 2026-07-28: zero matches) | — | — |
| DG-2 | `frontend/src/lib/docs.ts:31,55` | Tests execute in "Steel Browser" | Primary web run uses local Playwright; Steel unwired (UW-2, ADR-002 pending) | Open — depends on UW-2 product decision |
| DG-3 | `frontend/src/lib/docs.ts:396,417` | ~~Webhook has no production caller~~ ✅ Resolved: `StartFailureNotifier` calls `TriggerFailure` on `run_failed` (UW-6, 2026-07-27) | — | `internal/api/failure_notifier.go:52` |
| DG-4 | `frontend/src/lib/docs.ts:615,638` | ~~"PHPUnit for unit tests"~~ ✅ Resolved: docs.ts:617,640 now explicitly state PHPUnit is NOT implemented and the agent generates Playwright | — | `frontend/src/lib/docs.ts:617,640` |
| DG-5 | `docs/docker.md:26-33` | ~~Steel port wrong~~ ✅ Resolved: docker.md services table lists steel-browser at 3010 (container port 3000) | — | `docs/docker.md` services table |
| DG-6 | `docs/docker.md:59-66` | Anthropic described as required; model examples reference older defaults | Multi-provider routing now consistent (DL-2 resolved 2026-07-28) but docker.md wording predates it | Open — docs.md refresh |
| DG-7 | `.env.example:4-5,32-33` | ~~`JWT_SECRET`/`GITHUB_WEBHOOK_SECRET` unused~~ ✅ Resolved: both consumed (`internal/api/server.go:69,288`) | — | grep 2026-07-28 |
| DG-8 | `.env.example` | `VISION_MODEL`, `BRAINTRUST_API_KEY`, `ENABLE_VISUAL_REGRESSION`, `ENABLE_ADVANCED_AGENT` documented | Now correctly listed under UNUSED marker in `.env.example`; runtime consumers still absent (UW-4/UW-5 product decision) | `.env.example` UNUSED section |
| DG-9 | `.env.example:39-41` | ~~`MAX_FIX_ATTEMPTS`/`DEFAULT_TIMEOUT_SECONDS` ignored~~ ✅ Resolved: read via `getEnvInt` (UC-2/UC-3, commit 5e7175d) | — | `internal/config/config.go` |
| DG-10 | `README.md:80-84` | Google and DeepSeek listed as supported providers | ✅ Substantially resolved: both layers now route google/deepseek to OpenAI-compatible transport with correct default endpoints (DL-2 Phase 1, commit 3b56bc9); `GOOGLE_API_KEY`/`DEEPSEEK_API_KEY` env names remain README-only (use `LLM_API_KEY`) | `internal/ai/client.go`; UC-1 |
| DG-11 | `frontend/README.md:1-36` | Generic create-next-app template | Unchanged — cosmetic (ST-1) | Open |
| DG-12 | `README.md:201-204` | ~~`CORS_ALLOWED_ORIGINS` recommended but hard-coded wildcard~~ ✅ Resolved (2026-07-28): `CORS_ALLOWED_ORIGINS` implemented — comma-separated allowlist, origin echo + `Vary: Origin`, wildcard only when unset/`*`; 5 middleware tests | — | `internal/api/server.go` (`newCORSMiddleware`); `internal/api/cors_test.go` |

## Missing Documentation

Areas where no tracked documentation exists but implementation requires it:

| ID | Missing | Importance | Notes |
|---|---|---|---|
| MG-1 | `GOTEST_AI_PLANNING` usage | HIGH | Controls whether AI feature extraction is used; absent from `.env.example` and Compose |
| MG-2 | `GOTEST_APPROVED_CASE_RUNNER=docker` | HIGH | Required for real approved-case execution; default path synthesizes passes |
| MG-3 | `LLM_PROVIDER`, `LLM_API_KEY`, `LLM_BASE_URL` | HIGH | Code-consumed multi-provider settings; absent from `.env.example` and Compose |
| MG-4 | `GOOGLE_API_KEY`, `DEEPSEEK_API_KEY` consumers | MEDIUM | README mentions them but no code reads them from environment |
| MG-5 | Frontend authentication mechanism | HIGH | No documentation on how dashboard should authenticate when `API_KEY` is set |
| MG-6 | Production deployment guide | HIGH | Only local/demo Compose is documented |
| MG-7 | Run state machine per entry point | MEDIUM | Different web/Agent/approved-case/sidecar states are not documented |
| MG-8 | Artifact lifecycle (videos, screenshots, events) | MEDIUM | Retention, cleanup, and storage paths vary by runner |
| MG-9 | Provider routing behavior | HIGH | Which providers actually work in real execution vs. test connection |
| MG-10 | API pagination and rate limiting | MEDIUM | README says rate limiting is not implemented; no pagination docs |
| MG-11 | Migration and rollback procedures | HIGH | No documentation on safe migration application or recovery |
| MG-12 | Backup and restore procedures | HIGH | Only destructive reset documented (`docs/docker.md:92-101`) |

## Stale Documentation

| ID | Document | Issue | Recommendation |
|---|---|---|---|
| ST-1 | `frontend/README.md` | Entirely generic create-next-app template | Replace with project-specific frontend guidance |
| ST-2 | `docs/docker.md` | Predates several configuration additions; Steel port is wrong | Update with current Compose topology |
| ST-3 | `README.md` Stack section | Lists `claude-sonnet-4-5` as default; `.env.example` and Compose still reference this | Update to reflect current multi-provider setup |

## Environment Variable Documentation Drift

Reconciled 2026-07-28. Remaining drift only:

| Variable | Status |
|---|---|
| `GOOGLE_API_KEY` / `DEEPSEEK_API_KEY` | README-only names; actual config uses `LLM_API_KEY` (UC-1) |
| `BRAINTRUST_API_KEY`, `VISION_MODEL`, `ENABLE_VISUAL_REGRESSION`, `ENABLE_ADVANCED_AGENT` | Correctly marked UNUSED in `.env.example`; consumers pending UW-4/UW-5 decisions |
| `ANTHROPIC_API_KEY` | Web run prefers DB settings over env value — documented behavior, worth a note in OPERATIONS.md |

Previously-flagged drift now resolved: `LLM_PROVIDER`/`LLM_API_KEY`/`LLM_BASE_URL`, `GOTEST_AI_PLANNING`, `GOTEST_APPROVED_CASE_RUNNER` documented in `.env.example`; `JWT_SECRET` and `GITHUB_WEBHOOK_SECRET` consumed; `MAX_FIX_ATTEMPTS`/`DEFAULT_TIMEOUT_SECONDS`/`STEEL_MAX_SESSIONS` read from env (UC-2/UC-3); `CORS_ALLOWED_ORIGINS` implemented and documented (DG-12).

## Priority Order for Documentation Fixes

1. **Critical:** Add missing code-consumed config to `.env.example` and Compose (`LLM_PROVIDER`, `LLM_API_KEY`, `LLM_BASE_URL`, `GOTEST_AI_PLANNING`, `GOTEST_APPROVED_CASE_RUNNER`).
2. **High:** Remove or mark unused `.env.example` variables (`JWT_SECRET`, `GITHUB_WEBHOOK_SECRET`, `BRAINTRUST_API_KEY`, `VISION_MODEL`, `GOOGLE_API_KEY`, `DEEPSEEK_API_KEY`).
3. **High:** Update embedded docs to match active execution paths (`TODO-020`).
4. **High:** Replace `README.md:32` `planning/` reference with `.ai/` (`TODO-019`).
5. **Medium:** Update `docs/docker.md` Steel port and model defaults.
6. **Medium:** Replace `frontend/README.md` with project-specific content (`TODO-022`).
7. **Low:** Remove or mark stale `.env.example` values that code ignores (`MAX_FIX_ATTEMPTS`, etc.).
