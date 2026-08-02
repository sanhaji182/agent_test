-- 015_api_keys_extend: tambah kolom untuk multi-user API key management
-- (role, active, created_by) supaya API key bisa disimpan persistent di DB.
-- Up
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS role VARCHAR(32) NOT NULL DEFAULT 'viewer';
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS active BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS created_by VARCHAR(64) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys (active);

-- Down
-- ALTER TABLE api_keys DROP COLUMN IF EXISTS role;
-- ALTER TABLE api_keys DROP COLUMN IF EXISTS active;
-- ALTER TABLE api_keys DROP COLUMN IF EXISTS created_by;
