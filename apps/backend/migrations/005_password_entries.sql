-- 005 password entries
CREATE TABLE password_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES password_groups(id),

    scope TEXT NOT NULL,
    title TEXT NOT NULL,
    site_url TEXT,
    domain TEXT,
    favicon_url TEXT,
    icon_source TEXT NOT NULL DEFAULT 'group',

    login TEXT,
    encrypted_password BYTEA NOT NULL,
    password_nonce BYTEA NOT NULL,
    password_key_version INT NOT NULL DEFAULT 1,

    encrypted_comment BYTEA,
    comment_nonce BYTEA,
    comment_key_version INT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT chk_password_entries_scope CHECK (scope IN ('personal', 'commercial')),
    CONSTRAINT chk_password_entries_company CHECK (
        (scope = 'personal' AND company_id IS NULL)
        OR
        (scope = 'commercial' AND company_id IS NOT NULL)
    ),
    CONSTRAINT chk_password_entries_icon_source CHECK (icon_source IN ('favicon', 'group'))
);

CREATE INDEX idx_password_entries_owner ON password_entries(owner_user_id);
CREATE INDEX idx_password_entries_company ON password_entries(company_id);
CREATE INDEX idx_password_entries_domain ON password_entries(domain);
CREATE INDEX idx_password_entries_group ON password_entries(group_id);
CREATE INDEX idx_password_entries_scope ON password_entries(scope);

CREATE INDEX idx_password_entries_owner_domain
ON password_entries(owner_user_id, domain)
WHERE deleted_at IS NULL;
