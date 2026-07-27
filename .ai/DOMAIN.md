# Domain Model

**Owner:** Engineering  
**Authoritative sources:** Domain types and business-flow handlers in tracked source  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-26  
**Verification performed:** Static inspection  
**Confidence:** High for type fields and code-evidenced rules; canonical lifecycle and business policy are partly UNKNOWN

## Business Context

**Verified:** Product docs describe an AI-assisted, self-hosted test platform for developers, QA teams, and technical leads (`frontend/src/lib/docs.ts:23-46`).

**Inferred:** The domain models a governed lifecycle from project/spec intake to reusable approved cases, list/schedule execution, audit evidence, maintenance, and release decisions.

**UNKNOWN:** Commercial roles, billing, organizational ownership, tenancy, compliance, formal service-level rules, and which features are production-supported.

## Explicit User Personas Versus Authorization Roles

### Personas (product documentation)

- Developer
- QA team/member
- Technical lead
- Self-hosting team/operator

*Evidence:* `frontend/src/lib/docs.ts:35-46`.

### Authorization roles

**UNKNOWN:** No active user/role/tenant model is enforced. REST uses one optional shared API key; JWT claims exist but are unwired (`internal/api/server.go:2512-2524`; `internal/auth/auth.go:21-95`). Do not infer owner/admin/reviewer permissions from UI labels.

## Core Concepts and Entities

### Project

**Verified:** `project.Project` represents reusable test intake/configuration: identity/name, path/base URL, language/framework, test type, environment, spec, API docs, authentication type/credentials, focus/skip hints, and feature map (`internal/project/store.go:16-31`).

**Verified rules:** Name is required on creation; code populates feature map during project creation/extraction (`internal/api/server.go:273-290,403-420`). Credentials are ordinary fields, not verified secret references.

### FeatureMap

**Verified:** `agent.FeatureMap` contains features/endpoints/areas derived from requirements/spec/API docs and carried by projects/runs (`internal/agent/agent.go`; `internal/api/server.go:1592-1717`).

**UNKNOWN:** Whether generated feature maps require human approval before use.

### TestPlan and Scenario

**Verified:** `agent.TestPlan` has summary and scenarios; scenarios represent named/priority test intents (`internal/agent/agent.go:32-57`). The reusable Agent generates a plan before scripts (`internal/agent/agent.go:242-258`).

### TestPlanDraft

**Verified schema/handler concept:** A draft belongs optionally to a project, has status and JSON cases, and can be regenerated/edited/approved (`internal/db/migrations/004_test_planning_review.sql:1-8`; `internal/api/server.go:459-598`).

**UNKNOWN:** Repository implementation and precise persistence rules because `internal/planning` source is absent.

### TestCase

**Verified schema/handler concept:** Reusable case with title, type, feature, priority, steps, assertions, tags, version, and optional project/plan links (`internal/db/migrations/004_test_planning_review.sql:10-24`). Cases can be run and refined through change proposals (`internal/api/server.go:766-935`).

### ChangeProposal

**Verified schema/flow:** A proposal links to one test case and stores pending/approved/rejected state, prompt/rationale, original/proposed JSON, and review metadata (`internal/db/migrations/008_change_proposals.sql:1-17`; `internal/api/server.go:788-935`). Approval updates the case then the proposal in separate operations (`internal/api/server.go:850-899`).

### TestList

**Verified schema/flow:** Named grouping of test-case IDs with project link, tags, pinning, and run/history operations (`internal/db/migrations/005_test_lists.sql:1-13`; `internal/api/server.go:1045-1147`). Membership is stored as JSON IDs rather than a junction table.

### TestRun

**Verified:** `agent.TestRun` is the central execution aggregate: ID, project target, requirements, mode/type, state, analysis/plan/files/result, attempts/error, credentials/context, video metadata, case/list links, and timestamps (`internal/agent/agent.go:74-109`).

**Verified:** Runs may originate from direct creation, rerun, approved case/list, schedule, webhook, or MCP (MCP runs are not the same durable web runs).

### RunResult and Failure

**Verified:** Results contain passed/failed/total counts, failure details, and video path (`internal/agent/agent.go`). **Verified inconsistency:** some execution paths synthesize success; see `TODO-010`.

### Schedule

**Verified:** Schedule config includes project/list links, target requirements/environment/base URL, mode, frequency/timezone/cron, enabled/next/last run metadata, failure notification, and webhook URL (`internal/schedule/store.go:23-44`; `007_schedules.sql:1-23`).

**Verified:** `CalcNextRun` handles daily/weekly/monthly/custom cron but ignores stored timezone (`internal/schedule/store.go:321-339`).

### ExecutionEvent

**Verified:** Events are run-scoped type/state/message/metadata/timestamp records emitted to in-memory history and subscribers (`internal/events/store.go:13-100`). They are operational projections, not durable source-of-truth facts.

### Recording and VisualArtifact

**Verified:** Recording metadata and visual artifacts are memory-backed runtime evidence (`internal/recordings/store.go`; `internal/visual/store.go`; `internal/execution/context.go`). Video files may live under `/tmp/agent_test/videos` or `/data/videos` depending on runner (`internal/api/server.go:1922-1929`; `internal/runner/docker.go:94-104`).

### Release

**Verified:** `release.Release` and memory store represent release grouping/status data consumed by confidence/risk/summary endpoints (`internal/release/store.go`; `internal/api/server.go:2746-2977`).

**UNKNOWN:** Required durability, release-creation UI workflow, and whether release confidence is considered production evidence.

### Review and Suite

**Verified:** Workflow package defines human review and suite records in memory stores (`internal/workflow/store.go:13-100,103-181`). Request-changes currently uses rejection semantics (`internal/api/server.go:3053-3066`).

### Metrics, Risk, and Intelligence

**Verified:** Metrics derive summary, hotspots, flakiness, and trends from runs (`internal/metrics/metrics.go`). Intelligence derives risk, recommendations, suite selection, release confidence, and explanations (`internal/intelligence/engine.go`). These are calculations over available run/project data, so synthetic/missing input evidence affects their trustworthiness.

### Provider and Settings

**Verified:** Global key-value settings include LLM provider/model/key/base URL, temperature/tokens, browser mode, and fix attempts (`internal/db/migrations/001_init.sql:46-62`; `internal/db/settings_store.go`).

**Verified inconsistency:** provider connection testing and real execution route provider IDs differently (`internal/api/server.go:1889-1894,3516-3535`).

## Business Rules Verified in Code

- A project creation request requires a nonblank name (`internal/api/server.go:273-290`).
- A test list requires name and at least one case ID (`internal/api/server.go:1045-1065`).
- Running a list skips case IDs that cannot be resolved and errors only when none are runnable (`internal/api/server.go:1131-1147`).
- Proposals must be pending before approve/reject (`internal/api/server.go:850-935`).
- API key middleware allows all requests when configured key is empty; otherwise exact shared-key equality is required (`internal/api/server.go:2512-2524`).
- Schedule due processing checks `enabled` and `next_run_at`; selection is not atomically claimed (`internal/schedule/store.go:245-269`).
- Settings update keys are allowlisted, but global settings have no role boundary (`internal/api/server.go:3423-3452`).

## Lifecycle State Machines

There is no owner-confirmed single canonical run lifecycle.

| Path | Verified states |
|---|---|
| Primary web create/rerun/webhook | `idle → analyzing → running → done | failed` (`internal/api/server.go:1543-1953,2165-2195`) |
| Reusable Agent/MCP | `idle → analyzing → plan_generated → writing_tests → running ↔ fixing → done | failed` (`internal/agent/agent.go:17-26,223-345`) |
| Approved case, simulated default | `idle → writing_tests → running → done` (`internal/api/server.go:1203-1248`) |
| Approved case, Docker | `idle → writing_tests → running → done | failed` (`internal/api/server.go:1250-1286`) |
| Non-list schedule | Remains `idle` (`internal/api/server.go:2671-2691,2721-2739`) |
| Sidecar job | `running → completed | failed` (`sidecar/main.py:33-69`) |
| Plan draft | `draft → approved` (`internal/api/server.go:459-598`) |
| Change proposal | `pending → approved | rejected` (`internal/api/server.go:788-935`) |
| Human review | `pending → approved | rejected` (`internal/workflow/store.go`) |

See [`ADR-001`](DECISIONS.md#adr-001--choose-the-canonical-execution-orchestrator).

## Main Domain Workflows

1. **Guided run:** requirements/target → feature map → durable run → LLM plan/scripts → browser → result/events.
2. **Reusable test design:** project → feature map/API parse → plan draft → edited/approved cases.
3. **Governed maintenance:** case → run history/maintenance flags → refinement proposal → review/version update.
4. **Automation:** case membership → test list → direct/list schedule → monitoring/history.
5. **Audit/release:** run evidence → comparison/metrics/risk/review/release confidence/export.
6. **IDE:** MCP request → reusable Agent → synchronous result in independent memory map.

## Glossary

| Term | Meaning |
|---|---|
| Approved case | Reusable `TestCase` materialized from a reviewed plan or edited library |
| Advanced mode | UI/Agent mode intended to use sidecar graph; active dashboard semantics are not implemented |
| Assertion | Expected outcome attached to a test case/result |
| Control Room | Dashboard landing page with live run projection |
| Feature map | Derived product-area/endpoint structure used for planning |
| Run | One execution attempt and its state/result/evidence |
| Scenario | Planned test intent within an Agent `TestPlan` |
| Suite | Legacy/in-memory workflow suite; distinct from database-backed TestList |
| Test list | Database/intended persisted group of approved case IDs used for execution/scheduling |
| Visual artifact | Screenshot/baseline/diff metadata in process memory |

Persisted representation: [`DATABASE.md`](DATABASE.md). Transport contracts: [`API.md`](API.md). Runtime interactions: [`ARCHITECTURE.md`](ARCHITECTURE.md).
