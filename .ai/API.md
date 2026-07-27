# API Reference

**Owner:** Engineering  
**Authoritative sources:** Route registry and handler bodies in `internal/api/server.go`; MCP and sidecar server source  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-26  
**Verification performed:** Static route/handler inspection; no HTTP execution  
**Confidence:** High for registration and static handler behavior; runtime responses and missing planning handlers are partly UNKNOWN

## Versioning and Base URLs

- **Verified:** Main API is path-versioned at `/api/v1` (`internal/api/server.go:133-235`).
- **Verified:** Public health and temporary video routes sit outside `/api/v1`; a protected video file server is registered separately (`internal/api/server.go:124-131,257-261`).
- **Verified:** Frontend defaults to `http://localhost:8080` (`frontend/src/lib/api.ts:1`).
- **UNKNOWN:** Backward-compatibility/deprecation policy, production API origin, and proxy behavior.

## Authentication and Authorization

- **Verified:** All routes inside `/api/v1` use shared `X-Api-Key` middleware. Empty configured key bypasses authentication; nonempty key requires exact equality (`internal/api/server.go:133-135,2512-2524`).
- **Verified:** No active per-user identity, role, tenant, ownership, or resource authorization checks were found.
- **Verified:** JWT code exists but is not wired to production routes (`internal/auth/auth.go:21-95`).
- **Verified:** GitHub webhook uses HMAC only when the supplied secret is nonempty. Router supplies `cfg.APIKey`, not documented `GITHUB_WEBHOOK_SECRET` (`internal/api/server.go:237-255`; `internal/webhook/github.go:59-100`).
- **Verified:** Frontend shared fetch sends no API-key header; native EventSource path cannot set it (`frontend/src/lib/api.ts:222-229,391-415`).
- **UNKNOWN:** Owner-approved production auth model (`ADR-005`).

### Legend

- `P` — Public
- `K` — Conditional `X-Api-Key` (`API_KEY=""` means public)
- `H` — Conditional GitHub HMAC
- Validation: `ID` uses 1–64 alphanumeric/underscore/hyphen helper (`internal/api/server.go:111-122`); `J` means JSON decode; `R` means additional rule; `—` means no meaningful validation observed.

## Public and Special Routes

| Method | Path | Handler/purpose | Auth | Validation/response |
|---|---|---|---|---|
| GET | `/health` | Process health | P | `200` JSON status; no dependency readiness (`server.go:268-270`) |
| GET | `/videos/{filename}` | Serve direct-run temp video | P | Filename path; file response (`server.go:127-131`) |
| POST | `/api/v1/webhooks/github` | GitHub push callback → async run | H | HMAC if secret nonempty; push JSON; `200` before callback completion (`server.go:237-255`; `webhook/github.go:46-100`) |
| ANY | `/videos/*` | Persistent `/data/videos` file server | K | File server (`server.go:257-261`) |

## REST Endpoint Inventory

All rows below use `K` authentication.

### Projects

| Method | Path | Handler | Purpose / validation / response |
|---|---|---|---|
| POST | `/api/v1/projects` | `handleCreateProject` | Create; J; nonblank name; JSON project |
| GET | `/api/v1/projects` | `handleListProjects` | List; JSON array |
| GET | `/api/v1/projects/{id}` | `handleGetProject` | ID + existence; JSON project |
| PATCH | `/api/v1/projects/{id}` | `handleUpdateProject` | ID/J/existence; JSON project |
| POST | `/api/v1/projects/{id}/api-docs` | `handleUploadAPIDocs` | ID/J/existence; updated project |
| POST | `/api/v1/projects/{id}/parse-api` | `handleParseAPIDocs` | ID/project/docs; parsed draft/cases |
| POST | `/api/v1/projects/{id}/extract-features` | `handleExtractProjectFeatures` | ID/existence; feature map |
| POST | `/api/v1/projects/{id}/test-plan` | `handleGenerateProjectTestPlan` | ID/project existence; draft plan |

*Registration:* `internal/api/server.go:136-143`; handler bodies `:273-485`.

### Test Plans and Cases

| Method | Path | Handler | Purpose / validation |
|---|---|---|---|
| GET | `/api/v1/test-plans/{id}` | `handleGetTestPlan` | ID/existence |
| PATCH | `/api/v1/test-plans/{id}/cases/{caseId}` | `handleUpdateTestPlanCase` | J; plan/case membership |
| POST | `/api/v1/test-plans/{id}/regenerate` | `handleRegenerateTestPlan` | Plan/project existence |
| POST | `/api/v1/test-plans/{id}/approve` | `handleApproveTestPlan` | Plan existence; materialize enabled cases |
| GET | `/api/v1/test-cases` | `handleListTestCases` | Optional unvalidated `project_id` |
| GET | `/api/v1/test-cases/maintenance` | `handleTestCaseMaintenance` | Optional `project_id`; derived flags |
| GET | `/api/v1/test-cases/{id}` | `handleGetTestCase` | ID/existence |
| PATCH | `/api/v1/test-cases/{id}` | `handleUpdateTestCase` | J/existence; no enum validation |
| POST | `/api/v1/test-cases/{id}/run` | `handleRunTestCase` | ID/existence; `202` run |
| POST | `/api/v1/test-cases/{id}/refine` | `handleRefineTestCase` | ID/J/nonblank prompt; proposal |

*Registration:* `server.go:144-153`; handlers `:488-817`. Runtime planning behavior is UNKNOWN due missing package.

### Change Proposals

| Method | Path | Handler | Validation |
|---|---|---|---|
| GET | `/api/v1/test-cases/{id}/proposals` | `handleListTestCaseProposals` | ID |
| GET | `/api/v1/change-proposals` | `handleListChangeProposals` | Optional validated test-case ID |
| POST | `/api/v1/change-proposals/{id}/approve` | `handleApproveChangeProposal` | ID, pending state; optional body |
| POST | `/api/v1/change-proposals/{id}/reject` | `handleRejectChangeProposal` | ID, pending state; optional body |

*Evidence:* `server.go:154-157,820-935`.

### Test Lists

| Method | Path | Handler | Validation/response |
|---|---|---|---|
| POST | `/api/v1/test-lists` | `handleCreateTestList` | J; name + nonempty IDs; JSON list |
| GET | `/api/v1/test-lists` | `handleListTestLists` | Optional `project_id` |
| GET | `/api/v1/test-lists/{id}` | `handleGetTestList` | ID/existence |
| GET | `/api/v1/test-lists/{id}/history` | `handleTestListHistory` | ID/existence; derived history |
| POST | `/api/v1/test-lists/{id}/run` | `handleRunTestList` | ID/runnable cases; `202` run ID array |

*Evidence:* `server.go:158-162,1045-1147`.

### Runs and Evidence

| Method | Path | Handler | Validation/response |
|---|---|---|---|
| POST | `/api/v1/runs` | `handleCreateRun` | J only; persists idle; `202` run ID/state |
| GET | `/api/v1/runs` | `handleListRuns` | Hard-coded first 50; JSON array |
| GET | `/api/v1/runs/{id}` | `handleGetRun` | ID/existence; JSON run |
| POST | `/api/v1/runs/{id}/rerun` | `handleRerun` | Existence; `202` new run |
| GET | `/api/v1/runs/{id}/stream` | `handleSSEStream` | No upfront existence check; SSE |
| GET | `/api/v1/runs/{id}/events` | `handleGetEvents` | Memory event array |
| GET | `/api/v1/runs/{id}/api-logs` | `handleGetAPILogs` | ID/existence; explicit empty placeholder |
| GET | `/api/v1/runs/{id}/report` | `handleReport` | Existence; HTML |
| POST | `/api/v1/runs/{id}/analyze-failure` | `handleAnalyzeFailure` | Existence; analysis JSON |
| GET | `/api/v1/runs/{id}/compare/{otherId}` | `handleCompare` | Both exist; comparison JSON |
| GET | `/api/v1/runs/{id}/recordings` | `handleGetRecordings` | Memory metadata array |
| GET | `/api/v1/runs/{id}/visual` | `handleGetVisualArtifacts` | Memory artifact array |
| GET | `/api/v1/runs/{id}/video` | `handleGetVideoMetadata` | Existence; metadata JSON |
| DELETE | `/api/v1/runs/{id}` | `handleDeleteRun` | Always `204`; no deletion (`server.go:2508-2510`) |
| GET | `/api/v1/recordings` | `handleListAllRecordings` | All memory recordings |
| GET | `/api/v1/stream` | `handleGlobalStream` | Global SSE |

*Registration:* `server.go:164-180`; handlers `:1543-2505,3168-3215`.

### Monitoring and Schedules

| Method | Path | Handler | Validation/response |
|---|---|---|---|
| GET | `/api/v1/monitoring/summary` | `handleMonitoringSummary` | Derived summary |
| POST | `/api/v1/schedules` | `handleCreateSchedule` | J; optional list existence; JSON schedule |
| GET | `/api/v1/schedules` | `handleListSchedules` | JSON array |
| GET | `/api/v1/schedules/{id}` | `handleGetSchedule` | Existence |
| PATCH | `/api/v1/schedules/{id}` | `handleUpdateSchedule` | J map; unchecked type assertions can panic |
| DELETE | `/api/v1/schedules/{id}` | `handleDeleteSchedule` | `204`/404 |
| POST | `/api/v1/schedules/{id}/run-now` | `handleRunNow` | Schedule/list existence; non-list path remains idle |

*Evidence:* `server.go:181-188,2064-2147,2529-2691`.

### Releases and Intelligence

| Method | Path | Handler |
|---|---|---|
| POST | `/api/v1/releases` | `handleCreateRelease` |
| GET | `/api/v1/releases` | `handleListReleases` |
| GET | `/api/v1/releases/{id}` | `handleGetRelease` |
| PATCH | `/api/v1/releases/{id}` | `handleUpdateRelease` (unchecked patch assertions) |
| GET | `/api/v1/releases/{id}/summary` | `handleReleaseSummary` |
| GET | `/api/v1/releases/{id}/confidence` | `handleReleaseConfidence` |
| GET | `/api/v1/releases/{id}/risk` | `handleReleaseRisk` |
| GET | `/api/v1/releases/{id}/explanation` | `handleReleaseExplanation` |
| POST | `/api/v1/suite-selection` | `handleSuiteSelection` |

*Evidence:* `server.go:190-208,2746-2994`.

### Metrics

| Method | Path | Handler |
|---|---|---|
| GET | `/api/v1/metrics/summary` | `handleMetricsSummary` |
| GET | `/api/v1/metrics/hotspots` | `handleMetricsHotspots` |
| GET | `/api/v1/metrics/flaky` | `handleMetricsFlaky` |
| GET | `/api/v1/metrics/trend` | `handleMetricsTrend` |
| GET | `/api/v1/metrics/risk` | `handleMetricsRisk` |
| GET | `/api/v1/metrics/recommendations` | `handleMetricsRecommendations` |

*Evidence:* `server.go:198-203,2834-2903`.

### Reviews, Suites, Notifications, Alerts

| Method | Path | Handler |
|---|---|---|
| POST | `/api/v1/reviews` | `handleCreateReview` |
| GET | `/api/v1/runs/{id}/reviews` | `handleGetRunReviews` |
| POST | `/api/v1/reviews/{id}/approve` | `handleApproveReview` |
| POST | `/api/v1/reviews/{id}/reject` | `handleRejectReview` |
| POST | `/api/v1/reviews/{id}/request-changes` | `handleRequestChangesReview` (implemented as reject) |
| GET | `/api/v1/reviews` | `handleListAllReviews` |
| POST | `/api/v1/suites` | `handleCreateSuite` |
| GET | `/api/v1/suites` | `handleListSuites` |
| GET | `/api/v1/suites/{id}` | `handleGetSuite` |
| DELETE | `/api/v1/suites/{id}` | `handleDeleteSuite` |
| GET | `/api/v1/notifications` | `handleListNotifications` |
| POST | `/api/v1/alert-rules/evaluate` | `handleEvaluateAlertRules` (calculates, does not persist/deliver) |

*Evidence:* `server.go:196,210-222,2823-2829,2999-3154`.

### Settings, AI, Demo, Exports

| Method | Path | Handler / notes |
|---|---|---|
| GET | `/api/v1/settings` | `handleGetSettings`; masks long LLM key |
| PUT | `/api/v1/settings` | `handleUpdateSettings`; key-name allowlist; no destination policy |
| GET | `/api/v1/ai/providers` | `handleListAIProviders` |
| POST | `/api/v1/ai/test-provider` | `handleTestAIProvider`; failure encoded as HTTP 200 + success false |
| POST | `/api/v1/demo/seed` | `handleDemoSeed` |
| GET | `/api/v1/runs/{id}/export` | `handleExportRun` |
| GET | `/api/v1/runs/{id}/compare/{otherId}/export` | `handleExportCompare` |
| GET | `/api/v1/metrics/risk/export` | `handleExportRisk` |
| GET | `/api/v1/releases/{id}/confidence/export` | `handleExportConfidence` |

*Evidence:* `server.go:224-234,3227-3556`.

## Request Flow and Validation

1. Global Chi request ID, logger, recoverer, wildcard CORS (`server.go:55-60,98-108`).
2. `/api/v1` conditional key middleware.
3. Handlers directly decode JSON, perform handwritten validation, call stores/functions/integrations, and encode response.
4. Run endpoints start background goroutines and return `202`.

**Verified gaps:** No shared schema validator, `DisallowUnknownFields`, trailing-JSON check, global body limit, consistent typed PATCH model, common JSON error DTO, or graceful server lifecycle. Wrong PATCH types panic into Chi recovery (`server.go:2581-2617,2778-2792`).

## Response and Error Formats

- Success: ad-hoc domain structs/maps encoded as JSON.
- Errors: plain-text `http.Error`, generally 400/401/404/500.
- HTML: run report.
- SSE: per-run/global event streams.
- Files: video endpoints.
- Provider test: HTTP 200 with `{success:false}` on connection failure.
- No standard error code/envelope/version migration policy found.

## SSE

- Per-run: event replay + subscription + two-second durable state polling (`server.go:2197-2262`).
- Global: all-event subscription + heartbeat (`server.go:3168-3215`).
- Frontend per-run EventSource closes permanently on error; global page reconnects and polls (`frontend/src/lib/api.ts:391-415`; `frontend/src/app/page.tsx:32-66`).

## MCP Tools

| Tool | Verified behavior |
|---|---|
| `run_tests` | Creates local run, synchronously executes reusable Agent, returns JSON result |
| `analyze_project` | LLM project analysis |
| `generate_test_plan` | LLM plan generation |
| `get_run_status` | Reads MCP-local memory run map |

*Evidence:* `internal/mcp/server.go:33-143`. MCP state is independent from web PostgreSQL/SSE.

## Sidecar API

| Method | Path | Behavior/auth |
|---|---|---|
| GET | `/health` | Public process health |
| POST | `/agent/run` | Public job creation, Pydantic body, background thread |
| GET | `/agent/{job_id}` | Public status/result by UUID |

*Evidence:* `sidecar/main.py:15-69`. No inbound auth found; see `TODO-003`.

## Known Gaps

- Missing planning implementation makes planning route runtime behavior UNKNOWN.
- Conditional API-key model conflicts with browser client (`TODO-002`, `TODO-005`).
- Webhook secret/config mismatch (`TODO-004`).
- LLM base URL/credential boundary (`TODO-006`).
- No-op run deletion (`TODO-009`).
- Synthetic execution results (`TODO-010`).
- Validation/error/lifecycle consistency (`TODO-018`).
