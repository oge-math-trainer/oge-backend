BEGIN;

DELETE FROM prepared_tasks
WHERE used = true;

WITH ranked AS (
    SELECT
        id,
        row_number() OVER (
            PARTITION BY
                oge_number,
                subtype_code,
                md5(regexp_replace(lower(trim(question)), '\s+', ' ', 'g'))
            ORDER BY created_at ASC, id ASC
        ) AS duplicate_rank
    FROM prepared_tasks
    WHERE used = false
)
DELETE FROM prepared_tasks p
USING ranked r
WHERE p.id = r.id
  AND r.duplicate_rank > 1;

CREATE UNIQUE INDEX IF NOT EXISTS prepared_tasks_ready_question_uidx
    ON prepared_tasks (
        oge_number,
        subtype_code,
        md5(regexp_replace(lower(trim(question)), '\s+', ' ', 'g'))
    )
    WHERE used = false;

CREATE INDEX IF NOT EXISTS generated_tasks_user_question_hash_idx
    ON generated_tasks (
        user_id,
        mode,
        oge_number,
        subtype_code,
        md5(regexp_replace(lower(trim(question)), '\s+', ' ', 'g'))
    );

COMMIT;
