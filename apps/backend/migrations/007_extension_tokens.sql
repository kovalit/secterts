-- 007 chrome extension tokens
CREATE TABLE extension_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL DEFAULT 'Chrome Extension',
    scopes TEXT[] NOT NULL DEFAULT ARRAY['passwords:read'],
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at TIMESTAMPTZ
);

CREATE INDEX idx_extension_tokens_user ON extension_tokens(user_id);
CREATE INDEX idx_extension_tokens_expires ON extension_tokens(expires_at);
