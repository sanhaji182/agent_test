-- 017_fix_run_events_run_id: fix run_events.run_id type.
-- Migrasi 009 originally created run_id as VARCHAR(32), but run IDs are UUIDs
-- (36 chars), so every INSERT failed with "value too long for type character
-- varying(32)". Existing databases keep the wrong column type, so this
-- migration widens it to UUID (with FK to test_runs) to match the fixed 009.
-- Up
ALTER TABLE run_events
    ALTER COLUMN run_id TYPE UUID USING run_id::uuid;

-- Add FK constraint if not present (idempotent guard via DO block).
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'run_events_run_id_fkey'
    ) THEN
        ALTER TABLE run_events
            ADD CONSTRAINT run_events_run_id_fkey
            FOREIGN KEY (run_id) REFERENCES test_runs(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Down
-- ALTER TABLE run_events DROP CONSTRAINT IF EXISTS run_events_run_id_fkey;
-- ALTER TABLE run_events ALTER COLUMN run_id TYPE VARCHAR(32) USING substring(run_id::text, 1, 32);
