-- 013_audit_log: append-only audit trail for RBAC and security events.
-- Up
CREATE TABLE IF NOT EXISTS audit_log (
    id          VARCHAR(64)  PRIMARY KEY,
    actor_id    VARCHAR(64)  NOT NULL,
    actor_role  VARCHAR(32)  NOT NULL DEFAULT '',
    action      VARCHAR(64)  NOT NULL,
    resource    VARCHAR(64)  NOT NULL,
    resource_id VARCHAR(128) NOT NULL DEFAULT '',
    detail      TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_actor ON audit_log (actor_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_resource ON audit_log (resource, resource_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_log (action, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_created_at ON audit_log (created_at DESC);

-- Down
-- DROP TABLE IF EXISTS audit_log;
