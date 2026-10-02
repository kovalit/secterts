-- 011 secret notes
--
-- Free-form secret notes: a title, a type (api_key / token / credentials / ...)
-- and an encrypted text body. The text is stored only in encrypted form, the
-- same way password and app-secret values are.
CREATE TABLE secret_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    title TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'generic',

    encrypted_text BYTEA NOT NULL,
    text_nonce BYTEA NOT NULL,
    text_key_version INT NOT NULL DEFAULT 1,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT chk_secret_notes_type CHECK (type IN (
        'api_key', 'token', 'credentials', 'server', 'database',
        'ssh_key', 'certificate', 'environment', 'generic'
    ))
);

CREATE INDEX idx_secret_notes_owner ON secret_notes(owner_user_id);

CREATE INDEX idx_secret_notes_owner_active
ON secret_notes(owner_user_id, updated_at DESC)
WHERE deleted_at IS NULL;
