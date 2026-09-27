# 02. Backend проект — Go, PostgreSQL, Docker

## Цель

Сделать backend API для Secrets Center.

Стек:

```text
Go
Chi Router
PostgreSQL
pgx
sqlc
Argon2id
AES-256-GCM
Docker Compose
```

---

# 1. Структура backend

```text
apps/backend/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── auth/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── password.go
│   │   ├── email_2fa.go
│   │   └── sessions.go
│   │
│   ├── crypto/
│   │   ├── encryptor.go
│   │   ├── aes_gcm.go
│   │   └── keyring.go
│   │
│   ├── passwords/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── favicon.go
│   │   └── domain.go
│   │
│   ├── appsecrets/
│   │   ├── handler.go
│   │   └── service.go
│   │
│   ├── extension/
│   │   ├── handler.go
│   │   └── service.go
│   │
│   ├── companies/
│   │   ├── handler.go
│   │   └── service.go
│   │
│   ├── audit/
│   │   └── service.go
│   │
│   ├── backup/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── export.go
│   │
│   ├── db/
│   │   ├── queries/
│   │   └── sqlc/
│   │
│   ├── httpx/
│   │   ├── middleware.go
│   │   ├── response.go
│   │   └── errors.go
│   │
│   └── config/
│       └── config.go
│
├── migrations/
├── sqlc.yaml
├── Dockerfile
├── go.mod
└── go.sum
```

---

# 2. Основные переменные окружения

```env
APP_ENV=dev
APP_PORT=8080
DATABASE_URL=postgres://postgres:postgres@postgres:5432/secrets_center?sslmode=disable
APP_MASTER_KEY_V1_BASE64=base64_32_bytes_key
JWT_ACCESS_SECRET=...
COOKIE_DOMAIN=localhost
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=...
SMTP_PASSWORD=...
EMAIL_FROM=no-reply@example.com
CORS_ORIGINS=http://localhost:5173
```

---

# 3. Авторизация

## Поток входа

```text
1. Пользователь вводит email + password.
2. Backend проверяет пароль.
3. Backend создает email_auth_codes с purpose=login_2fa.
4. Backend отправляет код на email.
5. Пользователь вводит код.
6. Backend создает session.
7. Web получает HttpOnly Secure cookie.
```

## API

```http
POST /api/auth/register
POST /api/auth/login
POST /api/auth/verify-email-code
POST /api/auth/logout
GET  /api/auth/me
POST /api/auth/password-reset/request
POST /api/auth/password-reset/confirm
```

## DTO

```json
{
  "email": "user@example.com",
  "password": "password"
}
```

```json
{
  "challenge_id": "uuid",
  "code": "123456"
}
```

---

# 4. Шифрование

## Интерфейс

```go
type Encryptor interface {
    EncryptString(plain string) (ciphertext []byte, nonce []byte, keyVersion int, err error)
    DecryptString(ciphertext []byte, nonce []byte, keyVersion int) (string, error)
}
```

## Алгоритм MVP

```text
AES-256-GCM
Key length: 32 bytes
Nonce: 12 bytes random
Master key: ENV APP_MASTER_KEY_V1_BASE64
```

## Важные правила

```text
- Не логировать открытые значения паролей и секретов.
- Не возвращать пароль в списках.
- Расшифровывать пароль только в reveal/copy endpoint.
- Все reveal/copy действия писать в audit log.
```

---

# 5. Личные пароли

## API

```http
GET    /api/password-groups
GET    /api/passwords
POST   /api/passwords
GET    /api/passwords/{id}
PUT    /api/passwords/{id}
DELETE /api/passwords/{id}
POST   /api/passwords/{id}/reveal
```

## Фильтры списка

```text
scope=personal|commercial
company_id=uuid
group_id=uuid
q=text
domain=github.com
```

## Создание записи

```json
{
  "scope": "commercial",
  "company_id": "uuid",
  "group_id": "uuid",
  "title": "GitHub",
  "site_url": "https://github.com",
  "login": "dev@example.com",
  "password": "plain password",
  "comment": "2FA у владельца"
}
```

## Ответ списка

```json
{
  "id": "uuid",
  "scope": "commercial",
  "company_id": "uuid",
  "group_id": "uuid",
  "title": "GitHub",
  "site_url": "https://github.com",
  "domain": "github.com",
  "favicon_url": "https://github.com/favicon.ico",
  "icon_source": "favicon",
  "login": "dev@example.com",
  "has_password": true,
  "created_at": "...",
  "updated_at": "..."
}
```

Пароль в списке не возвращается.

## Reveal response

```json
{
  "password": "plain password"
}
```

---

# 6. Favicon

Минимальный алгоритм:

```text
1. Из site_url достать hostname.
2. Попробовать https://{hostname}/favicon.ico.
3. Если ответ 200 и content-type image/*, сохранить favicon_url.
4. Если нет — icon_source=group.
```

Для MVP можно не скачивать и не хранить файл, а хранить только URL.

Позже можно добавить собственный cache favicon.

---

# 7. Компании

## API

```http
GET    /api/companies
POST   /api/companies
GET    /api/companies/{id}
PUT    /api/companies/{id}
DELETE /api/companies/{id}
```

## DTO

```json
{
  "name": "My Company"
}
```

---

# 8. Секреты приложений

## API проектов

```http
GET    /api/app-projects
POST   /api/app-projects
GET    /api/app-projects/{id}
PUT    /api/app-projects/{id}
DELETE /api/app-projects/{id}
```

## API окружений

```http
GET    /api/app-projects/{project_id}/environments
POST   /api/app-projects/{project_id}/environments
PUT    /api/app-environments/{id}
DELETE /api/app-environments/{id}
```

## API секретов

```http
GET    /api/app-projects/{project_id}/secrets?env=prod
POST   /api/app-projects/{project_id}/secrets
PUT    /api/app-secrets/{id}
DELETE /api/app-secrets/{id}
POST   /api/app-secrets/{id}/reveal
```

## Создание секрета

```json
{
  "environment_id": "uuid",
  "key": "DATABASE_URL",
  "value": "postgres://...",
  "comment": "prod db"
}
```

---

# 9. API для Chrome extension

## Подключение extension

```http
POST /api/extension/tokens
DELETE /api/extension/tokens/{id}
GET /api/extension/tokens
```

Токен создается только в web UI после полной авторизации с email 2FA.

## Lookup по домену

```http
GET /api/extension/lookup?domain=github.com
Authorization: Bearer extension_token
```

Ответ:

```json
{
  "domain": "github.com",
  "items": [
    {
      "id": "uuid",
      "title": "GitHub",
      "login": "dev@example.com",
      "favicon_url": "https://github.com/favicon.ico"
    }
  ]
}
```

## Получить пароль для копирования

```http
POST /api/extension/passwords/{id}/reveal
Authorization: Bearer extension_token
```

Тело:

```json
{
  "domain": "github.com"
}
```

Backend должен проверить, что `domain` совпадает с доменом записи.

---

# 10. Backup export

## API

```http
POST /api/backups/export
GET  /api/backups
```

## MVP подход

```text
1. Создать логический export JSON.
2. Архивировать.
3. Зашифровать age/public key или master backup key.
4. Сохранить локально или отправить во внешний storage.
```

В MVP можно начать с CLI/cron-скрипта без UI.

---

# 11. Docker Compose

```yaml
services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: secrets_center
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  backend:
    build:
      context: ../apps/backend
    env_file:
      - ../deploy/backend.env
    ports:
      - "8080:8080"
    depends_on:
      - postgres

volumes:
  postgres_data:
```

---

# 12. Этапы реализации backend с промптами

## Этап 1. Создать каркас Go API

### Промпт

```text
Создай backend проект на Go для monorepo apps/backend.
Стек: Go, chi router, pgx, sqlc, PostgreSQL, Dockerfile.
Нужны:
- cmd/api/main.go;
- internal/config;
- internal/httpx;
- health endpoint GET /health;
- docker build;
- подключение к PostgreSQL через DATABASE_URL;
- graceful shutdown.
Код должен быть простым и production-friendly.
```

---

## Этап 2. Добавить миграции и sqlc

### Промпт

```text
Добавь в Go backend миграции PostgreSQL и sqlc.
Используй структуру базы из docs/01_database_postgres.md.
Создай SQL migrations:
001_extensions.sql
002_users_auth.sql
003_companies.sql
004_password_groups.sql
005_password_entries.sql
006_app_secrets.sql
007_extension_tokens.sql
008_audit_logs.sql
009_backup_exports.sql
Добавь sqlc.yaml и базовые queries для users, sessions, password_entries, app_projects, app_secrets.
```

---

## Этап 3. Реализовать auth + email 2FA

### Промпт

```text
Реализуй модуль auth.
Требования:
- регистрация пользователя по email/password;
- Argon2id password hashing;
- login step 1: проверка email/password и отправка email-кода;
- login step 2: verify email code и создание session;
- refresh_token хранить hash в sessions;
- access session через HttpOnly Secure cookie;
- rate limit на login и verify code;
- коды 2FA хранить только в виде hash;
- audit log для login_success/login_failed/email_2fa_sent/email_2fa_success.
```

---

## Этап 4. Реализовать crypto layer

### Промпт

```text
Реализуй internal/crypto для шифрования значений.
Требования:
- AES-256-GCM;
- master key берется из APP_MASTER_KEY_V1_BASE64;
- nonce random 12 bytes;
- интерфейс EncryptString/DecryptString;
- keyVersion=1;
- не логировать plain text;
- добавить unit tests на encrypt/decrypt и ошибку при неверном ключе.
```

---

## Этап 5. Реализовать личные пароли

### Промпт

```text
Реализуй модуль passwords.
Нужны endpoints:
GET /api/password-groups
GET /api/passwords
POST /api/passwords
GET /api/passwords/{id}
PUT /api/passwords/{id}
DELETE /api/passwords/{id}
POST /api/passwords/{id}/reveal
Требования:
- password и comment хранить зашифрованными;
- в списке пароль не отдавать;
- reveal возвращает password и пишет audit log;
- поддержать scope personal/commercial;
- если scope commercial, company_id обязателен;
- извлекать domain из site_url;
- пробовать favicon по https://domain/favicon.ico;
- если favicon не найден, icon_source=group.
```

---

## Этап 6. Реализовать компании

### Промпт

```text
Реализуй модуль companies.
Нужны CRUD endpoints для компаний владельца:
GET /api/companies
POST /api/companies
GET /api/companies/{id}
PUT /api/companies/{id}
DELETE /api/companies/{id}
В MVP компания принадлежит текущему пользователю.
Проверяй, что пользователь может работать только со своими компаниями.
```

---

## Этап 7. Реализовать секреты приложений

### Промпт

```text
Реализуй модуль appsecrets.
Нужны проекты, окружения и секреты.
Endpoints:
GET/POST/PUT/DELETE /api/app-projects
GET/POST /api/app-projects/{project_id}/environments
GET/POST /api/app-projects/{project_id}/secrets?env=prod
PUT/DELETE /api/app-secrets/{id}
POST /api/app-secrets/{id}/reveal
Значение секрета и комментарий хранить зашифрованными.
В списке показывать key, даты и признак has_value, но не value.
Reveal должен писать audit log.
```

---

## Этап 8. Реализовать API для Chrome extension

### Промпт

```text
Реализуй extension API.
Нужны:
POST /api/extension/tokens
GET /api/extension/tokens
DELETE /api/extension/tokens/{id}
GET /api/extension/lookup?domain=github.com
POST /api/extension/passwords/{id}/reveal
Токены хранить в extension_tokens только hash.
Token scopes: passwords:read.
Lookup должен возвращать только записи текущего пользователя по domain.
Reveal должен проверять, что domain из запроса совпадает с domain записи.
Каждое раскрытие пароля записывать в audit log.
```

---

## Этап 9. Добавить backup export

### Промпт

```text
Добавь простой backup/export механизм.
Сделай CLI command или endpoint для создания JSON export:
- users metadata без password_hash;
- companies;
- password groups;
- encrypted password_entries;
- app projects/environments/secrets;
- key versions;
- audit не обязательно.
Export должен сохранять encrypted values как base64.
Добавь shell script, который архивирует export и шифрует через age public key.
```

---

## Этап 10. Тесты и безопасность

### Промпт

```text
Добавь backend tests.
Покрыть:
- Argon2id password verify;
- AES-GCM encrypt/decrypt;
- login flow with email 2FA mocked email sender;
- password create/list/reveal;
- app secret create/list/reveal;
- extension lookup and reveal;
- запрет доступа к чужим company/password/app_secret.
Добавь security middleware: CORS whitelist, request id, recovery, rate limit for auth.
```
