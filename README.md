# Secrets Center — минимальная версия

Документация описывает самый простой self-hosted сервис для хранения личных паролей и секретов приложений.

## Цель MVP

Сделать один monorepo с тремя приложениями:

```text
secrets-center/
├── apps/
│   ├── backend/              # Go API
│   ├── frontend/             # React web UI
│   └── chrome-extension/     # простой Chrome plugin
├── docs/
├── deploy/
│   ├── docker-compose.yml
│   └── nginx/
├── migrations/
└── README.md
```

## Основные разделы продукта

```text
Secrets Center
├── Личные пароли
│   ├── Личное
│   └── Коммерческое / по компаниям
│
├── Секреты приложений
│   ├── Проекты
│   ├── Окружения
│   └── Переменные / секреты
│
└── Настройки безопасности
    ├── Пароль
    ├── Email 2FA
    ├── Сессии
    └── Резервные копии
```

## Предустановленные группы паролей

```text
1. Почтовые ящики
2. Социальные сети
3. Хостинг и инфраструктура
4. Разработка
```

Эти группы создаются миграцией и доступны всем пользователям. Позже можно добавить пользовательские группы.

## Запись пароля

Каждая запись содержит:

```text
- иконка;
- адрес сайта;
- логин;
- пароль;
- комментарий;
- группа;
- тип: личное / коммерческое;
- компания, если запись коммерческая.
```

Иконка определяется так:

```text
1. Пробуем получить favicon по домену сайта.
2. Если favicon не найден, используем иконку группы.
```

## Запись секрета приложения

Минимальный набор полей:

```text
- проект;
- окружение: dev / staging / prod;
- ключ: DATABASE_URL, JWT_SECRET, S3_SECRET_KEY;
- значение секрета;
- комментарий;
- дата обновления.
```

## Авторизация

Обязательно:

```text
1. Email + пароль.
2. Второй шаг: код на email.
```

В MVP email-код — обязательный второй фактор. В будущем лучше добавить TOTP / WebAuthn, потому что email 2FA слабее отдельного второго фактора.

## Принцип хранения секретов

Пароли и секреты нельзя хранить в открытом виде.

Минимальная схема MVP:

```text
PostgreSQL хранит только зашифрованные значения.
Go backend шифрует/расшифровывает значения через AES-256-GCM.
Master key хранится вне базы, например в ENV на сервере.
```

Пример:

```text
APP_MASTER_KEY_BASE64=...
```

Важно: если потерять master key, расшифровать старые записи будет невозможно.

## Что НЕ делать в MVP

```text
- Не хранить пароли открытым текстом.
- Не делать автозаполнение на сайтах через content script.
- Не хранить master key в PostgreSQL.
- Не хранить emergency-доступы только внутри этого же сервиса.
- Не отдавать пароль в Chrome extension без явного действия пользователя.
```

## Безопасность MVP

Минимальные требования:

```text
- Argon2id для хэша пользовательского пароля.
- AES-256-GCM для шифрования секретов.
- HTTPS обязательно.
- HttpOnly Secure cookies для web-сессии.
- Отдельный токен для Chrome extension.
- Audit log для просмотра/копирования/изменения секретов.
- Rate limit на login и email-код.
```

## Файлы документации

```text
01_database_postgres.md
02_backend_go_project.md
03_frontend_react_project.md
04_chrome_extension_project.md
05_repo_structure_and_stages.md
```

---

## Реализация

Monorepo из трёх приложений:

```text
apps/backend/           # Go API (chi, pgx, argon2id, AES-256-GCM)
apps/frontend/          # React + TypeScript + Vite web UI
apps/chrome-extension/  # Chrome MV3 extension (TypeScript + Vite)
deploy/                 # docker-compose, nginx, env-примеры
scripts/                # backup/restore/generate-master-key
```

### Быстрый старт через Docker Compose

```bash
# 1. Сгенерировать master key и создать deploy/backend.env
cp deploy/backend.env.example deploy/backend.env
./scripts/generate-master-key.sh   # вставьте APP_MASTER_KEY_V1_BASE64 в deploy/backend.env

# 2. Поднять PostgreSQL + backend + frontend
make docker-up      # или: docker compose -f deploy/docker-compose.yml up -d
```

- Frontend: http://localhost:5173
- Backend API: http://localhost:8080 (health: `GET /health`)

Backend автоматически применяет миграции при старте.

### Локальная разработка

```bash
# PostgreSQL (например, через docker) + переменные окружения из deploy/backend.env

# Backend
make backend        # go run ./cmd/api   (миграции применяются автоматически)
make migrate        # применить миграции отдельно
make backend-test   # unit-тесты (crypto, argon2id, извлечение домена)

# Frontend
cd apps/frontend && cp ../../deploy/frontend.env.example .env.local && npm install && npm run dev

# Chrome extension
cd apps/chrome-extension && npm install && npm run build
# затем chrome://extensions → «Загрузить распакованное расширение» → apps/chrome-extension/dist
# в options укажите API URL (http://localhost:8080) и токен из веб-приложения
```

### Поток авторизации (email 2FA)

Вход всегда двухшаговый: `email + пароль` → код на email → сессия в HttpOnly cookie.
Если `SMTP_HOST` не задан, код 2FA выводится в логи backend (удобно для разработки).

### Ключевые переменные окружения backend

```env
DATABASE_URL=postgres://postgres:postgres@postgres:5432/secrets_center?sslmode=disable
APP_MASTER_KEY_V1_BASE64=<32 байта в base64>   # обязательно; потеря = невозможность расшифровки
CORS_ORIGINS=http://localhost:5173
COOKIE_SECURE=false                            # true только за HTTPS
```

Полный список — в `deploy/backend.env.example` и `docs/02_backend_go_project.md`.
