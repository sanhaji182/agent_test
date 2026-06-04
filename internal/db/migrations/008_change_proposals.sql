CREATE TABLE IF NOT EXISTS change_proposals (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    test_case_id   UUID NOT NULL REFERENCES test_cases(id) ON DELETE CASCADE,
    status         TEXT NOT NULL DEFAULT 'pending',
    prompt         TEXT NOT NULL DEFAULT '',
    rationale      TEXT NOT NULL DEFAULT '',
    original       JSONB NOT NULL,
    proposed       JSONB NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at    TIMESTAMPTZ,
    reviewer       TEXT DEFAULT '',
    review_comment TEXT DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_change_proposals_test_case ON change_proposals(test_case_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_change_proposals_status ON change_proposals(status);
