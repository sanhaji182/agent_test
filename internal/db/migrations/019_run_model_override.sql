-- Per-run model override: pengguna bisa memaksa model LLM tertentu untuk run
-- tertentu (prioritas di atas profile/settings/env). Dipakai agar user dengan
-- beberapa model bisa memilih model yang jelas per test — tidak "berubah-ubah".
ALTER TABLE test_runs ADD COLUMN IF NOT EXISTS model_override TEXT;
