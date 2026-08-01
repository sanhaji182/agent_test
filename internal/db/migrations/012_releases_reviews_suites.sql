-- 012_releases_reviews_suites: persist release, review, and suite workflow metadata.
-- Up
CREATE TABLE IF NOT EXISTS releases (
    id         VARCHAR(64)  PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    version    VARCHAR(128) NOT NULL DEFAULT '',
    project_id VARCHAR(64)  NOT NULL DEFAULT '',
    status     VARCHAR(50)  NOT NULL DEFAULT 'active',
    run_ids    JSONB        NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_releases_project_id ON releases (project_id);
CREATE INDEX IF NOT EXISTS idx_releases_status ON releases (status);
CREATE INDEX IF NOT EXISTS idx_releases_created_at ON releases (created_at DESC);

CREATE TABLE IF NOT EXISTS reviews (
    id         VARCHAR(64)  PRIMARY KEY,
    run_id     VARCHAR(64)  NOT NULL,
    type       VARCHAR(64)  NOT NULL,
    status     VARCHAR(50)  NOT NULL DEFAULT 'pending',
    reviewer   VARCHAR(255) NOT NULL DEFAULT '',
    comment    TEXT         NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reviews_run_id ON reviews (run_id);
CREATE INDEX IF NOT EXISTS idx_reviews_status ON reviews (status);
CREATE INDEX IF NOT EXISTS idx_reviews_created_at ON reviews (created_at DESC);

CREATE TABLE IF NOT EXISTS suites (
    id          VARCHAR(64)  PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    project_id  VARCHAR(64)  NOT NULL DEFAULT '',
    environment VARCHAR(64)  NOT NULL DEFAULT '',
    tags        JSONB        NOT NULL DEFAULT '[]',
    pinned      BOOLEAN      NOT NULL DEFAULT FALSE,
    run_ids     JSONB        NOT NULL DEFAULT '[]',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_suites_project_id ON suites (project_id);
CREATE INDEX IF NOT EXISTS idx_suites_pinned ON suites (pinned);
CREATE INDEX IF NOT EXISTS idx_suites_tags ON suites USING GIN (tags);
CREATE INDEX IF NOT EXISTS idx_suites_created_at ON suites (created_at DESC);

-- Down
-- DROP TABLE IF EXISTS suites;
-- DROP TABLE IF EXISTS reviews;
-- DROP TABLE IF EXISTS releases;
