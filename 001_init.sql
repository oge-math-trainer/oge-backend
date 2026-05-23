BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_email_not_empty_chk CHECK (length(trim(email)) > 0),
    CONSTRAINT users_password_hash_not_empty_chk CHECK (length(trim(password_hash)) > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_lower_uidx
    ON users (lower(email));

CREATE TABLE IF NOT EXISTS topics ( 
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT topics_code_not_empty_chk CHECK (length(trim(code)) > 0),
    CONSTRAINT topics_title_not_empty_chk CHECK (length(trim(title)) > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS topics_code_uidx
    ON topics (code);

CREATE TABLE IF NOT EXISTS subtopics (
    id BIGSERIAL PRIMARY KEY,
    topic_id BIGINT NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT subtopics_code_not_empty_chk CHECK (length(trim(code)) > 0),
    CONSTRAINT subtopics_title_not_empty_chk CHECK (length(trim(title)) > 0),
    CONSTRAINT subtopics_id_topic_id_uidx UNIQUE (id, topic_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS subtopics_topic_code_uidx
    ON subtopics (topic_id, code);

CREATE INDEX IF NOT EXISTS subtopics_topic_id_idx
    ON subtopics (topic_id);

CREATE TABLE IF NOT EXISTS task_types (
    id BIGSERIAL PRIMARY KEY,
    oge_number INT NOT NULL,
    topic_id BIGINT NOT NULL,
    subtopic_id BIGINT NOT NULL,
    subtype_code TEXT NOT NULL,
    title TEXT,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT task_types_oge_number_chk CHECK (oge_number BETWEEN 6 AND 19),
    CONSTRAINT task_types_subtype_code_not_empty_chk CHECK (length(trim(subtype_code)) > 0),
    CONSTRAINT task_types_topic_id_fk FOREIGN KEY (topic_id)
        REFERENCES topics(id) ON DELETE RESTRICT,
    CONSTRAINT task_types_subtopic_topic_fk FOREIGN KEY (subtopic_id, topic_id)
        REFERENCES subtopics(id, topic_id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX IF NOT EXISTS task_types_oge_subtype_uidx
    ON task_types (oge_number, subtype_code);

CREATE UNIQUE INDEX IF NOT EXISTS task_types_oge_subtopic_uidx
    ON task_types (oge_number, subtopic_id);

CREATE INDEX IF NOT EXISTS task_types_topic_id_idx
    ON task_types (topic_id);

CREATE INDEX IF NOT EXISTS task_types_subtopic_id_idx
    ON task_types (subtopic_id);

CREATE TABLE IF NOT EXISTS generated_tasks (
    id BIGSERIAL PRIMARY KEY,
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
    generation_source TEXT NOT NULL DEFAULT 'qwen',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT generated_tasks_oge_number_chk CHECK (oge_number BETWEEN 6 AND 19),
    CONSTRAINT generated_tasks_subtype_code_not_empty_chk CHECK (length(trim(subtype_code)) > 0),
    CONSTRAINT generated_tasks_visual_data_object_chk CHECK (visual_data IS NULL OR jsonb_typeof(visual_data) = 'object')
);

CREATE INDEX IF NOT EXISTS generated_tasks_oge_subtype_idx
    ON generated_tasks (oge_number, subtype_code);

CREATE INDEX IF NOT EXISTS generated_tasks_task_type_id_idx
    ON generated_tasks (task_type_id);

CREATE INDEX IF NOT EXISTS generated_tasks_created_at_idx
    ON generated_tasks (created_at);

CREATE TABLE IF NOT EXISTS fallback_tasks (
    id BIGSERIAL PRIMARY KEY,
    task_type_id BIGINT REFERENCES task_types(id) ON DELETE SET NULL,
    oge_number INT NOT NULL,
    subtype_code TEXT NOT NULL,
    question TEXT NOT NULL,
    correct_answer TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fallback_tasks_oge_number_chk CHECK (oge_number BETWEEN 6 AND 19),
    CONSTRAINT fallback_tasks_subtype_code_not_empty_chk CHECK (length(trim(subtype_code)) > 0)
);

CREATE INDEX IF NOT EXISTS fallback_tasks_oge_subtype_idx
    ON fallback_tasks (oge_number, subtype_code);

CREATE INDEX IF NOT EXISTS fallback_tasks_task_type_id_idx
    ON fallback_tasks (task_type_id);

CREATE INDEX IF NOT EXISTS fallback_tasks_created_at_idx
    ON fallback_tasks (created_at);

CREATE TABLE IF NOT EXISTS diagnostic_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'started',
    ai_analysis JSONB,
    weak_topics JSONB,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS diagnostic_sessions_user_id_idx
    ON diagnostic_sessions (user_id);

CREATE INDEX IF NOT EXISTS diagnostic_sessions_started_at_idx
    ON diagnostic_sessions (started_at);

CREATE TABLE IF NOT EXISTS diagnostic_answers (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES diagnostic_sessions(id) ON DELETE CASCADE,
    generated_task_id BIGINT NOT NULL REFERENCES generated_tasks(id) ON DELETE RESTRICT,
    student_answer TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL DEFAULT false,
    ai_feedback JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS diagnostic_answers_session_id_idx
    ON diagnostic_answers (session_id);

CREATE INDEX IF NOT EXISTS diagnostic_answers_generated_task_id_idx
    ON diagnostic_answers (generated_task_id);

CREATE INDEX IF NOT EXISTS diagnostic_answers_created_at_idx
    ON diagnostic_answers (created_at);

CREATE TABLE IF NOT EXISTS attempts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    generated_task_id BIGINT NOT NULL REFERENCES generated_tasks(id) ON DELETE RESTRICT,
    mode TEXT NOT NULL,
    student_answer TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL DEFAULT false,
    ai_feedback JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT attempts_mode_chk CHECK (mode IN ('weak', 'all', 'custom', 'diagnostic'))
);

CREATE INDEX IF NOT EXISTS attempts_user_id_idx
    ON attempts (user_id);

CREATE INDEX IF NOT EXISTS attempts_generated_task_id_idx
    ON attempts (generated_task_id);

CREATE INDEX IF NOT EXISTS attempts_mode_idx
    ON attempts (mode);

CREATE INDEX IF NOT EXISTS attempts_created_at_idx
    ON attempts (created_at);

CREATE TABLE IF NOT EXISTS progress (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_type_id BIGINT REFERENCES task_types(id) ON DELETE SET NULL,
    oge_number INT NOT NULL,
    subtype_code TEXT NOT NULL,
    attempts_count INT NOT NULL DEFAULT 0,
    correct_count INT NOT NULL DEFAULT 0,
    mastery_score NUMERIC(5,4) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT progress_oge_number_chk CHECK (oge_number BETWEEN 6 AND 19),
    CONSTRAINT progress_subtype_code_not_empty_chk CHECK (length(trim(subtype_code)) > 0),
    CONSTRAINT progress_attempts_count_non_negative_chk CHECK (attempts_count >= 0),
    CONSTRAINT progress_correct_count_non_negative_chk CHECK (correct_count >= 0),
    CONSTRAINT progress_correct_count_lte_attempts_chk CHECK (correct_count <= attempts_count),
    CONSTRAINT progress_mastery_score_chk CHECK (mastery_score >= 0 AND mastery_score <= 1)
);

CREATE UNIQUE INDEX IF NOT EXISTS progress_user_oge_subtype_uidx
    ON progress (user_id, oge_number, subtype_code);

CREATE INDEX IF NOT EXISTS progress_user_id_idx
    ON progress (user_id);

CREATE INDEX IF NOT EXISTS progress_task_type_id_idx
    ON progress (task_type_id);

CREATE INDEX IF NOT EXISTS progress_oge_subtype_idx
    ON progress (oge_number, subtype_code);

CREATE INDEX IF NOT EXISTS progress_mastery_score_idx
    ON progress (mastery_score);

CREATE TABLE IF NOT EXISTS ai_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    endpoint TEXT NOT NULL,
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    request_payload JSONB,
    response_payload JSONB,
    prompt_tokens INT,
    completion_tokens INT,
    total_tokens INT,
    cost NUMERIC(12,6),
    error_text TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ai_logs_prompt_tokens_non_negative_chk CHECK (prompt_tokens IS NULL OR prompt_tokens >= 0),
    CONSTRAINT ai_logs_completion_tokens_non_negative_chk CHECK (completion_tokens IS NULL OR completion_tokens >= 0),
    CONSTRAINT ai_logs_total_tokens_non_negative_chk CHECK (total_tokens IS NULL OR total_tokens >= 0),
    CONSTRAINT ai_logs_cost_non_negative_chk CHECK (cost IS NULL OR cost >= 0)
);

CREATE INDEX IF NOT EXISTS ai_logs_user_id_idx
    ON ai_logs (user_id);

CREATE INDEX IF NOT EXISTS ai_logs_endpoint_idx
    ON ai_logs (endpoint);

CREATE INDEX IF NOT EXISTS ai_logs_created_at_idx
    ON ai_logs (created_at);

COMMIT;
