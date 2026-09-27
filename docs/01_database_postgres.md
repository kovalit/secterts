# 01. Структура базы данных PostgreSQL

## Цель

Описать минимальную структуру PostgreSQL для сервиса хранения личных паролей и секретов приложений.

## Расширения PostgreSQL

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;
```

`pgcrypto` нужен для `gen_random_uuid()`.  
`citext` удобен для email без учета регистра.

---

# 1. Пользователи и авторизация

## users

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email CITEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    email_2fa_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### Комментарии

```text
password_hash — Argon2id hash.
email_2fa_enabled — для MVP всегда true.
```

---

## email_auth_codes

Коды для подтверждения входа по email.

```sql
CREATE TABLE email_auth_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash TEXT NOT NULL,
    purpose TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    attempts INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_email_auth_codes_user_id ON email_auth_codes(user_id);
CREATE INDEX idx_email_auth_codes_expires_at ON email_auth_codes(expires_at);
```

### purpose

```text
login_2fa
email_verify
password_reset
```

Код в базе хранится только как hash.

---

## sessions

```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash TEXT NOT NULL UNIQUE,
    user_agent TEXT,
    ip INET,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
```

---

# 2. Компании и коммерческий доступ

## companies

```sql
CREATE TABLE companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(owner_user_id, name)
);
```

Для MVP компания принадлежит одному владельцу. Позже можно добавить командный доступ.

---

## company_members

Таблица нужна, если сразу закладывать будущую командную работу.

```sql
CREATE TABLE company_members (
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY(company_id, user_id)
);
```

### role

```text
owner
admin
member
readonly
```

В MVP можно использовать только `owner`.

---

# 3. Группы личных паролей

## password_groups

```sql
CREATE TABLE password_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL,
    icon TEXT NOT NULL,
    is_system BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

## Предустановленные группы

```sql
INSERT INTO password_groups (slug, label, icon, sort_order) VALUES
('email', 'Почтовые ящики', 'mail', 10),
('social', 'Социальные сети', 'users', 20),
('hosting_infra', 'Хостинг и инфраструктура', 'server', 30),
('development', 'Разработка', 'code', 40)
ON CONFLICT (slug) DO NOTHING;
```

---

# 4. Личные и коммерческие пароли

## password_entries

```sql
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
```

### Почему комментарий тоже лучше шифровать

В комментарии часто пишут чувствительные данные:

```text
- backup email;
- recovery codes;
- ответы на секретные вопросы;
- где лежит SSH-ключ;
- особенности входа.
```

Поэтому `comment` лучше хранить зашифрованным.

---

# 5. Секреты приложений

## app_projects

```sql
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
```

---

## app_environments

```sql
CREATE TABLE app_environments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES app_projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(project_id, name)
);
```

Типовые окружения:

```text
dev
staging
prod
```

---

## app_secrets

```sql
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
```

---

# 6. Ключи шифрования

## app_key_versions

Таблица хранит только версии ключей, но не сами ключи.

```sql
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
```

Сам ключ хранится вне базы:

```text
APP_MASTER_KEY_V1_BASE64=...
```

---

# 7. Токены для Chrome extension

## extension_tokens

```sql
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
```

В Chrome extension хранится сам токен. В базе — только hash.

---

# 8. Audit log

## audit_logs

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    company_id UUID REFERENCES companies(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id UUID,
    ip INET,
    user_agent TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_company ON audit_logs(company_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at);
```

Примеры действий:

```text
login_success
login_failed
email_2fa_sent
email_2fa_success
password_entry_created
password_entry_updated
password_entry_revealed
password_entry_copied_by_extension
app_secret_created
app_secret_updated
app_secret_revealed
backup_export_created
```

---

# 9. Резервные копии

## backup_exports

```sql
CREATE TABLE backup_exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    status TEXT NOT NULL,
    file_name TEXT,
    storage_target TEXT,
    sha256 TEXT,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ,
    CONSTRAINT chk_backup_exports_status CHECK (status IN ('started', 'success', 'failed'))
);
```

---

# 10. Нормализация домена

Для поиска записи по текущему сайту Chrome extension backend должен нормализовать домен.

Пример:

```text
https://mail.google.com/mail/u/0/#inbox -> google.com или mail.google.com
https://github.com/settings/profile -> github.com
https://sub.example.co.uk -> example.co.uk или sub.example.co.uk
```

Для MVP можно хранить `domain` как hostname без протокола:

```text
github.com
mail.google.com
reg.ru
```

Позже можно добавить public suffix list и умный поиск по parent domain.

---

# 11. Минимальный порядок миграций

```text
001_extensions.sql
002_users_auth.sql
003_companies.sql
004_password_groups.sql
005_password_entries.sql
006_app_secrets.sql
007_extension_tokens.sql
008_audit_logs.sql
009_backup_exports.sql
```

---

# 12. Важные индексы для MVP

```sql
CREATE INDEX idx_password_entries_owner_domain
ON password_entries(owner_user_id, domain)
WHERE deleted_at IS NULL;

CREATE INDEX idx_app_secrets_project_env_key
ON app_secrets(project_id, environment_id, key)
WHERE deleted_at IS NULL;
```

---

# 13. Что можно добавить позже

```text
- пользовательские группы;
- шаринг записей между пользователями;
- client-side encryption;
- TOTP/WebAuthn;
- история версий секретов;
- ротация секретов;
- импорт из CSV / 1Password / Bitwarden;
- экспорт emergency backup;
- object lock для backup-хранилища.
```
