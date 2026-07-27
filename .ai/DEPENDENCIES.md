# Dependencies

**Owner:** Engineering  
**Authoritative sources:** `go.mod`, `go.sum`, `frontend/package.json`, `frontend/package-lock.json`, `sidecar/requirements.txt`, Dockerfiles, `docker-compose.yml`  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-26  
**Verification performed:** Static manifest and import inspection; no dependency audit executed  
**Confidence:** High for declared versions and import evidence; runtime wiring status is Verified-static and CVE status is UNKNOWN

## Go Direct Dependencies

Declared at `go.mod:5-14`. “Wired” means constructed by an executable entry point (`cmd/server` and/or `cmd/mcp`); “dormant” means implemented and imported but not wired into a running process at this revision.

| Module | Version | Purpose | Import evidence | Runtime status |
|---|---|---|---|---|
| `github.com/anthropics/anthropic-sdk-go` | v0.2.0-beta.3 | Anthropic LLM client | `internal/agent/llm_anthropic.go:9-10`; `internal/ai/client.go:13-14` | Wired (web + MCP + planning) |
| `github.com/go-chi/chi/v5` | v5.2.1 | HTTP router/middleware | `internal/api/server.go:13-14,55-60` | Wired |
| `github.com/golang-jwt/jwt/v5` | v5.3.1 | JWT signing/validation | `internal/auth/auth.go:15-16,38-67` | Dormant (no production route) |
| `github.com/google/uuid` | v1.6.0 | UUID generation | `internal/api/server.go:35`; `internal/mcp/server.go:12` | Wired |
| `github.com/hibiken/asynq` | v0.25.1 | Redis task queue | `internal/queue/worker.go:14,33-46,81-92` | Dormant (no executable wiring) |
| `github.com/jackc/pgx/v5` | v5.9.2 | PostgreSQL driver/pool | `internal/db/store.go:13,24-44` | Wired (optional) |
| `github.com/mark3labs/mcp-go` | v0.28.0 | MCP stdio server | `internal/mcp/server.go:13-14,26-71` | Wired (MCP process) |
| `github.com/playwright-community/playwright-go` | v0.5700.1 | Browser automation | `internal/agent/playwright_runner.go:11`; `test_playwright.go:1-8` | Wired (web path) |
| `github.com/robfig/cron/v3` | v3.0.1 | Cron expression parsing | `internal/schedule/store.go:11,321-339` | Wired |

**Verified:** No direct Go module is source-unused. `jwt/v5` and `asynq` back dormant runtime paths, not dead imports. Removal is an architecture decision (`ADR-001`, `ADR-005`), not a cleanup.

**Verified:** Indirect modules are pinned with checksums (`go.mod:17-38`; `go.sum`). Notable indirect `github.com/redis/go-redis/v9` supports the Asynq path.

## Frontend Dependencies

Declared at `frontend/package.json:11-28`; resolution pinned by lockfile v3 (`frontend/package-lock.json`), installed with `npm ci` (`frontend/Dockerfile:3-4`).

### Runtime

| Package | Version constraint | Purpose | Usage evidence |
|---|---|---|---|
| `next` | `16.2.7` | App Router framework | Broad imports; `frontend/src/app/layout.tsx:1-5` |
| `react` | `19.2.7` | UI runtime | Broad imports |
| `react-dom` | `19.2.7` | React renderer | Version-matched renderer for Next; no explicit source import expected |
| `lucide-react` | `^1.17.0` | Icons | e.g. `frontend/src/app/alerts/page.tsx:3-6` |
| `clsx` | `^2.1.1` | Class composition | `frontend/src/lib/utils.ts:1-6` |
| `tailwind-merge` | `^3.6.0` | Tailwind class merge | `frontend/src/lib/utils.ts:1-6` |
| `class-variance-authority` | `^0.7.1` | Variant styling | **Verified:** no source import found; variants are hand-written (`frontend/src/components/ui/badge.tsx`) |

**Likely unused:** `class-variance-authority` (`TODO-021`). Confirm no planned generator before removal.

### Dev

| Package | Constraint | Purpose |
|---|---|---|
| `@tailwindcss/postcss`, `tailwindcss` | `^4` | Tailwind 4 build (`frontend/postcss.config.mjs:1-7`) |
| `typescript`, `@types/node`, `@types/react`, `@types/react-dom` | `^5` / `^20` / `^19` | Typing/build (`frontend/tsconfig.json`) |
| `eslint`, `eslint-config-next` | `^9` / `16.2.7` | Linting (`frontend/eslint.config.mjs`) |

**Verified absence:** No test runner or test dependency (`frontend/package.json:5-9,20-29`). See `TODO-012`.

## Python Sidecar Dependencies

Declared with lower-bound ranges and no lockfile (`sidecar/requirements.txt:1-6`).

| Package | Constraint | Purpose | Import evidence |
|---|---|---|---|
| `langgraph` | `>=0.2.60` | Agent state graph | `sidecar/graph.py:1,18-42` |
| `langchain-anthropic` | `>=0.3.12` | `ChatAnthropic` in agents | `sidecar/agents/planner.py:1-6`; `critic.py:1-6` |
| `fastapi` | `>=0.115.0` | HTTP app | `sidecar/main.py:3-4,8-25` |
| `uvicorn` | `>=0.34.0` | ASGI server | `sidecar/Dockerfile:7` |
| `httpx` | `>=0.28.0` | Backend callback client | `sidecar/agents/executor.py:1-2,9-25` |
| `pydantic` | `>=2.11.0` | Request/response schemas | `sidecar/main.py:8-25` |

**Verified:** All declared dependencies are used. **Verified:** `langchain_core` and `typing_extensions` are imported directly but only declared transitively (`sidecar/agents/planner.py:2`; `sidecar/state.py:2`). Consider declaring them explicitly (`TODO-013`).

**Verified:** No Python lockfile; `>=` ranges permit dependency drift (`TODO-013`).

## Runtime Service and Image Dependencies

| Service | Image/tag | Purpose | Reproducibility |
|---|---|---|---|
| Backend | `golang:1.26-bookworm` runtime; Playwright `@latest` | API/MCP build/runtime | **Verified:** Playwright pinned only to `@latest` and install failure suppressed with `|| true` (`Dockerfile:13-14`) |
| Frontend | `node:22-alpine` build/runtime | Dashboard | `npm ci` + lockfile; digest not pinned |
| PostgreSQL | `postgres:16.14-alpine` | Database | Version tag; no digest |
| Redis | `redis:7-alpine` | Asynq target | Major tag; no digest; dormant in server |
| Steel Browser | `steel-browser:latest` | Browser service | **Verified:** mutable `latest` tag (`docker-compose.yml:89-92`) |
| Sidecar | `python:3.12` base | LangGraph agents | Range deps; no lock |

*Evidence:* `Dockerfile:1-18`; `frontend/Dockerfile:1-25`; `sidecar/Dockerfile:1-7`; `docker-compose.yml:57-132`.

## Reproducibility and Lock Status

| Ecosystem | Lock status | Notes |
|---|---|---|
| Go | Locked | `go.mod` + `go.sum` with checksums |
| Frontend | Locked | lockfile v3 + `npm ci` |
| Python | Not locked | `>=` ranges, no lock/hashes |
| Container images | Partially pinned | Version tags without digests; Steel/Playwright mutable |

## Replacement Considerations (Evidence-Grounded)

- **`class-variance-authority`:** Remove if confirmed unused; existing `clsx`/`tailwind-merge` cover current styling (`TODO-021`).
- **`asynq`/Redis:** Keep or remove only as part of the execution/scaling decisions (`ADR-001`, `ADR-004`, `TODO-007`, `TODO-011`).
- **`jwt/v5`:** Retain until the authentication model is decided (`ADR-005`, `TODO-002`).
- **Steel/vision/Braintrust code and dependencies:** Retain pending browser-backend and feature decisions (`ADR-002`); do not treat as dead without owner input.
- **Playwright `@latest`:** Pin to a version compatible with `playwright-go` and remove `|| true` (`TODO-016` scope note; build reliability).
- **Duplicate LLM layers (`internal/agent/llm_*` vs `internal/ai`):** Consolidate under one provider factory as part of execution convergence (`TODO-007`).

## Vulnerability and Audit Status

- **UNKNOWN:** `govulncheck`, `npm audit`, `pip-audit`, and native `gitleaks` were not executed at this revision. Do not describe any ecosystem as CVE-clean without recorded evidence. See `TODO-015`.

## UNKNOWN

- Actual CVE exposure across Go, frontend, Python, and container images.
- Whether Steel/vision/Braintrust/Asynq/JWT dependencies are intended for near-term wiring or removal.
- Whether `class-variance-authority` is reserved for a planned component system.
- Exact upstream guarantees for undeclared transitive Python imports.
