# 05. Monorepo структура и общий план реализации

## Цель

Все три части проекта должны жить в одном репозитории:

```text
1. Backend на Go
2. Frontend на React
3. Chrome extension
```

---

# 1. Monorepo структура

```text
secrets-center/
├── apps/
│   ├── backend/
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── migrations/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── sqlc.yaml
│   │
│   ├── frontend/
│   │   ├── src/
│   │   ├── public/
│   │   ├── Dockerfile
│   │   ├── package.json
│   │   └── vite.config.ts
│   │
│   └── chrome-extension/
│       ├── src/
│       ├── public/
│       ├── manifest.json
│       ├── package.json
│       └── vite.config.ts
│
├── docs/
│   ├── 01_database_postgres.md
│   ├── 02_backend_go_project.md
│   ├── 03_frontend_react_project.md
│   └── 04_chrome_extension_project.md
│
├── scripts/
│   ├── backup.sh
│   ├── restore.sh
│   └── generate-master-key.sh
│
├── .gitignore
├── .env
├── docker-compose.yml
├── README.md
└── Makefile
```

---

# 2. .gitignore

```gitignore
.env
*.env
*.env.local
node_modules/
dist/
build/
.DS_Store
coverage/
*.log

# secrets
*.key
*.pem
*.age
backup/*.tar.gz
backup/*.sql
```

---

# 3. Makefile

```makefile
.PHONY: dev backend frontend extension docker-up docker-down migrate

dev:
	docker compose --env-file .env -f docker-compose.yml up

backend:
	cd apps/backend && go run ./cmd/api

frontend:
	cd apps/frontend && npm run dev

extension:
	cd apps/chrome-extension && npm run dev

docker-up:
	docker compose --env-file .env -f docker-compose.yml up -d

docker-down:
	docker compose --env-file .env -f docker-compose.yml down

migrate:
	cd apps/backend && go run ./cmd/migrate
```

---

# 4. docker-compose.yml

```yaml
services:
  backend:
    image: ${BACKEND_IMAGE:?BACKEND_IMAGE is required}
    restart: unless-stopped
    env_file:
      - .env
    ports:
      - "127.0.0.1:${BACKEND_PORT:-9063}:8080"
```

---

# 5. Общий план реализации

## Фаза 1. Базовый каркас

```text
- создать monorepo;
- подключить существующую PostgreSQL;
- создать Go API;
- создать React UI;
- создать Chrome extension skeleton;
- настроить Docker Compose.
```

### Промпт

```text
Создай monorepo secrets-center с тремя приложениями:
apps/backend на Go,
apps/frontend на React + TypeScript + Vite,
apps/chrome-extension на Manifest V3 + TypeScript.
Добавь корневой docker-compose.yml только с backend; PostgreSQL уже существует.
Frontend запускается через Vite локально и деплоится как статический `dist`.
Добавь Makefile и README.
Проект должен запускаться локально.
```

---

## Фаза 2. База и авторизация

```text
- миграции PostgreSQL;
- users/sessions/email_auth_codes;
- register/login/email 2FA;
- protected routes на frontend.
```

### Промпт

```text
Реализуй базу данных и авторизацию для Secrets Center.
Используй PostgreSQL schema из docs/01_database_postgres.md.
Backend: Go, Chi, pgx, sqlc.
Frontend: login/register/verify email code.
Обязательный вход: password + email code.
```

---

## Фаза 3. Шифрование

```text
- AES-256-GCM;
- master key из ENV;
- encrypted fields для password и app secrets;
- unit tests.
```

### Промпт

```text
Добавь слой шифрования для секретов.
Все passwords, comments и app secret values должны храниться в PostgreSQL только зашифрованными.
Используй AES-256-GCM и APP_MASTER_KEY_V1_BASE64.
Добавь key version и unit tests.
```

---

## Фаза 4. Личные пароли

```text
- группы;
- личные/коммерческие записи;
- компании;
- favicon;
- reveal/copy.
```

### Промпт

```text
Реализуй раздел личных паролей.
Нужны системные группы: Почтовые ящики, Социальные сети, Хостинг и инфраструктура, Разработка.
Каждая запись содержит: иконка, сайт, логин, пароль, комментарий.
Записи делятся на personal и commercial. Для commercial обязательна компания.
Добавь frontend UI для списка, фильтров, создания, редактирования, удаления и копирования.
```

---

## Фаза 5. Секреты приложений

```text
- проекты;
- окружения;
- секреты;
- reveal/copy;
- UI.
```

### Промпт

```text
Реализуй раздел Application Secrets.
Нужны проекты, окружения dev/staging/prod и secrets key/value.
Значение хранить зашифрованным.
В списке значение не показывать.
Добавить reveal/copy только по кнопке.
Сделать React UI с проектами, вкладками окружений и таблицей секретов.
```

---

## Фаза 6. Chrome extension

```text
- токен extension;
- определение домена;
- lookup;
- copy login/password.
```

### Промпт

```text
Реализуй Chrome extension для Secrets Center.
При клике на иконку определить домен активной вкладки.
Вызвать backend lookup по домену.
Если запись есть — показать логин и две кнопки: Copy login, Copy password.
Copy password должен получать пароль только по клику через reveal endpoint.
Extension token хранить в chrome.storage.local.
```

---

## Фаза 7. Backup

```text
- export encrypted data;
- backup.sh;
- age encryption;
- upload через rclone;
- restore.md.
```

### Промпт

```text
Добавь простой backup mechanism.
Нужен scripts/backup.sh:
- создать export базы или pg_dump;
- добавить restore.md;
- упаковать tar.gz;
- зашифровать через age public key;
- загрузить через rclone во внешний storage;
- удалить временные файлы.
Добавь scripts/restore.sh и документацию восстановления.
```

---

## Фаза 8. Минимальный production hardening

```text
- HTTPS через Nginx;
- secure cookies;
- rate limit;
- audit log;
- CORS whitelist;
- revoke sessions;
- revoke extension tokens.
```

### Промпт

```text
Подготовь MVP к production.
Добавь:
- secure HttpOnly cookies;
- CORS whitelist;
- rate limit на auth endpoints;
- audit log для reveal/copy/update/delete;
- revoke sessions;
- revoke extension tokens;
- nginx config для frontend и backend;
- checklist перед деплоем.
```

---

# 6. Минимальный порядок запуска проекта

```text
1. Создать repo и docker-compose.
2. Реализовать backend healthcheck.
3. Подключить PostgreSQL.
4. Добавить миграции.
5. Добавить auth + email 2FA.
6. Добавить crypto layer.
7. Добавить пароли.
8. Добавить секреты приложений.
9. Добавить frontend UI.
10. Добавить Chrome extension.
11. Добавить backup.
12. Закрыть HTTPS и security настройки.
```

---

# 7. Что лучше оставить на потом

```text
- командный доступ с ролями;
- client-side encryption;
- TOTP/WebAuthn;
- автозаполнение форм;
- мобильное приложение;
- импорт из 1Password/Bitwarden;
- история версий секретов;
- автоматическая ротация секретов;
- HA-кластер;
- сложная система прав на уровне отдельных записей.
```

---

# 8. Главный принцип MVP

```text
Простой UI и быстрый запуск важнее сложной архитектуры,
но секреты нельзя хранить открытым текстом даже в MVP.
```
