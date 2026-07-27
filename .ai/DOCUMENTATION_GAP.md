# Documentation Gap Analysis

**Owner:** Engineering  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-27  
**Verification performed:** Static inspection of all tracked documentation against tracked source  
**Confidence:** High for identified conflicts; UNKNOWN for external documentation not in this repository

## Primary Source/Documentation Conflicts

These are cases where tracked documentation makes claims that conflict with tracked implementation.

| ID | Document | Claim | Reality | Evidence |
|---|---|---|---|---|
| DG-1 | `README.md:32` | Architecture blueprints, PRDs, and master prompts are in `planning/` | `planning/` is gitignored (`.gitignore:72-73`) and absent | No tracked `planning/` exists in any commit |
| DG-2 | `frontend/src/lib/docs.ts:31,55` | Tests execute in "Steel Browser" | Primary web run uses local Playwright (`internal/api/server.go:1922-1929`) | Steel client and runner exist but are not wired |
| DG-3 | `frontend/src/lib/docs.ts:396,417` | "Webhook — get notified on failure via Slack/Telegram" | `TriggerFailure` has no production caller | `internal/notify/store.go:58-99` |
| DG-4 | `frontend/src/lib/docs.ts:615,638` | "PHPUnit for unit tests" | Both generators (Anthropic and OpenAI) request Playwright | `internal/agent/llm_anthropic.go:60-97`; `internal/agent/llm_openai.go:152-181` |
| DG-5 | `docs/docker.md:26-33` | Steel listed as port 3000 with step-by-step walkthrough | Compose publishes `3010:3000` (host:container) | `docker-compose.yml:92-93` |
| DG-6 | `docs/docker.md:59-66` | Anthropic described as required; model examples reference older defaults | Multi-provider architecture exists but routing is inconsistent | `internal/api/server.go:1889-1894`; `README.md:76-88` |
| DG-7 | `.env.example:4-5,32-33` | `JWT_SECRET` and `GITHUB_WEBHOOK_SECRET` documented | `JWT_SECRET` has no code consumer; `GITHUB_WEBHOOK_SECRET` is ignored — runtime reuses `API_KEY` | `internal/config/config.go:7-41`; `internal/api/server.go:237` |
| DG-8 | `.env.example:29-33,39-43` | `VISION_MODEL`, `BRAINTRUST_API_KEY`, `ENABLE_VISUAL_REGRESSION`, `ENABLE_ADVANCED_AGENT` documented | No runtime consumer found for any of these | Codebase-wide text search |
| DG-9 | `.env.example:39-41` | `MAX_FIX_ATTEMPTS=3`, `DEFAULT_TIMEOUT_SECONDS=300` | Runtime hard-codes 3 and 300, ignoring environment | `internal/config/config.go:37-38` |
| DG-10 | `README.md:80-84` | Google and DeepSeek listed as supported providers | No runtime consumer for `GOOGLE_API_KEY`/`DEEPSEEK_API_KEY`; provider testing routes them to OpenAI-compatible client but real execution falls through to Anthropic | `internal/api/server.go:1889-1894,3530-3535` |
| DG-11 | `frontend/README.md:1-36` | Generic create-next-app template with Vercel deployment guidance | Self-hosted Compose is the only tracked deployment method | `docker-compose.yml:1-136`; `README.md:182-206` |
| DG-12 | `README.md:201-204` | Mentions `CORS_ALLOWED_ORIGINS` as a recommended enhancement | CORS is hard-coded wildcard with no config field | `internal/api/server.go:98-108`; `internal/config/config.go:7-41` |

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

| Variable | .env.example | Compose | README | Code consumer | Status |
|---|---|---|---|---|---|
| `API_KEY` | Present | Present | Present | `internal/config/config.go:28` | OK |
| `ANTHROPIC_API_KEY` | Present | Present (as `GOTEST_API_KEY` for sidecar) | Present | `internal/config/config.go:31` | Web run ignores env value, reads DB settings only |
| `LLM_PROVIDER` | Absent | Absent | Absent | `internal/ai/client.go:57-73` | MISSING from ops docs |
| `LLM_API_KEY` | Absent | Absent | Absent | `internal/ai/client.go:57-73` | MISSING from ops docs |
| `LLM_BASE_URL` | Absent | Absent | Absent | `internal/ai/client.go:57-73` | MISSING from ops docs |
| `GOTEST_AI_PLANNING` | Absent | Absent | Absent | `internal/api/server.go:1837` | MISSING from ops docs |
| `GITHUB_WEBHOOK_SECRET` | Present | Absent | Mentioned | No consumer | Unused; webhook uses `API_KEY` |
| `JWT_SECRET` | Present | Absent | Absent | No consumer | Unused |
| `BRAINTRUST_API_KEY` | Present | Absent | Absent | No consumer | Unused |
| `VISION_MODEL` | Present | Absent | Absent | No consumer | Unused |
| `GOOGLE_API_KEY` | Absent | Absent | Present | No consumer | README-only |
| `DEEPSEEK_API_KEY` | Absent | Absent | Present | No consumer | README-only |
| `MAX_FIX_ATTEMPTS` | Present | Absent | Absent | Ignored; hard-coded | Stale |
| `DEFAULT_TIMEOUT_SECONDS` | Present | Absent | Absent | Ignored; hard-coded | Stale |
| `STEEL_MAX_SESSIONS` | Present | Absent | Absent | Ignored; hard-coded | Stale |
| `GOTEST_APPROVED_CASE_RUNNER` | Absent | Absent | Absent | `internal/api/server.go:1214` | MISSING from ops docs |

## Priority Order for Documentation Fixes

1. **Critical:** Add missing code-consumed config to `.env.example` and Compose (`LLM_PROVIDER`, `LLM_API_KEY`, `LLM_BASE_URL`, `GOTEST_AI_PLANNING`, `GOTEST_APPROVED_CASE_RUNNER`).
2. **High:** Remove or mark unused `.env.example` variables (`JWT_SECRET`, `GITHUB_WEBHOOK_SECRET`, `BRAINTRUST_API_KEY`, `VISION_MODEL`, `GOOGLE_API_KEY`, `DEEPSEEK_API_KEY`).
3. **High:** Update embedded docs to match active execution paths (`TODO-020`).
4. **High:** Replace `README.md:32` `planning/` reference with `.ai/` (`TODO-019`).
5. **Medium:** Update `docs/docker.md` Steel port and model defaults.
6. **Medium:** Replace `frontend/README.md` with project-specific content (`TODO-022`).
7. **Low:** Remove or mark stale `.env.example` values that code ignores (`MAX_FIX_ATTEMPTS`, etc.).
