# 03. Frontend проект — React

## Цель

Сделать простой web UI для управления:

```text
- личными паролями;
- коммерческими паролями по компаниям;
- секретами приложений;
- токеном для Chrome extension;
- настройками безопасности.
```

Стек MVP:

```text
React
TypeScript
Vite
React Router
TanStack Query
React Hook Form
Zod
Tailwind CSS
Lucide Icons
```

---

# 1. Структура frontend

```text
apps/frontend/
├── src/
│   ├── app/
│   │   ├── App.tsx
│   │   ├── router.tsx
│   │   └── queryClient.ts
│   │
│   ├── api/
│   │   ├── client.ts
│   │   ├── auth.ts
│   │   ├── passwords.ts
│   │   ├── companies.ts
│   │   ├── appSecrets.ts
│   │   └── extension.ts
│   │
│   ├── features/
│   │   ├── auth/
│   │   ├── passwords/
│   │   ├── companies/
│   │   ├── app-secrets/
│   │   ├── extension-tokens/
│   │   └── settings/
│   │
│   ├── components/
│   │   ├── layout/
│   │   ├── ui/
│   │   └── icons/
│   │
│   ├── lib/
│   │   ├── domain.ts
│   │   ├── clipboard.ts
│   │   └── formatDate.ts
│   │
│   └── main.tsx
│
├── index.html
├── package.json
├── vite.config.ts
└── Dockerfile
```

---

# 2. Основной layout

```text
┌─────────────────────────────────────────────────────┐
│ Secrets Center                                      │
├───────────────┬─────────────────────────────────────┤
│ Пароли        │                                     │
│  - Все        │      Основная область               │
│  - Личное     │                                     │
│  - Компании   │                                     │
│               │                                     │
│ Секреты       │                                     │
│  - Проекты    │                                     │
│               │                                     │
│ Настройки     │                                     │
└───────────────┴─────────────────────────────────────┘
```

UI должен быть максимально простой: таблица, фильтры, кнопка добавления, drawer/modal редактирования.

---

# 3. Страницы

## Auth

```text
/login
/register
/verify-email-code
```

### Login flow

```text
1. /login — ввод email + password.
2. Backend возвращает challenge_id.
3. Переход на /verify-email-code.
4. Пользователь вводит код из email.
5. После успеха переход на /passwords.
```

---

## Passwords

```text
/passwords
/passwords?scope=personal
/passwords?scope=commercial&company_id=...
```

### Компоненты

```text
PasswordListPage
PasswordFilters
PasswordTable
PasswordCardMobile
PasswordEditorDrawer
PasswordRevealButton
PasswordCopyButton
GroupBadge
CompanyBadge
FaviconIcon
```

### Фильтры

```text
- поиск;
- группа;
- личное / коммерческое;
- компания;
```

### Таблица

```text
Иконка | Название | Сайт | Логин | Группа | Тип | Компания | Действия
```

Действия:

```text
- копировать логин;
- показать/скопировать пароль;
- редактировать;
- удалить.
```

Пароль не показывать сразу. Только после нажатия `Показать` или `Скопировать`.

---

## Companies

```text
/companies
```

Минимально:

```text
- список компаний;
- добавить компанию;
- переименовать;
- удалить;
```

---

## App Secrets

```text
/app-secrets
/app-secrets/:projectId
```

### Экран проектов

```text
Проект | Компания | Количество окружений | Количество секретов | Дата обновления
```

### Экран проекта

```text
Tabs: dev / staging / prod
```

Таблица секретов:

```text
KEY | Есть значение | Комментарий | Обновлено | Действия
```

Действия:

```text
- reveal;
- copy value;
- edit;
- delete.
```

---

## Extension Tokens

```text
/settings/extension
```

Функции:

```text
- создать токен для Chrome extension;
- скопировать токен один раз;
- посмотреть список активных токенов;
- отозвать токен.
```

Важно: после создания токен показывается только один раз.

---

## Settings / Security

```text
/settings/security
```

Показываем:

```text
- email;
- email 2FA включена;
- список активных сессий;
- смена пароля;
```

---

# 4. Типы данных TypeScript

```ts
export type PasswordScope = 'personal' | 'commercial'

export type PasswordGroup = {
  id: string
  slug: string
  label: string
  icon: string
  sort_order: number
}

export type Company = {
  id: string
  name: string
}

export type PasswordEntry = {
  id: string
  scope: PasswordScope
  company_id?: string | null
  group_id: string
  title: string
  site_url?: string | null
  domain?: string | null
  favicon_url?: string | null
  icon_source: 'favicon' | 'group'
  login?: string | null
  has_password: boolean
  created_at: string
  updated_at: string
}

export type AppProject = {
  id: string
  company_id?: string | null
  name: string
  description?: string | null
  created_at: string
  updated_at: string
}

export type AppEnvironment = {
  id: string
  project_id: string
  name: string
  sort_order: number
}

export type AppSecret = {
  id: string
  project_id: string
  environment_id: string
  key: string
  has_value: boolean
  created_at: string
  updated_at: string
}
```

---

# 5. API client

`api/client.ts`:

```ts
export async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${import.meta.env.VITE_API_URL}${path}`, {
    ...options,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {}),
    },
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({}))
    throw new Error(error.message || 'Request failed')
  }

  return response.json()
}
```

---

# 6. UX правила

```text
- Пароль не отображать в таблице.
- Кнопка Copy password вызывает backend reveal.
- После копирования показывать toast: Скопировано.
- Reveal password скрывать обратно через 30 секунд.
- Для удаления показывать confirm modal.
- Для коммерческой записи обязательно выбрать компанию.
- Если favicon нет, показывать иконку группы.
```

---

# 7. Этапы реализации frontend с промптами

## Этап 1. Создать React app

### Промпт

```text
Создай React + TypeScript frontend в apps/frontend.
Стек: Vite, React Router, TanStack Query, Tailwind CSS, React Hook Form, Zod, Lucide Icons.
Добавь базовый layout с left sidebar и routes:
/login
/register
/verify-email-code
/passwords
/companies
/app-secrets
/settings/security
/settings/extension
Сделай простой чистый UI без лишней графики.
```

---

## Этап 2. Реализовать API client и auth

### Промпт

```text
Реализуй api client и auth flow.
Нужны:
- apiFetch with credentials include;
- login form email/password;
- verify email code page;
- register page;
- useMe query;
- protected routes;
- logout;
- обработка ошибок и toast notifications.
Login endpoint возвращает challenge_id, после чего нужно перейти на verify-email-code.
```

---

## Этап 3. Реализовать раздел паролей

### Промпт

```text
Реализуй раздел /passwords.
Нужны:
- загрузка password groups;
- загрузка companies;
- список password entries;
- фильтры: q, scope, company_id, group_id;
- таблица desktop;
- карточки mobile;
- drawer для создания/редактирования;
- поля: title, site_url, login, password, comment, group_id, scope, company_id;
- кнопки copy login, reveal/copy password, edit, delete;
- password reveal/copy вызывает POST /api/passwords/{id}/reveal.
```

---

## Этап 4. Реализовать компании

### Промпт

```text
Реализуй страницу /companies.
Нужны:
- список компаний;
- создать компанию;
- переименовать компанию;
- удалить компанию;
- после создания компания должна быть доступна в форме коммерческого пароля.
```

---

## Этап 5. Реализовать секреты приложений

### Промпт

```text
Реализуй раздел /app-secrets.
Нужны:
- список app projects;
- создать/редактировать/удалить project;
- внутри проекта вкладки environments: dev, staging, prod;
- таблица secrets по выбранному environment;
- создать/редактировать/удалить secret;
- reveal/copy secret value;
- value не показывать в списке;
- copy вызывает reveal endpoint.
```

---

## Этап 6. Реализовать настройки Chrome extension

### Промпт

```text
Реализуй страницу /settings/extension.
Нужны:
- кнопка Create extension token;
- показать созданный токен только один раз;
- кнопка Copy token;
- список активных токенов;
- revoke token;
- предупреждение, что token дает доступ к чтению паролей для Chrome extension.
```

---

## Этап 7. Улучшить безопасность UI

### Промпт

```text
Добавь frontend security UX:
- автоматическое скрытие revealed password через 30 секунд;
- очистка password value из состояния после закрытия drawer;
- confirm modal перед удалением;
- toast для всех copy actions;
- предупреждение при создании extension token;
- logout при 401;
- не писать секреты в console.log.
```

---

# 8. Минимальный дизайн

## Визуальный стиль

```text
- светлый интерфейс;
- много воздуха;
- акцентный цвет: синий;
- таблицы с мягкими границами;
- иконки Lucide;
- маленькие badges для групп и компаний;
```

## Цвета

```text
background: #F8FAFC
surface: #FFFFFF
border: #E2E8F0
text: #0F172A
muted: #64748B
primary: #2563EB
primary hover: #1D4ED8
warning: #F59E0B
error: #DC2626
```
