CREATE TABLE IF NOT EXISTS test_lists (
    id            UUID PRIMARY KEY,
    name          TEXT NOT NULL,
    project_id    UUID REFERENCES projects(id),
    tags          JSONB NOT NULL DEFAULT '[]',
    test_case_ids JSONB NOT NULL DEFAULT '[]',
    pinned        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_test_lists_project ON test_lists(project_id);
CREATE INDEX IF NOT EXISTS idx_test_lists_pinned ON test_lists(pinned);
