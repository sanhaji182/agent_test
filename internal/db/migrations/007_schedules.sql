CREATE TABLE IF NOT EXISTS schedules (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id      UUID REFERENCES projects(id) ON DELETE SET NULL,
    test_list_id    UUID REFERENCES test_lists(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    project_path    TEXT DEFAULT '',
    requirements    TEXT DEFAULT '',
    mode            TEXT DEFAULT 'simple',
    environment     TEXT DEFAULT '',
    base_url        TEXT DEFAULT '',
    frequency       TEXT NOT NULL DEFAULT 'daily',
    cron_expr       TEXT DEFAULT '',
    timezone        TEXT NOT NULL DEFAULT 'UTC',
    enabled         BOOLEAN NOT NULL DEFAULT FALSE,
    next_run_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_run_at     TIMESTAMPTZ,
    last_run_id     UUID REFERENCES test_runs(id) ON DELETE SET NULL,
    last_run_status TEXT DEFAULT '',
    notify_on_fail  BOOLEAN NOT NULL DEFAULT FALSE,
    webhook_url     TEXT DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_schedules_due ON schedules(enabled, next_run_at);
CREATE INDEX IF NOT EXISTS idx_schedules_test_list ON schedules(test_list_id);
CREATE INDEX IF NOT EXISTS idx_schedules_project ON schedules(project_id);
