BEGIN;

ALTER TABLE prepared_tasks
    ADD COLUMN IF NOT EXISTS shown_count INT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS prepared_task_views (
    prepared_task_id BIGINT NOT NULL REFERENCES prepared_tasks(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT prepared_task_views_task_user_uidx UNIQUE (prepared_task_id, user_id)
);

UPDATE prepared_tasks
SET shown_count = 20
WHERE used = true
  AND shown_count < 20;

DELETE FROM prepared_tasks
WHERE shown_count >= 20;

DROP INDEX IF EXISTS prepared_tasks_used_oge_subtype_idx;

CREATE INDEX IF NOT EXISTS prepared_tasks_available_oge_subtype_idx
    ON prepared_tasks (oge_number, subtype_code, shown_count, created_at)
    WHERE used = false AND shown_count < 20;

CREATE INDEX IF NOT EXISTS prepared_task_views_user_id_idx
    ON prepared_task_views (user_id);

COMMIT;
