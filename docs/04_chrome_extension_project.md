# 04. Chrome Extension проект

## Цель

Сделать очень простой Google Chrome plugin:

```text
1. Пользователь нажимает на иконку расширения.
2. Расширение определяет текущий домен активной вкладки.
3. Запрашивает backend: есть ли пароли для этого домена.
4. Если есть — показывает мини-форму.
5. В форме две кнопки:
   - скопировать логин;
   - скопировать пароль.
```

Расширение не делает autofill и не внедряет scripts на страницу в MVP.

---

# 1. Техническая основа

Использовать Chrome Extension Manifest V3.

Manifest V3 использует `manifest_version: 3`, а действие по клику на иконку задается через `action`. Для фоновой логики в MV3 используется service worker.

---

# 2. Структура

```text
apps/chrome-extension/
├── public/
│   ├── icon16.png
│   ├── icon48.png
│   └── icon128.png
├── src/
│   ├── popup/
│   │   ├── popup.html
│   │   ├── popup.ts
│   │   └── popup.css
│   ├── background/
│   │   └── service-worker.ts
│   ├── options/
│   │   ├── options.html
│   │   ├── options.ts
│   │   └── options.css
│   ├── api.ts
│   ├── domain.ts
│   └── storage.ts
├── manifest.json
├── package.json
├── tsconfig.json
└── vite.config.ts
```

---

# 3. manifest.json

```json
{
  "manifest_version": 3,
  "name": "Secrets Center",
  "version": "0.1.0",
  "description": "Copy login and password from Secrets Center by current domain.",
  "icons": {
    "16": "icon16.png",
    "48": "icon48.png",
    "128": "icon128.png"
  },
  "action": {
    "default_title": "Secrets Center",
    "default_popup": "src/popup/popup.html"
  },
  "background": {
    "service_worker": "src/background/service-worker.js"
  },
  "permissions": [
    "activeTab",
    "storage",
    "clipboardWrite"
  ],
  "host_permissions": [
    "https://your-secrets-api.example.com/*"
  ],
  "options_page": "src/options/options.html"
}
```

Для локальной разработки:

```json
"host_permissions": [
  "http://localhost:8080/*",
  "https://your-secrets-api.example.com/*"
]
```

---

# 4. Хранение настроек extension

В `chrome.storage.local` хранить:

```json
{
  "apiBaseUrl": "https://your-secrets-api.example.com",
  "extensionToken": "token"
}
```

Токен создается в web UI:

```text
Settings → Chrome Extension → Create token
```

Пользователь копирует токен и вставляет его в options page расширения.

---

# 5. Popup flow

```text
Открыли popup
↓
Получили активную вкладку через chrome.tabs.query
↓
Достали hostname из tab.url
↓
Вызвали GET /api/extension/lookup?domain=hostname
↓
Если записей нет — показать "Для этого сайта записей нет"
↓
Если запись одна — показать ее
↓
Если записей несколько — показать select/list
↓
Кнопка "Логин" копирует login
↓
Кнопка "Пароль" вызывает reveal endpoint и копирует password
```

---

# 6. API extension

## Lookup

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

## Reveal password

```http
POST /api/extension/passwords/{id}/reveal
Authorization: Bearer extension_token
Content-Type: application/json
```

Тело:

```json
{
  "domain": "github.com"
}
```

Ответ:

```json
{
  "password": "plain password"
}
```

---

# 7. Минимальный popup UI

```text
┌────────────────────────────┐
│ Secrets Center             │
│ github.com                 │
├────────────────────────────┤
│ GitHub                     │
│ dev@example.com            │
│                            │
│ [Copy login] [Copy pass]   │
└────────────────────────────┘
```

Если записей несколько:

```text
┌────────────────────────────┐
│ github.com                 │
├────────────────────────────┤
│ ○ GitHub personal          │
│ ○ GitHub company           │
│                            │
│ [Copy login] [Copy pass]   │
└────────────────────────────┘
```

---

# 8. Security правила

```text
- Не хранить пароли в chrome.storage.local.
- Не делать autofill в MVP.
- Не внедрять content scripts.
- Не запрашивать host permissions для всех сайтов.
- Не логировать password в console.
- Password получать только после клика пользователя.
- После копирования не сохранять password в состоянии дольше нескольких секунд.
- Extension token можно отозвать из web UI.
```

---

# 9. domain.ts

```ts
export function getHostnameFromUrl(url: string): string | null {
  try {
    const parsed = new URL(url)
    if (!['http:', 'https:'].includes(parsed.protocol)) return null
    return parsed.hostname.replace(/^www\./, '')
  } catch {
    return null
  }
}
```

---

# 10. api.ts

```ts
export async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
  const settings = await getSettings()

  if (!settings.apiBaseUrl || !settings.extensionToken) {
    throw new Error('Extension is not configured')
  }

  const response = await fetch(`${settings.apiBaseUrl}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${settings.extensionToken}`,
      ...(options.headers || {})
    }
  })

  if (!response.ok) {
    throw new Error('Request failed')
  }

  return response.json()
}
```

---

# 11. popup.ts логика

```ts
async function initPopup() {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true })
  const domain = tab?.url ? getHostnameFromUrl(tab.url) : null

  if (!domain) {
    renderMessage('Не удалось определить домен')
    return
  }

  const result = await apiFetch<LookupResponse>(`/api/extension/lookup?domain=${encodeURIComponent(domain)}`)

  if (result.items.length === 0) {
    renderMessage('Для этого сайта записей нет')
    return
  }

  renderItems(result.items, domain)
}
```

---

# 12. Copy login

```ts
async function copyLogin(login: string) {
  await navigator.clipboard.writeText(login)
  showToast('Логин скопирован')
}
```

---

# 13. Copy password

```ts
async function copyPassword(entryId: string, domain: string) {
  const result = await apiFetch<{ password: string }>(`/api/extension/passwords/${entryId}/reveal`, {
    method: 'POST',
    body: JSON.stringify({ domain })
  })

  await navigator.clipboard.writeText(result.password)
  showToast('Пароль скопирован')
}
```

---

# 14. Options page

Поля:

```text
API URL
Extension token
[Save]
[Test connection]
```

Минимальная проверка:

```http
GET /api/extension/lookup?domain=example.com
```

Если токен валидный, backend вернет пустой список или записи.

---

# 15. Этапы реализации Chrome extension с промптами

## Этап 1. Создать каркас extension

### Промпт

```text
Создай Chrome Extension Manifest V3 в apps/chrome-extension.
Нужны:
- manifest.json;
- popup.html/popup.ts/popup.css;
- options.html/options.ts/options.css;
- service-worker.ts;
- TypeScript build через Vite;
- permissions: activeTab, storage, clipboardWrite;
- host_permissions для http://localhost:8080/*.
Popup должен открываться по клику на иконку расширения.
```

---

## Этап 2. Настройки extension

### Промпт

```text
Реализуй options page для Chrome extension.
Поля:
- API URL;
- Extension token.
Сохранять в chrome.storage.local.
Добавить кнопку Test connection.
Не логировать token в console.
```

---

## Этап 3. Определение домена

### Промпт

```text
Реализуй popup logic:
- получить активную вкладку через chrome.tabs.query;
- определить hostname из tab.url;
- убрать www.;
- если URL не http/https, показать ошибку;
- вывести domain в popup.
```

---

## Этап 4. Lookup записей

### Промпт

```text
Добавь api client для extension.
При открытии popup вызвать GET /api/extension/lookup?domain={domain}.
Если записей нет, показать сообщение.
Если есть одна запись, показать title, login, favicon.
Если несколько, показать список выбора.
```

---

## Этап 5. Copy login/password

### Промпт

```text
Добавь кнопки Copy login и Copy password.
Copy login копирует login из lookup response.
Copy password вызывает POST /api/extension/passwords/{id}/reveal с body {domain}, затем копирует password через navigator.clipboard.writeText.
После копирования показать toast.
Не хранить password в chrome.storage.local и не выводить в console.
```

---

## Этап 6. Безопасность и полировка

### Промпт

```text
Улучши безопасность и UX Chrome extension:
- не сохранять password в state после копирования;
- обработать 401: показать ссылку на options page;
- обработать отсутствие API URL/token;
- loading state;
- error state;
- маленький аккуратный UI 320px шириной;
- кнопка Open web app.
```
