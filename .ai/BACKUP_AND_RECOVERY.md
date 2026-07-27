# Backup and Recovery

**Owner:** Engineering (SRE/operations ownership UNKNOWN)  
**Authoritative sources:** Compose volumes, video/artifact paths, migration code, documentation  
**Last verified revision:** `7b54053642e614cccf5e1128defabd25ac88b437`  
**Last updated:** 2026-07-27  
**Verification performed:** Static inspection; no backup/restore was observed or tested  
**Confidence:** High for what is absent; recommended procedures are templated, not tested

## Current State

**No backup, restore, retention, or disaster recovery implementation exists in the tracked repository.** The only documented data operation is destructive volume reset:

> "To reset everything: `docker compose down -v && docker compose up -d`"  
> — `docs/docker.md:97-101`

### What Data Exists

| Data | Storage | Survives Container Restart? | Survives Volume Loss? |
|---|---|---|---|
| PostgreSQL (runs, projects, settings, schedules, planning) | Named volume `pgdata` (`docker-compose.yml:134-136`) | Yes | No |
| Application files (screenshots, reports) | Named volume `app_data` (`docker-compose.yml:27-28`) | Yes | No |
| Events, recordings metadata, visual metadata, releases, notifications, reviews, suites | Process memory (`internal/api/server.go:71-85`) | No | No |
| Sidecar jobs | Python process memory (`sidecar/main.py:11-12`) | No | No |
| Temporary Playwright videos | `/tmp/agent_test/videos` (`internal/api/server.go:1923`) | No | No |
| Docker runner videos | `/data/videos/{runID}` (`internal/runner/docker.go:98-100`) | Yes (if `/data` is the `app_data` volume) | No |

## Recommended Backup Architecture

### Tier 1 — Critical (PostgreSQL)

PostgreSQL is the only durable store for application state. Without it, all runs, projects, plans, cases, lists, and settings are lost.

**Implementation:**

```bash
# Daily logical backup
pg_dump -h localhost -U postgres -d gotest_agent \
  --format=custom --file="/data/backups/gotest-$(date +%Y%m%d).dump"
```

**Requirements:**
- Schedule: at least daily; more frequent if RPO requires it.
- Retention: defined by business requirement (suggested: 30 daily, 12 monthly).
- Storage: backup files must be stored off-host (object storage, network volume, or replicated location).
- Verification: restore a backup to a temporary database and verify schema/data integrity at least weekly.
- Encryption: backup files should be encrypted at rest if they contain credential fields (currently plaintext; see `ADR-005`).

**Recovery procedure:**

```bash
# 1. Stop the application
docker compose stop app

# 2. Restore
pg_restore -h localhost -U postgres -d gotest_agent \
  --clean --if-exists /path/to/backup.dump

# 3. Run migrations to ensure schema is current
# (the application does this on startup)

# 4. Start the application
docker compose up -d

# 5. Verify
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/runs
```

### Tier 2 — Application Files

`/data` holds screenshots and persistent video recordings.

**Implementation:**
- Include `/data` in the same backup schedule as PostgreSQL, or
- Use `rsync`/`rclone` to sync to object storage.
- Define retention policy (e.g., artifacts older than 90 days are archived).

### Tier 3 — Volatile State

Events, recordings metadata, visual artifacts, releases, notifications, reviews, suites, and sidecar jobs are lost on restart by design in the current architecture. This needs an owner decision (`ADR-003`).

**Options:**
1. Persist to PostgreSQL and include in backups.
2. Persist to a separate event/log store with its own backup.
3. Document as intentionally ephemeral (acceptable for demo/development, not for production audit trails).

## Retention

| Data | Suggested Retention | Rationale |
|---|---|---|
| Run results + video | 90 days | Audit evidence; storage cost of video |
| Events | 30 days | Debugging value decays quickly |
| Screenshots | 90 days | Matches run retention |
| Releases | Indefinite | Small data; release history |
| Projects, cases, lists | Indefinite | Configuration, not execution data |
| Backups | 30 daily + 12 monthly | Standard tiered retention |

## Migration Safety During Recovery

**Current risk:** Migrations are non-atomic (`internal/db/migrate.go:45-69`) and failure does not block startup (`cmd/server/main.go:28-34`). During recovery:

1. Restore the database.
2. Start the application with migrations enabled.
3. If migrations fail, investigate the `schema_migrations` table before retrying.
4. Never run migrations manually against a restored database without recording the version.

**Improvement needed:** Add a migration dry-run mode and per-migration transactions. See `TODO-017` (schedule claiming) and the migration improvements listed in `.ai/DATABASE.md`.

## Disaster Recovery Scenarios

### Scenario 1: Host failure, volumes intact

1. Provision new host with Docker and Compose.
2. Mount existing `pgdata` and `app_data` volumes.
3. Start Compose. Application resumes with existing data.
4. Volatile state (events, sidecar jobs, notifications) is lost — this is current expected behavior.

### Scenario 2: Volume corruption or loss

1. Provision new host.
2. Restore latest PostgreSQL backup.
3. Sync application files from backup.
4. Start Compose.
5. All volatile state is lost; all durable state is at the backup point.

### Scenario 3: Accidental data deletion

1. Identify the deletion timestamp.
2. Restore the most recent backup before that timestamp to a temporary database.
3. Extract affected rows and re-insert into production.
4. For run data: re-running is usually preferable to restoring individual runs.

## Verification Checklist

Before claiming backup/recovery readiness:

- [ ] Automated daily PostgreSQL backup running and monitored.
- [ ] Backup files stored off-host.
- [ ] Restore procedure tested end-to-end at least once.
- [ ] Restore time measured and within RTO.
- [ ] `/data` backup included or explicitly excluded by policy.
- [ ] Migration compatibility verified against restored backups.
- [ ] Encryption at rest confirmed if credential fields are stored.
- [ ] Retention and archival policies documented.
- [ ] Recovery runbook accessible to operators.
