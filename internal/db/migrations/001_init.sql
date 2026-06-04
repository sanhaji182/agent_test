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
    video_url      TEXT,
    video_status   VARCHAR(50),
    video_duration REAL,
    video_size     BIGINT,
    video_failure_marker_at REAL,
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

CREATE TABLE IF NOT EXISTS settings (
    key        VARCHAR(100) PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Default LLM settings
INSERT INTO settings (key, value) VALUES
  ('llm_provider', 'anthropic'),
  ('llm_model', 'claude-sonnet-4-5'),
  ('llm_api_key', ''),
  ('llm_temperature', '0.2'),
  ('llm_max_tokens', '4096'),
  ('browser_headless', 'true'),
  ('browser_timeout', '300'),
  ('max_fix_attempts', '3')
ON CONFLICT (key) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_runs_state ON test_runs(state);
CREATE INDEX IF NOT EXISTS idx_runs_project ON test_runs(project_id);
CREATE INDEX IF NOT EXISTS idx_runs_created ON test_runs(created_at DESC);

ALTER TABLE test_runs ADD COLUMN IF NOT EXISTS project_path TEXT;
ALTER TABLE test_runs ADD COLUMN IF NOT EXISTS test_type VARCHAR(20);
ALTER TABLE test_runs ADD COLUMN IF NOT EXISTS prd TEXT;
ALTER TABLE test_runs ADD COLUMN IF NOT EXISTS api_docs TEXT;
ALTER TABLE test_runs ADD COLUMN IF NOT EXISTS auth_type VARCHAR(50);
ALTER TABLE test_runs ADD COLUMN IF NOT EXISTS credentials TEXT;
ALTER TABLE test_runs ADD COLUMN IF NOT EXISTS focus_hints TEXT;
ALTER TABLE test_runs ADD COLUMN IF NOT EXISTS skip_hints TEXT;
ALTER TABLE test_runs ADD COLUMN IF NOT EXISTS feature_map JSONB;
