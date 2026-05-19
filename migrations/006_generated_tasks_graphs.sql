BEGIN;

ALTER TABLE generated_tasks
    ADD COLUMN IF NOT EXISTS graphs JSONB;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'generated_tasks_graphs_array_chk'
    ) THEN
        ALTER TABLE generated_tasks
            ADD CONSTRAINT generated_tasks_graphs_array_chk
            CHECK (graphs IS NULL OR jsonb_typeof(graphs) = 'array');
    END IF;
END $$;

COMMIT;
