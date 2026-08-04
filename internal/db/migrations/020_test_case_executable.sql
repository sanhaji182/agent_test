-- Deterministic test cases: simpan konten executable (browser-action JSON)
-- langsung pada test case supaya case dari hasil AI/recorder bisa dijalankan
-- ulang secara deterministik tanpa regenerate AI (mirip rekam-putar Katalon).
ALTER TABLE test_cases ADD COLUMN IF NOT EXISTS executable_content TEXT;
-- source_run_id menandai run asal (untuk dedupe auto-save + link kembali).
ALTER TABLE test_cases ADD COLUMN IF NOT EXISTS source_run_id UUID;
CREATE INDEX IF NOT EXISTS idx_test_cases_source_run ON test_cases(source_run_id);
