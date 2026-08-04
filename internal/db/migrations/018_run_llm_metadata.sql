-- Run-level LLM metadata: mencatat provider & model yang benar-benar dipakai
-- untuk setiap run (termasuk fallback), supaya UI & laporan bisa menampilkan
-- "model mana yang dipakai" dan pengguna bisa mendeteksi jika model berubah-ubah.
ALTER TABLE test_runs ADD COLUMN IF NOT EXISTS llm_provider TEXT;
ALTER TABLE test_runs ADD COLUMN IF NOT EXISTS llm_model TEXT;
ALTER TABLE test_runs ADD COLUMN IF NOT EXISTS llm_fallback_provider TEXT;
ALTER TABLE test_runs ADD COLUMN IF NOT EXISTS llm_fallback_model TEXT;
