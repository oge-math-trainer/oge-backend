BEGIN;

ALTER TABLE users
    ALTER COLUMN password_hash DROP NOT NULL;

CREATE TABLE IF NOT EXISTS user_identities (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    provider_user_id TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT user_identities_provider_chk CHECK (provider IN ('google', 'yandex')),
    CONSTRAINT user_identities_provider_user_id_not_empty_chk CHECK (length(trim(provider_user_id)) > 0),
    CONSTRAINT user_identities_email_not_empty_chk CHECK (length(trim(email)) > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS user_identities_provider_user_uidx
    ON user_identities (provider, provider_user_id);

CREATE INDEX IF NOT EXISTS user_identities_user_id_idx
    ON user_identities (user_id);

CREATE INDEX IF NOT EXISTS user_identities_email_lower_idx
    ON user_identities (lower(email));

COMMIT;
