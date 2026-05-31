CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS projects (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       VARCHAR(255) NOT NULL,
    path       TEXT,
    language   VARCHAR(50),
    framework  VARCHAR(50),
    config     JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS test_runs (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id     UUID REFERENCES projects(id),
    state          VARCHAR(50) NOT NULL DEFAULT 'idle',
    mode           VARCHAR(20) DEFAULT 'simple',
    requirements   TEXT,
    code_analysis  TEXT,
    test_plan      JSONB,
    test_files     JSONB,
    run_result     JSONB,
    screenshots    JSONB,
    fix_attempts   INT DEFAULT 0,
    error_msg      TEXT,
    duration_ms    INT,
    llm_tokens_used INT,
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    updated_at     TIMESTAMPTZ DEFAULT NOW(),
    finished_at    TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS api_keys (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key_hash     VARCHAR(255) UNIQUE NOT NULL,
    name         VARCHAR(100),
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    last_used_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_runs_state ON test_runs(state);
CREATE INDEX IF NOT EXISTS idx_runs_project ON test_runs(project_id);
CREATE INDEX IF NOT EXISTS idx_runs_created ON test_runs(created_at DESC);
