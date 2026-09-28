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
├── docker-compose.yml
├── .env                      # локальная production-конфигурация, не в git
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
docker-compose.yml      # backend-контейнер
.env                    # конфигурация backend и сборки frontend
scripts/                # backup/restore/generate-master-key
```

### Быстрый старт через Docker Compose

```bash
# 1. Создать .env в корне и указать подключение к существующей PostgreSQL
./scripts/generate-master-key.sh   # вставьте ключ в .env

# 2. Поднять backend
make docker-up

# 3. Запустить frontend локально
cd apps/frontend && npm install && npm run dev
```

- Frontend: http://localhost:5173
- Backend API: http://localhost:9063 (health: `GET /health`)

Backend автоматически применяет миграции при старте.

### Локальная разработка

```bash
# Существующая PostgreSQL + переменные окружения из корневого .env

# Backend
make backend        # go run ./cmd/api   (миграции применяются автоматически)
make migrate        # применить миграции отдельно
make backend-test   # unit-тесты (crypto, argon2id, извлечение домена)

# Frontend
cd apps/frontend && npm install && VITE_API_URL=http://localhost:8080 npm run dev

# Chrome extension
cd apps/chrome-extension && npm install && npm run build
# затем chrome://extensions → «Загрузить распакованное расширение» → apps/chrome-extension/dist
# в options укажите API URL (http://localhost:8080) и токен из веб-приложения
```

### Деплой на сервер

Backend деплоится через Docker Registry и SSH, по аналогии с `ai-engine`.
Frontend собирается локально и копируется на сервер как статические файлы —
Docker и Node.js на production-сервере для него не нужны.

```bash
# заполнить корневой .env: DATABASE_URL, VITE_API_URL, ключи, домены и SMTP
make deploy
```

По умолчанию backend-образ публикуется в `ttdocker.me`, контейнер запускается на
`176.57.218.35` и доступен для reverse proxy на `127.0.0.1:9063`. Статика
frontend размещается в `/home/react/secrets`. Системный nginx раздаёт её на
`locker.ttspace.ru` и проксирует `locker.ttspace.ru/api` в backend. Значение
`VITE_API_URL=https://locker.ttspace.ru` встраивается во frontend во время
сборки; `/api` уже присутствует в путях клиента. Для отдельного деплоя только
frontend используйте `make deploy-frontend`.

### Поток авторизации (email 2FA)

Вход всегда двухшаговый: `email + пароль` → код на email → сессия в HttpOnly cookie.
Если `SMTP_HOST` не задан, код 2FA выводится в логи backend (удобно для разработки).
Для Timeweb используются `SMTP_HOST=smtp.timeweb.ru`, `SMTP_PORT=2525`,
`SMTP_USER=system@whatsbetter.me`, `SMTP_PASSWORD` и такой же `EMAIL_FROM`.

### Ключевые переменные окружения backend

```env
DATABASE_URL=postgres://USER:PASSWORD@DB_HOST:5432/secrets_center?sslmode=disable
APP_MASTER_KEY_V1_BASE64=<32 байта в base64>   # обязательно; потеря = невозможность расшифровки
CORS_ORIGINS=http://localhost:5173
COOKIE_SECURE=false                            # true только за HTTPS
```

Полный список описан в `docs/02_backend_go_project.md`.
