# GoTest Agent Internal Knowledge Base

**Owner:** Engineering  
**Last updated:** 2026-07-26T21:52:37+07:00  
**Repository revision:** `7b54053642e614cccf5e1128defabd25ac88b437` (`master`, no release tag)  
**Verification:** Static repository inspection. Build, tests, runtime flows, and dependency scanners were not successfully executed.  
**Confidence:** High for tracked source structure and static content; runtime correctness remains UNKNOWN.

## Purpose

This tracked `.ai/` directory is the internal engineering knowledge base for GoTest Agent. Its purpose is to let engineers and AI agents understand the project without rereading the entire repository while keeping factual claims tied to repository evidence.

The source tree remains authoritative. If a knowledge document conflicts with code, migrations, manifests, or lockfiles at the recorded revision, the tracked source wins and the affected knowledge document must be corrected. **Verified**, **Inferred**, and **UNKNOWN** labels are intentional evidence boundaries, not prose decoration.

## Authority Order

1. Tracked executable code, manifests, migrations, and lockfiles at the recorded revision
2. Observed build, test, scanner, and runtime evidence
3. [`PROJECT_STATE.md`](PROJECT_STATE.md), the canonical current status snapshot
4. Detailed evidence documents in this directory
5. Product README, operator docs, and embedded user documentation
6. Deleted historical planning artifacts and commit-message intent

Git commit subjects establish chronology, not runtime correctness or architectural rationale.

## Document Map

| Document | Canonical responsibility |
|---|---|
| [`PROJECT_STATE.md`](PROJECT_STATE.md) | Current implementation status, blockers, health, priorities, and verification state |
| [`DISCOVERY.md`](DISCOVERY.md) | Repository and business overview, stack, structure, evidence catalog, confidence, and unknowns |
| [`ARCHITECTURE.md`](ARCHITECTURE.md) | Runtime topology, boundaries, layers, component interactions, flows, constraints, and architectural risks |
| [`CODEMAP.md`](CODEMAP.md) | Directory/package/page navigation, entry points, key symbols, hotspots, and common change paths |
| [`DATABASE.md`](DATABASE.md) | Physical schema, relationships, indexes, migration behavior, repositories, seeds, and persistence boundaries |
| [`API.md`](API.md) | REST/SSE, MCP, and sidecar transport contracts; auth, validation, responses, errors, and versioning |
| [`DOMAIN.md`](DOMAIN.md) | Business concepts, entities, lifecycle semantics, roles, workflows, and glossary |
| [`DEPENDENCIES.md`](DEPENDENCIES.md) | Packages, runtime services, versions, purposes, wiring status, replacement considerations, and audit state |
| [`DECISIONS.md`](DECISIONS.md) | Architecture decision records and explicit decision-needed entries |
| [`ROADMAP.md`](ROADMAP.md) | Outcome-oriented milestones, priority horizons, blockers, and refactoring opportunities |
| [`TODO.md`](TODO.md) | Single actionable backlog with IDs, acceptance criteria, dependencies, impact, and status |
| [`CHANGELOG_AI.md`](CHANGELOG_AI.md) | Append-only provenance of AI-maintained knowledge and future AI-generated modifications |

## Evidence Labels

- **Verified** — Exact static content exists in tracked source at the recorded revision, or an observed command/runtime result proves it.
- **Inferred** — A conclusion follows from static evidence but is not explicitly confirmed by an owner or runtime observation.
- **UNKNOWN** — Information cannot currently be verified. Never fill UNKNOWN fields with assumptions.

## Stable IDs

- `ADR-###` — decisions and decision-needed records
- `TODO-###` — executable work items
- `RISK-###` — persistent project risks
- `UNK-###` — unresolved questions

Documents should link these IDs rather than copying complete narratives.

## Maintenance Rules

After every meaningful implementation:

1. Review the knowledge base.
2. Update only documents affected by the change.
3. Preserve historical ADRs and changelog entries; mark superseded information rather than deleting history.
4. Update `Last verified revision` only after reconciling the document against that revision.
5. Add or update TODO/ADR IDs instead of creating parallel backlogs.
6. Append an entry to `CHANGELOG_AI.md` for AI-generated modifications.
7. Record exactly which verification commands or end-to-end flows ran. Do not convert static inspection into runtime proof.
8. If source and docs conflict, update docs immediately and record the discrepancy in `CHANGELOG_AI.md`.

## Known Documentation Conflict

**Verified:** `README.md:32` points readers to a `planning/` directory, while `.gitignore:72-73` ignores that directory and no tracked `planning/` exists. This `.ai/` directory is the new tracked internal knowledge location. Updating the root README reference is tracked as [`TODO-019`](TODO.md#todo-019-correct-root-readme-knowledge-base-reference).
