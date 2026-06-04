CREATE TABLE IF NOT EXISTS test_plan_drafts (
    id         UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id),
    status     VARCHAR(40) NOT NULL DEFAULT 'draft',
    cases      JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS test_cases (
    id         UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id),
    plan_id    UUID REFERENCES test_plan_drafts(id),
    title      TEXT NOT NULL,
    type       VARCHAR(20) NOT NULL DEFAULT 'ui',
    feature    TEXT,
    priority   VARCHAR(20) NOT NULL DEFAULT 'medium',
    steps      JSONB NOT NULL DEFAULT '[]',
    assertions JSONB NOT NULL DEFAULT '[]',
    tags       JSONB NOT NULL DEFAULT '[]',
    version    INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_test_plan_drafts_project ON test_plan_drafts(project_id);
CREATE INDEX IF NOT EXISTS idx_test_cases_project ON test_cases(project_id);
CREATE INDEX IF NOT EXISTS idx_test_cases_type ON test_cases(type);
