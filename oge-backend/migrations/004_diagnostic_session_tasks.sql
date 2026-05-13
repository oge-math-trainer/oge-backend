BEGIN;

-- Table to track which tasks belong to which diagnostic session
CREATE TABLE IF NOT EXISTS diagnostic_session_tasks (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES diagnostic_sessions(id) ON DELETE CASCADE,
    generated_task_id BIGINT NOT NULL REFERENCES generated_tasks(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT diagnostic_session_tasks_session_task_uidx UNIQUE (session_id, generated_task_id)
);

CREATE INDEX IF NOT EXISTS diagnostic_session_tasks_session_id_idx
    ON diagnostic_session_tasks (session_id);

CREATE INDEX IF NOT EXISTS diagnostic_session_tasks_generated_task_id_idx
    ON diagnostic_session_tasks (generated_task_id);

COMMIT;
