# Database

**Owner:** Engineering  
**Authoritative sources:** `internal/db/migrations/*.sql`, `internal/db/migrate.go`, PostgreSQL repository implementations  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-26  
**Verification performed:** Static schema/query inspection; no live PostgreSQL or migration execution  
**Confidence:** High for migration-declared schema; runtime planning queries and integration behavior partly UNKNOWN

## Engine and Configuration

- **Verified:** PostgreSQL 16.14 Compose image; pgx/v5 driver (`docker-compose.yml:57-74`; `go.mod:10`).
- **Verified:** Default config/Compose DSNs use user `postgres`, password `password`, and `sslmode=disable` (`internal/config/config.go:24-30`; `docker-compose.yml:9-12,59-65`).
- **Verified:** PostgreSQL data uses named `pgdata` volume; `/data` uses `app_data` (`docker-compose.yml:27-28,66-67,134-136`).
- **UNKNOWN:** Production credentials, TLS, backup encryption, RPO/RTO, and restore validation outside this repository.

## Schema Overview

### `schema_migrations`

- **Columns:** `version varchar(255) PRIMARY KEY`, `applied_at timestamptz DEFAULT now()`.
- **Purpose:** Migration application ledger, created by Go migration runner rather than a SQL migration.
- **Evidence:** `internal/db/migrate.go:21-29`.

### `projects`

- **PK:** `id uuid DEFAULT uuid_generate_v4()`.
- **Core columns:** required `name`; path/language/framework; `config jsonb`; intake fields `test_type`, `base_url`, `environment`, `spec`, `api_docs`, `auth_type`, `credentials`, focus/skip hints, `feature_map jsonb`; timestamps.
- **Indexes:** `created_at DESC`, `test_type`.
- **Evidence:** `001_init.sql:3-11`; `003_projects_intake.sql:1-14`.

### `test_runs`

- **PK:** `id uuid DEFAULT uuid_generate_v4()`.
- **FK:** nullable `project_id → projects(id)` with default NO ACTION behavior.
- **Columns:** state/mode, requirements, code analysis, JSONB plan/files/result/screenshots, fix/error/duration/token values, intake metadata/credentials, video metadata, nullable test-case/list IDs, timestamps.
- **Indexes:** state, project ID, created time descending, test-case ID, test-list ID.
- **Important:** `test_case_id` and `test_list_id` are not foreign keys.
- **Evidence:** `001_init.sql:13-36,64-76`; `002_testsprite_run_metadata.sql:1-9`; `006_run_test_links.sql:1-5`.

### `api_keys`

- **PK:** UUID.
- **Columns:** unique required hash, optional name, created/last-used timestamps.
- **Evidence:** `001_init.sql:38-44`.
- **Verified absence:** No repository/table consumer found; auth package only provides generation/hash helpers (`internal/auth/auth.go:97-114`).

### `settings`

- **PK:** string key.
- **Columns:** plaintext string value, updated timestamp.
- **Seed:** provider/model/API key/temperature/token/browser/fix-attempt defaults with `ON CONFLICT DO NOTHING`.
- **Evidence:** `001_init.sql:46-62`.

### `test_plan_drafts`

- **PK:** Application-supplied UUID (no DB default).
- **FK:** nullable `project_id → projects(id)`.
- **Columns:** status, required cases JSONB, timestamps.
- **Index:** project ID.
- **Evidence:** `004_test_planning_review.sql:1-8,26`.

### `test_cases`

- **PK:** Application-supplied UUID.
- **FKs:** nullable project and plan links.
- **Columns:** required title/type/priority, feature, required steps/assertions/tags JSONB, version, timestamps.
- **Indexes:** project ID, type; no plan ID index.
- **Evidence:** `004_test_planning_review.sql:10-28`.

### `test_lists`

- **PK:** Application-supplied UUID.
- **FK:** nullable project link.
- **Columns:** required name, tags JSONB, test-case IDs JSONB, pinned, timestamps.
- **Indexes:** project ID, pinned.
- **Evidence:** `005_test_lists.sql:1-13`.
- **Important:** JSON member IDs have no relational constraint to `test_cases`.

### `schedules`

- **PK:** UUID with DB default.
- **FKs/delete behavior:** project `ON DELETE SET NULL`; test list `ON DELETE CASCADE`; last run `ON DELETE SET NULL`.
- **Columns:** required name/frequency/timezone/enabled/next-run/timestamps; project/run target config; last-run status; failure notification and webhook.
- **Indexes:** `(enabled, next_run_at)`, test-list ID, project ID.
- **Evidence:** `007_schedules.sql:1-27`.

### `change_proposals`

- **PK:** UUID with DB default.
- **FK:** required test-case link `ON DELETE CASCADE`.
- **Columns:** status, prompt/rationale, required original/proposed JSONB, review metadata, timestamps.
- **Indexes:** `(test_case_id, created_at DESC)`, status.
- **Evidence:** `008_change_proposals.sql:1-17`.

## Relationship Summary

- `projects` optionally parents runs, drafts, cases, lists, and schedules.
- Drafts optionally parent cases.
- Cases strictly parent change proposals with cascade deletion.
- Test lists parent schedules with cascade deletion.
- Runs may be referenced by schedules as last run with `SET NULL`.
- Run→case/list and test-list→case membership are application-only IDs, not database-enforced relationships.

Domain meaning: [`DOMAIN.md`](DOMAIN.md).

## Migration Behavior

- **Verified:** SQL files are embedded, filtered to `.sql`, lexically sorted, checked against `schema_migrations`, executed, then recorded (`internal/db/migrate.go:14-69`).
- **Verified:** Each migration execution and version insert are separate operations; no explicit transaction/advisory lock exists (`internal/db/migrate.go:45-69`).
- **Verified:** Server logs migration failure and continues using PostgreSQL (`cmd/server/main.go:22-34`).
- **Verified absence:** No down migrations, checksums, drift detection, or rollback tooling in migration layer.
- **Verified:** `uuid-ossp` extension is required (`001_init.sql:1`).

## Seeds

- Settings seed: eight defaults with `ON CONFLICT DO NOTHING` (`001_init.sql:52-62`).
- Demo endpoint: five runs, in-memory events, one schedule, one in-memory release; repeated calls can duplicate data and insertion errors are ignored (`internal/api/server.go:3225-3334`).
- No standalone seed for API keys, projects, plans, cases, or lists found.

## Repository Mapping

| Entity | PostgreSQL | Memory | Evidence/notes |
|---|---|---|---|
| Runs | `db.Store` | `db.MemoryStore` | `internal/db/store.go`; `memory.go` |
| Settings | `SettingsStore` | None | Unavailable in fallback mode |
| Projects | `project.DBStore` | `project.MemoryStore` | Selected by concrete run-store type |
| Planning entities | `planning.MemoryStore` / `planning.DBStore` | Memory/PostgreSQL | Reconstructed 2026-07-26; DB store mirrors `project.DBStore` |
| Schedules | `schedule.DBStore` | `schedule.Store` | `internal/schedule/store.go` |
| Events | No | Always memory | Unbounded history/drop-prone subscribers |
| Recordings/visuals | No | Always memory | Restart loss |
| Releases/notifications/reviews/suites | No | Always memory | Restart loss |
| Sidecar jobs | No | Python process memory | Restart loss |

*Composition evidence:* `internal/api/server.go:62-85`; `cmd/server/main.go:19-37`.

## Transaction and Consistency Behavior

- **Verified:** Settings `SetMany` is the only explicit database transaction found (`internal/db/settings_store.go:54-72`).
- **Verified:** Migration DDL/version recording is non-atomic.
- **Verified:** Proposal approval changes case then proposal separately (`internal/api/server.go:850-899`).
- **Verified:** Plan approval creates cases then updates draft; update error is ignored (`internal/api/server.go:566-597`). Planning-store internals UNKNOWN.
- **Verified:** List execution creates runs one-by-one and may partially succeed (`internal/api/server.go:1131-1147`).
- **Verified:** Schedule run creation and schedule advancement are separate; due selection has no lock/claim (`internal/api/server.go:2671-2741`; `internal/schedule/store.go:245-269`).

## Known Issues and Potential Improvements

| Priority | Verified issue | Improvement direction |
|---|---|---|
| Critical | Planning implementation absent | Restore package and integration tests (`TODO-001`) |
| High | Migration and version insert non-atomic; startup continues after failure | Per-migration transaction, advisory lock, fail startup |
| High | Run deletion absent | Add repository deletion and artifact cleanup (`TODO-009`) |
| High | Credentials/settings stored plaintext | Decide secret-reference/encryption model (`ADR-005`) |
| High | Non-atomic due-schedule claim | Lease/locking after `ADR-004` (`TODO-017`) |
| High | Application-only relationship integrity | Evaluate FKs/junction table based on lifecycle decisions |
| Medium | No CHECK constraints for state/type/priority/frequency/status | Add forward migration only if required and backward-compatible |
| Medium | Timezone stored but ignored by next-run calculation | Apply timezone-aware calculation and tests |
| Medium | Store code ignores some JSON/rows/update errors | Propagate and test failures |
| Medium | PostgreSQL integration tests not found | Add migration/repository/concurrency tests |
| Medium | No backup/restore/retention process found | Define RPO/RTO then implement/verify |

## UNKNOWN

- Planning queries, transactions, and repository error behavior.
- Production DB topology, privileges, TLS, backups, retention, volume encryption, and migration concurrency.
- Required deletion semantics for project dependents.
- Required durability for operational stores.
- Production data volume/cardinality and resulting index needs.
