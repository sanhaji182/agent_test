# Contributing

**Owner:** Engineering  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-27

## Before You Start

This is a polyglot project. Read `.ai/README.md` for the knowledge base map and authority order. Source code, manifests, migrations, and lockfiles are authoritative over documentation.

## Repository Layout

| Directory | Purpose |
|---|---|
| `cmd/server/` | Go HTTP API entry point |
| `cmd/mcp/` | Go MCP stdio entry point |
| `internal/` | All Go application packages (domain, persistence, transport, execution) |
| `frontend/` | Next.js dashboard application |
| `sidecar/` | Python LangGraph multi-agent service |
| `docs/` | Operator documentation |
| `.ai/` | Internal engineering knowledge base (tracked, maintained with every change) |

Full code map: `.ai/CODEMAP.md`. Architecture: `.ai/ARCHITECTURE.md`.

## Local Development

```bash
# Backend
go run ./cmd/server    # API on :8080
go run ./cmd/mcp       # MCP stdio server

# Frontend
cd frontend && npm install && npm run dev  # Dashboard on :3000

# Full stack via Docker
make up       # All 6 services
make logs     # Tails Compose logs
make down     # Stop
```

## Making Changes

### Before Implementing

1. Identify affected modules and dependencies (`.ai/CODEMAP.md`).
2. Understand existing patterns (`internal/project/store.go` for new stores, `.ai/ARCHITECTURE_PRINCIPLES.md` for design rules).
3. Check `TODO.md` and `DECISIONS.md` — your work may already be planned or gated on an ADR.
4. If your change requires a new decision (canonical executor, browser backend, auth model, durability, scaling), propose an ADR first.

### Coding Conventions

- **Go:** Package-per-domain under `internal/`. New stores follow Store/MemoryStore/DBStore. Parameterized SQL only. UUIDs from `uuid_generate_v4()`. Colocated `*_test.go`.
- **Frontend:** TypeScript strict. API calls through `src/lib/api.ts` functions. Tailwind for styling. `cn()` helper from `src/lib/utils.ts`.
- **Python:** Snake-case node functions. `TypedDict` state contract. Pydantic request schemas.
- **Database:** Add new migrations in `internal/db/migrations/###_name.sql`, numbered sequentially. Never modify applied migrations.
- **Tests:** Colocated Go tests for new packages. HTTP handlers tested via `httptest` with memory stores.

### Commit Protocol

1. `gofmt -w .` — zero diff.
2. `go vet ./...` — zero findings.
3. `go test -count=1 ./...` — all pass.
4. `npm test && npm run lint` — if frontend changed.
5. `pytest` — if sidecar changed.
6. Update affected `.ai/` documents.
7. Append to `CHANGELOG_AI.md` for AI-generated modifications.
8. Small, focused commits with clear messages.

### After Implementing

1. Review the knowledge base — update only affected documents.
2. Mark relevant TODOs, risks, and ADRs.
3. Never rewrite unrelated sections.
4. Never delete historical ADR or changelog entries.
5. Report: files modified, why, risk, breaking changes, migrations, deployment steps, verification evidence.

## Testing Standards

- New Go packages: unit tests for memory store, integration tests for PostgreSQL store.
- New API endpoints: `httptest` cases covering happy path, auth, validation, and error responses.
- Frontend changes: component tests + at least one integration-level page test.
- Sidecar changes: pytest for endpoints and graph nodes.
- Current coverage inventory and gap analysis: `.ai/TESTING.md`.

## Design Decisions

Architecture decisions are tracked in `.ai/DECISIONS.md`. Five decisions are currently needed:

- `ADR-001` — canonical execution orchestrator
- `ADR-002` — canonical browser/test backend
- `ADR-003` — durable vs. ephemeral state
- `ADR-004` — deployment cardinality (single vs. multi-instance)
- `ADR-005` — authentication and authorization model

Do not assume resolutions to these in implementation. If your work depends on one, raise it.

## Getting Help

The knowledge base is designed to let a new engineer understand the project without rereading the entire repository. Start with `.ai/README.md`, then follow the document map to the specific area you need.

The frontend has its own contributor notes at `frontend/AGENTS.md` — this Next.js version differs from model training data, and local installed docs take priority.
