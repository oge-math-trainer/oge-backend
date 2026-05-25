BEGIN;

CREATE TABLE IF NOT EXISTS password_reset_codes (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT password_reset_codes_hash_not_empty_chk CHECK (length(trim(code_hash)) > 0)
);

CREATE INDEX IF NOT EXISTS password_reset_codes_user_active_idx
    ON password_reset_codes (user_id, created_at DESC)
    WHERE used_at IS NULL;

CREATE INDEX IF NOT EXISTS password_reset_codes_expires_at_idx
    ON password_reset_codes (expires_at);

COMMIT;
