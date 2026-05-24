BEGIN;

ALTER TABLE IF EXISTS prepared_tasks RENAME TO prepared_tasks_old;

-- New prepared_tasks table for reusable generated tasks.
CREATE TABLE IF NOT EXISTS prepared_tasks (
    id BIGSERIAL PRIMARY KEY,
    mode TEXT NOT NULL DEFAULT 'custom',
    task_type_id BIGINT REFERENCES task_types(id) ON DELETE SET NULL,
    oge_number INT NOT NULL,
    subtype_code TEXT NOT NULL,
    question TEXT NOT NULL,
    correct_answer TEXT NOT NULL,
    solution_steps JSONB,
    self_check TEXT,
    is_valid BOOLEAN NOT NULL DEFAULT false,
    validation_notes TEXT,
    visual_data JSONB,
    graphs JSONB,
    generation_source TEXT NOT NULL DEFAULT 'gpt-4o-mini',
    used BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT prepared_tasks_mode_chk CHECK (mode IN ('weak', 'all', 'custom', 'diagnostic')),
    CONSTRAINT prepared_tasks_oge_number_chk CHECK (oge_number BETWEEN 6 AND 19),
    CONSTRAINT prepared_tasks_subtype_code_not_empty_chk CHECK (length(trim(subtype_code)) > 0),
    CONSTRAINT prepared_tasks_visual_data_object_chk CHECK (visual_data IS NULL OR jsonb_typeof(visual_data) = 'object')
);

CREATE INDEX IF NOT EXISTS prepared_tasks_used_oge_subtype_idx
    ON prepared_tasks (used, oge_number, subtype_code);

CREATE INDEX IF NOT EXISTS prepared_tasks_task_type_id_idx
    ON prepared_tasks (task_type_id);

CREATE INDEX IF NOT EXISTS prepared_tasks_created_at_idx
    ON prepared_tasks (created_at);

DROP TABLE IF EXISTS prepared_tasks_old;

COMMIT;
