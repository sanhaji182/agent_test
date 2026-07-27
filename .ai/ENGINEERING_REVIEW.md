# Engineering Review

**Owner:** Engineering  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-26  
**Verification performed:** Static inspection covering all verified audit and discovery evidence  
**Confidence:** High for static findings; build/test/scanner results partly UNKNOWN

## Summary

This review consolidates findings from the static discovery (`.ai/DISCOVERY.md`), architecture audit (root `AUDIT.md`), and targeted source inspections. It covers the full repository surface: architecture, backend, frontend, database, API, infrastructure, dependencies, code quality, security, and operations.

## Architecture

**Classification:** Modular monolith with in-process asynchronous execution, SSE-driven event projection, and an optional distributed deployment shell (Compose includes Redis/Steel/sidecar but none are in the primary web-run path).

**Strengths:**
- Package-per-domain organization under `internal/`.
- Repository interfaces at persistence and external-call boundaries (LLM, Runner, RunStore, project Store, schedule Repository).
- Embedded, ordered SQL migrations with schema tracking.
- Interface-mock-based tests for the agent and API packages.

**Weaknesses:**
- The ~3,556-line `internal/api/server.go` combines router, composition root, 85+ handlers, and application orchestration. Every cross-cutting change requires editing this file.
- Five divergent execution paths with different state machines, error handling, and persistence semantics.
- The previously missing `internal/planning` package was reconstructed (2026-07-26); its integration with the existing PostgreSQL/memory store selection pattern is complete.
- Several deployed or implemented components (Redis/Asynq, Steel, sidecar, vision, Braintrust) are not wired to the active execution path.

## Backend (Go)

**Strengths:**
- Chi-based routing with clean middleware composition.
- Parameterized SQL throughout all PostgreSQL stores.
- `html/template` auto-escaping for reports.
- `crypto/rand` for API key generation.
- Go module and sum-file pinning.

**Weaknesses:**
- Handlers mix transport, validation, business logic, and persistence.
- Inconsistent validation: route ID allowlist exists but several ID-bearing handlers bypass it; no body-size limits; no `DisallowUnknownFields`.
- PATCH handlers use unchecked `v.(bool)`/`v.(string)` type assertions that panic on wrong types (`internal/api/server.go:2581-2617,2778-2792`).
- `http.ListenAndServe` without timeouts or graceful shutdown (`cmd/server/main.go:44-46`).
- Two independent LLM client layers with divergent provider routing (test connection vs real execution).
- Environment LLM defaults are ignored by web execution when database settings are absent.
- Several `UpdateRun` errors are silently discarded.

## Frontend

**Strengths:**
- Consistent App Router pattern across 16 pages.
- Centralized typed API client (`frontend/src/lib/api.ts`).
- Shared UI and console component library.
- Bilingual embedded documentation.
- Light/dark theme with persistence.

**Weaknesses:**
- No frontend tests — zero test scripts, zero test dependencies in `package.json`.
- No authentication integration — fetches send no API key; EventSource cannot attach custom headers.
- Several pages silently catch network errors and render empty state (alerts, monitoring, risk, releases, overview).
- Many mutation handlers have no error feedback (projects, suites, tests, reviews).
- Per-run SSE closes permanently on error with no reconnect or polling fallback.
- Settings saves optimistically without checking `res.ok`.
- GitHub webhook URL uses `window.location.origin` instead of `NEXT_PUBLIC_API_URL`.
- Custom tabs and screenshot lightbox lack ARIA roles and keyboard support.
- Sidebar is fixed at 220px with no responsive collapse.
- `class-variance-authority` declared but unused.

## Database

**Strengths:**
- Ten well-structured tables covering the product domain.
- Indexes on common query paths (state, project, timestamps, test-case/list links).
- Default settings seed with `ON CONFLICT DO NOTHING`.
- Parameterized query support and JSONB for flexible payloads.

**Weaknesses:**
- Migration execution and version recording are non-atomic (`internal/db/migrate.go:45-69`).
- Migration failure only logs; startup continues serving against potentially broken schema (`cmd/server/main.go:28-34`).
- `test_runs.test_case_id` and `test_runs.test_list_id` are not foreign keys; test-list membership is stored as a JSON array rather than a junction table.
- Project credentials and LLM API key are stored as plaintext database values.
- Due schedule selection has no lock or lease, enabling duplicates in multi-instance or overlapping scheduler scenarios.
- Proposal approval, test-case batch creation, and schedule processing are multi-step and non-transactional.
- Several store operations miss `rows.Err()`, `RowsAffected` checks, or error propagation.
- No PostgreSQL integration tests exist; all current tests exercise memory stores only.
- No backup, restore, retention, or vacuum automation exists.

## API

**Strengths:**
- 85+ registered REST routes covering the full product surface.
- Conditional API-key middleware on all `/api/v1` routes.
- MCP stdio server with four IDE tools.
- SSE for live control-room and per-run event delivery.
- GitHub webhook with HMAC verification when configured.

**Weaknesses:**
- API authentication is fail-open — empty `API_KEY` bypasses all protection.
- No user identity, roles, ownership checks, or per-endpoint authorization.
- GitHub webhook reuses `API_KEY` as HMAC secret instead of reading the separately documented `GITHUB_WEBHOOK_SECRET`.
- Frontend cannot send API-key headers; dashboard breaks if authentication is enabled.
- Wildcard CORS is hard-coded with no configurable origin restriction.
- No shared error envelope; errors are plain-text strings with ad-hoc status codes.
- No request body-size limits, server timeouts, or graceful shutdown.
- Public unauthenticated `/videos/{filename}` route bypasses the separately authenticated video server.
- Provider testing returns HTTP 200 with `{"success":false}` for failures.
- API logs endpoint is an explicit empty placeholder.
- Run deletion endpoint returns 204 without deleting anything.

## Infrastructure and Operations

**Strengths:**
- Docker Compose full-stack deployment with health checks on all six services.
- Named volumes for PostgreSQL and application data.
- Multi-stage Dockerfiles for backend, frontend, and sidecar.
- Makefile for build and Compose lifecycle.
- `.gitignore` rules for secrets and data.
- Gitleaks configuration.

**Weaknesses:**
- No CI/CD workflow found.
- No IaC, Kubernetes manifests, or cloud deployment configuration.
- No backup, restore, or disaster recovery procedures.
- PostgreSQL and Redis are host-published with known credentials.
- Sidecar is host-published with no inbound authentication.
- Steel receives `SYS_ADMIN` and uses mutable `:latest` tag.
- Backend and sidecar run as image-default root.
- No TLS anywhere.
- No rate limiting.
- `SECURITY.md` catalogs confirmed default-configuration vulnerabilities in detail.

## Dependencies

**Strengths:**
- All nine direct Go modules have import evidence.
- Frontend lockfile + `npm ci` ensure reproducible resolution.
- Go `go.sum` chains all direct and indirect modules.

**Weaknesses:**
- `govulncheck`, `npm audit`, and `pip-audit` have not been executed.
- CVE status across all ecosystems is UNKNOWN — not "clean."
- `jwt/v5` and `asynq` modules are runtime-dormant (not source-unused).
- Python dependencies use `>=` with no lockfile.
- `langchain_core` and `typing_extensions` are directly imported but not explicitly declared in `requirements.txt`.
- Steel uses `:latest` tag; Playwright installed at `@latest` with `|| true`.
- Backend image tags and Redis tag are not digest-pinned.

## Testing

**Strengths:**
- Go tests exist for agent, API, auth, compare, events, gitdiff, intelligence, metrics, schedule, and visual packages.
- API test covers a full project→plan→case→list→run→schedule→monitoring lifecycle.
- Interface-mock approach used for LLM and Runner in agent tests.

**Weaknesses:**
- Go tests have not been executed at this revision.
- All current tests use memory stores only; no PostgreSQL integration tests exist.
- No frontend tests — zero test infrastructure.
- No Python sidecar tests.
- No end-to-end or cross-service tests.
- No performance, load, security, or chaos tests.
- No test for the reconstructed planning package's store contracts.

Full testing strategy: `.ai/TESTING.md`.

## Security

Full findings: `.ai/SECURITY.md`. Priority vulnerabilities in the default Compose configuration:

1. API auth fails open (empty `API_KEY`).
2. Unauthenticated sidecar with backend credential forwarding.
3. Host-published PostgreSQL with known credentials and disabled TLS.
4. Configurable LLM base URL reuses stored credential as Bearer token.
5. Docker runner TypeScript injection via unescaped project URL (requires explicit opt-in or MCP access).
6. No HTTP server timeouts, body limits, or concurrency controls.
7. Public unauthenticated temporary video route.

Plus design/deployment-dependent risks: unrestricted browser egress, shared authorization domain with plaintext credential fields, Redis/Steel exposure, and supply-chain reproducibility gaps.

## Maintainability

- `internal/api/server.go` at ~3,556 lines is the primary maintainability risk — it combines too many responsibilities.
- Several dead or unwired code paths (`simulateMockRun`, dormant JWT, Asynq, Steel, vision, Braintrust) should be resolved by action or explicit documentation.
- The two diverging LLM client layers create confusion for provider support.
- Inconsistent validation and error patterns across handlers make API behavior hard to predict.
- The documentation (README, `.env.example`, embedded docs, Docker guide, frontend README) contains conflicts with active implementation.

Full technical debt inventory: `.ai/TECHNICAL_DEBT.md`.

## Overall Assessment

The repository has a credible product surface (85+ routes, 16 frontend pages, 4 MCP tools, 10 database tables) but remains pre-production in reliability, security defaults, test coverage, and operational readiness. The recently removed critical compilation blocker (planning package) is the most significant improvement at this revision.

The highest-value work remaining is: close security defaults, converge execution paths, add test coverage, add server infrastructure (timeouts, graceful shutdown, bounded concurrency), and reconcile documentation.
