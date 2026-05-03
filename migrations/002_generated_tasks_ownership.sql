BEGIN;

ALTER TABLE generated_tasks
    ADD COLUMN IF NOT EXISTS user_id BIGINT REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE generated_tasks
    ADD COLUMN IF NOT EXISTS mode TEXT NOT NULL DEFAULT 'custom';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'generated_tasks_mode_chk'
    ) THEN
        ALTER TABLE generated_tasks
            ADD CONSTRAINT generated_tasks_mode_chk
            CHECK (mode IN ('weak', 'all', 'custom', 'diagnostic'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS generated_tasks_user_id_idx
    ON generated_tasks (user_id);

CREATE INDEX IF NOT EXISTS generated_tasks_user_mode_idx
    ON generated_tasks (user_id, mode);

COMMIT;
