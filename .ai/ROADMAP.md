# Engineering Roadmap

**Owner:** Engineering  
**Authoritative sources:** Accepted ADRs, [`PROJECT_STATE.md`](PROJECT_STATE.md), and [`TODO.md`](TODO.md)  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-26  
**Confidence:** High for blocker sequencing; milestone dates and owner capacity are UNKNOWN

## Roadmap Principles

- Outcomes, not unchecked historical feature lists.
- Buildability and security precede new features.
- One canonical TODO backlog; this document links tasks rather than duplicating their status.
- Architecture-changing work requires an accepted ADR.
- Preserve backward compatibility unless a documented migration is approved.
- No milestone is complete without observed verification and knowledge-base reconciliation.

## Milestone 0 — Restore a trustworthy baseline

**Priority:** Critical  
**Outcome:** A buildable repository with reproducible baseline evidence.

### Progress
- `internal/planning` reconstructed; `go vet` and `go build ./...` passed (2026-07-26) ✅
- `gofmt` formatting and `TestGenerateApprovePlanLifecycle` regression pending infrastructure ⏳
- Full `go test ./...` pending ⏳
- Scanner audits (`govulncheck`, `npm audit`, `pip-audit`, `gitleaks`) pending ⏳
- Root README knowledge-base reference corrected (`TODO-019`) ✅

### Scope

- Restore/resolve missing planning package (`TODO-001`).
- Run build/test/lint commands and record exact results (`TODO-014`).
- Run dependency and secret audits (`TODO-015`).
- Convert new failures into evidence-backed TODOs rather than expanding scope inline.

### Exit criteria

- `go build ./cmd/server` and `go test ./...` have observed outcomes.
- Frontend lint/build has an observed outcome.
- Scanner status is no longer UNKNOWN, or tool blockers are documented.
- `PROJECT_STATE.md` reflects observed evidence.

## Milestone 1 — Establish a safe deployment boundary

**Priority:** Critical / High  
**Target date:** UNKNOWN  
**Outcome:** The documented deployment no longer fails open or exposes an unauthenticated privileged sidecar.

### Scope

- Accept authentication model (`ADR-005`).
- Make API auth fail closed (`TODO-002`).
- Secure/internalize sidecar (`TODO-003`).
- Separate webhook secret (`TODO-004`).
- Align dashboard REST/SSE auth (`TODO-005`).
- Bind LLM credentials to approved origins (`TODO-006`).
- Restrict runner/browser egress (`TODO-016`).

### Exit criteria

- Production-mode startup requires explicit safe auth.
- Dashboard authenticated REST and SSE work end-to-end.
- Sidecar cannot be invoked by untrusted callers.
- Stored credentials cannot be sent to caller-selected destinations.
- Deployment/operator docs match behavior.

## Milestone 2 — Converge execution correctness

**Priority:** High  
**Target date:** UNKNOWN  
**Outcome:** All entry points use an owner-approved orchestration contract and do not report synthetic success as real execution.

### Scope

- Accept canonical orchestrator and browser backend (`ADR-001`, `ADR-002`).
- Consolidate execution paths (`TODO-007`).
- Fix non-list schedule execution (`TODO-008`).
- Implement real run deletion (`TODO-009`).
- Eliminate or label synthetic success (`TODO-010`).
- Standardize provider selection as part of canonical orchestration.

### Exit criteria

- HTTP, MCP, schedules, webhooks, and approved cases have documented shared semantics.
- Terminal states and failures are consistent.
- API/UI selections (`mode`, `test_type`, provider) affect execution as documented.
- Integration tests cover each supported entry point.

## Milestone 3 — Define durability and scaling

**Priority:** High / Medium  
**Target date:** UNKNOWN  
**Outcome:** Restart and multi-instance behavior are explicit, tested, and aligned with product needs.

### Scope

- Accept durability and deployment-cardinality decisions (`ADR-003`, `ADR-004`).
- Persist/bound operational state (`TODO-011`).
- Make schedule claiming atomic if multi-instance (`TODO-017`).
- Decide queue/worker ownership as part of `ADR-001`/`ADR-004`.
- Add backup/restore and retention requirements after owner confirms RPO/RTO.

### Exit criteria

- Every user-visible entity has documented durability/retention.
- Restart behavior is tested.
- Scheduler ownership and duplicate-prevention semantics are tested.
- Production backup/restore evidence exists if persistence is required.

## Milestone 4 — Strengthen maintainability and verification

**Priority:** Medium  
**Target date:** UNKNOWN  
**Outcome:** Core surfaces are protected by automated tests and consistent transport boundaries.

### Scope

- Frontend tests (`TODO-012`).
- Sidecar tests/lockfile (`TODO-013`).
- Standard validation/error/server lifecycle (`TODO-018`).
- Split oversized API module only after behavior is locked and as incremental PR-sized changes.
- Define DTOs/redaction where required by accepted auth model.

### Exit criteria

- CI-capable test commands exist for Go, frontend, and sidecar.
- Critical flows have end-to-end coverage.
- Error/validation contracts are documented and consistent.
- Refactoring preserves observed behavior.

## Milestone 5 — Documentation and developer experience

**Priority:** Low / ongoing  
**Target date:** Ongoing  
**Outcome:** Product, operator, and engineering documentation agree with tracked behavior.

### Scope

- Correct root knowledge-base reference (`TODO-019`).
- Reconcile embedded/operator docs (`TODO-020`).
- Remove confirmed unused dependency (`TODO-021`).
- Replace generic frontend README (`TODO-022`).
- Maintain `.ai/` after every meaningful change.

### Exit criteria

- No known source/document conflicts remain open without a TODO.
- New engineer can find architecture, API, database, domain, and development paths from `.ai/README.md`.
- Every AI-generated implementation has a changelog entry.

## Future Refactoring Opportunities

These are **not approved work** and must not bypass milestones or ADRs:

- Extract domain application services from `internal/api/server.go` after regression coverage exists.
- Consolidate the duplicate LLM client/factory layers.
- Introduce typed API DTOs and resource-level redaction.
- Normalize test-list membership if relational integrity is required.
- Replace process polling for run state with a durable event or notification mechanism.
- Generate route/schema/dependency inventory checks to detect documentation drift.

## Known Roadmap Blockers

- Owner decisions for `ADR-001` through `ADR-005`.
- Missing planning source.
- UNKNOWN build/test/scanner state.
- UNKNOWN product tenancy, durability, scaling, RPO/RTO, and external deployment controls.
