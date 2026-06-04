ALTER TABLE test_runs ADD COLUMN IF NOT EXISTS test_case_id UUID;
ALTER TABLE test_runs ADD COLUMN IF NOT EXISTS test_list_id UUID;

CREATE INDEX IF NOT EXISTS idx_runs_test_case ON test_runs(test_case_id);
CREATE INDEX IF NOT EXISTS idx_runs_test_list ON test_runs(test_list_id);
