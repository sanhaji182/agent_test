-- 010_drifts: Phase 3 continuous sync — webhook registrations, drift records,
-- and tests auto-generated from drifts (docs/PHASE-3-IMPLEMENTATION.md).
-- Up
CREATE TABLE IF NOT EXISTS webhook_registrations (
    id              VARCHAR(64)  PRIMARY KEY,
    user_id         VARCHAR(64)  NOT NULL,
    repository_id   VARCHAR(64)  NOT NULL DEFAULT '',
    repository_url  VARCHAR(500) NOT NULL,
    webhook_id      VARCHAR(255) NOT NULL,
    status          VARCHAR(50)  NOT NULL DEFAULT 'active',
    last_sync_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_registrations_user ON webhook_registrations (user_id);
CREATE INDEX idx_webhook_registrations_repo ON webhook_registrations (repository_url);

CREATE TABLE IF NOT EXISTS drifts (
    id             VARCHAR(64)  PRIMARY KEY,
    repository_id  VARCHAR(255) NOT NULL,
    type           VARCHAR(50)  NOT NULL,
    file_path      VARCHAR(500) NOT NULL,
    description    TEXT         NOT NULL DEFAULT '',
    severity       VARCHAR(50)  NOT NULL,
    status         VARCHAR(50)  NOT NULL DEFAULT 'pending',
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_drifts_repository ON drifts (repository_id);
CREATE INDEX idx_drifts_status ON drifts (status);
CREATE INDEX idx_drifts_severity ON drifts (severity);

CREATE TABLE IF NOT EXISTS drift_generated_tests (
    id               VARCHAR(64)  PRIMARY KEY,
    drift_id         VARCHAR(64)  REFERENCES drifts(id),
    test_name        VARCHAR(255) NOT NULL,
    test_code        TEXT         NOT NULL,
    confidence_score INTEGER,
    status           VARCHAR(50)  NOT NULL DEFAULT 'generated',
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_drift_generated_tests_drift ON drift_generated_tests (drift_id);

-- Down
-- DROP TABLE IF EXISTS drift_generated_tests;
-- DROP TABLE IF EXISTS drifts;
-- DROP TABLE IF EXISTS webhook_registrations;
