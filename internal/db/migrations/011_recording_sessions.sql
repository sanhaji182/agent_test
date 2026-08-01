-- 011_recording_sessions: Phase 2 Record & Playback — recording sessions and recorded events.
-- Up
CREATE TABLE IF NOT EXISTS recording_sessions (
    id           VARCHAR(64)  PRIMARY KEY,
    name         VARCHAR(255) NOT NULL,
    project_path VARCHAR(500) NOT NULL,
    base_url     VARCHAR(500) NOT NULL DEFAULT '',
    status       VARCHAR(50)  NOT NULL DEFAULT 'recording',
    metadata     JSONB        NOT NULL DEFAULT '{}',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_recording_sessions_status ON recording_sessions (status);
CREATE INDEX idx_recording_sessions_created ON recording_sessions (created_at DESC);

CREATE TABLE IF NOT EXISTS recorded_events (
    id             VARCHAR(64)  PRIMARY KEY,
    session_id     VARCHAR(64)  NOT NULL REFERENCES recording_sessions(id) ON DELETE CASCADE,
    event_type     VARCHAR(50)  NOT NULL,
    selector       VARCHAR(500) NOT NULL DEFAULT '',
    value          TEXT         NOT NULL DEFAULT '',
    url            VARCHAR(500) NOT NULL DEFAULT '',
    timestamp      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    metadata       JSONB        NOT NULL DEFAULT '{}',
    sequence_order INTEGER      NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_recorded_events_session ON recorded_events (session_id, sequence_order);
CREATE INDEX idx_recorded_events_type ON recorded_events (session_id, event_type);

-- Down
-- DROP TABLE IF EXISTS recorded_events;
-- DROP TABLE IF EXISTS recording_sessions;
