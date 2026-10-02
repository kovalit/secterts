-- 011 secret health, expiration, classification and usage tracking
--
-- Adds the metadata needed for the inventory, secret-health and expiration
-- features:
--   * entry_type       - classifies the record (password, api_key, ssh_key, ...)
--   * expires_at       - optional expiration date (API keys, certs, tokens, ...)
--   * owner            - free-text responsible owner; empty = "secret without owner"
--   * password_strength- 0..4 strength score computed on write (NULL = unknown)
--   * last_used_at      - last time the secret was revealed (NULL = never)
ALTER TABLE password_entries
    ADD COLUMN entry_type TEXT NOT NULL DEFAULT 'password',
    ADD COLUMN expires_at TIMESTAMPTZ,
    ADD COLUMN owner TEXT,
    ADD COLUMN password_strength SMALLINT,
    ADD COLUMN last_used_at TIMESTAMPTZ;

ALTER TABLE password_entries
    ADD CONSTRAINT chk_password_entries_entry_type
    CHECK (entry_type IN (
        'password', 'api_key', 'ssh_key', 'certificate',
        'token', 'license', 'database', 'secret_note', 'other'
    ));

CREATE INDEX idx_password_entries_entry_type
    ON password_entries(owner_user_id, entry_type)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_password_entries_expires_at
    ON password_entries(owner_user_id, expires_at)
    WHERE deleted_at IS NULL AND expires_at IS NOT NULL;
