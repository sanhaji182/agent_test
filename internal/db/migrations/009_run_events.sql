-- 009_run_events: Persist execution events for SSE replay and restart durability.
-- ADR-003 Phase 1 — event persistence.
-- Up
CREATE TABLE IF NOT EXISTS run_events (
    id          BIGSERIAL PRIMARY KEY,
    run_id      VARCHAR(32)  NOT NULL,
    event_type  VARCHAR(64)  NOT NULL,
    phase       VARCHAR(64)  NOT NULL DEFAULT '',
    message     TEXT         NOT NULL DEFAULT '',
    metadata    JSONB        NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_run_events_run_id ON run_events (run_id);
CREATE INDEX idx_run_events_created_at ON run_events (created_at);
CREATE INDEX idx_run_events_type ON run_events (run_id, event_type);

-- Down
-- DROP TABLE IF EXISTS run_events;
