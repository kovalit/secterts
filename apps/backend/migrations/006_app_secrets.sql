-- 006 app projects, environments, secrets, key versions
CREATE TABLE app_projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(owner_user_id, company_id, name)
);

CREATE INDEX idx_app_projects_owner ON app_projects(owner_user_id);
CREATE INDEX idx_app_projects_company ON app_projects(company_id);

CREATE TABLE app_environments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES app_projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(project_id, name)
);

CREATE TABLE app_secrets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES app_projects(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL REFERENCES app_environments(id) ON DELETE CASCADE,

    key TEXT NOT NULL,
    encrypted_value BYTEA NOT NULL,
    value_nonce BYTEA NOT NULL,
    value_key_version INT NOT NULL DEFAULT 1,

    encrypted_comment BYTEA,
    comment_nonce BYTEA,
    comment_key_version INT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,

    UNIQUE(project_id, environment_id, key)
);

CREATE INDEX idx_app_secrets_project ON app_secrets(project_id);
CREATE INDEX idx_app_secrets_environment ON app_secrets(environment_id);

CREATE INDEX idx_app_secrets_project_env_key
ON app_secrets(project_id, environment_id, key)
WHERE deleted_at IS NULL;

CREATE TABLE app_key_versions (
    version INT PRIMARY KEY,
    algorithm TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    retired_at TIMESTAMPTZ,
    CONSTRAINT chk_app_key_versions_algorithm CHECK (algorithm IN ('AES-256-GCM')),
    CONSTRAINT chk_app_key_versions_status CHECK (status IN ('active', 'retired'))
);

INSERT INTO app_key_versions (version, algorithm, status)
VALUES (1, 'AES-256-GCM', 'active')
ON CONFLICT (version) DO NOTHING;
