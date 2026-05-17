BEGIN;

ALTER TABLE generated_tasks
    ADD COLUMN IF NOT EXISTS visual_data JSONB;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'generated_tasks_visual_data_object_chk'
    ) THEN
        ALTER TABLE generated_tasks
            ADD CONSTRAINT generated_tasks_visual_data_object_chk
            CHECK (visual_data IS NULL OR jsonb_typeof(visual_data) = 'object');
    END IF;
END $$;

COMMIT;
