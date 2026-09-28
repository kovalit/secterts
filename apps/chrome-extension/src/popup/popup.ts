import { getHostnameFromUrl } from '../domain'
import { lookup, revealPassword, NotConfiguredError, UnauthorizedError } from '../api'
import { getSettings } from '../storage'
import type { LookupItem } from '../types'

const content = document.getElementById('content') as HTMLElement
const domainEl = document.getElementById('domain') as HTMLElement
const toastEl = document.getElementById('toast') as HTMLElement

let currentDomain = ''
let items: LookupItem[] = []
let selectedId = ''

function showToast(message: string) {
  toastEl.textContent = message
  toastEl.hidden = false
  window.setTimeout(() => {
    toastEl.hidden = true
  }, 2000)
}

function renderMessage(html: string) {
  content.innerHTML = `<div class="state">${html}</div>`
}

function renderError(html: string) {
  content.innerHTML = `<div class="error-box">${html}</div>`
}

function iconMarkup(item?: LookupItem): string {
  if (item?.favicon_url) {
    return `<span class="item-icon"><img src="${item.favicon_url}" alt="" onerror="this.style.display='none'" /></span>`
  }
  return `<span class="item-icon">🔑</span>`
}

function copyIcon(): string {
  return `<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>`
}

function fieldRow(label: string, valueHtml: string, act: 'login' | 'password', mono: boolean): string {
  return `
    <div class="field">
      <div class="field-body">
        <div class="field-label">${label}</div>
        <div class="field-value ${mono ? 'mono' : ''}">${valueHtml}</div>
      </div>
      <button class="copy-btn" data-act="${act}" title="Копировать ${label.toLowerCase()}">${copyIcon()}</button>
    </div>`
}

function renderItems() {
  const item = selected()
  const multiple = items.length > 1

  // Entry picker (only when several entries match the domain).
  const selector = multiple
    ? `<div class="selector">${items
        .map(
          (it) => `
            <label class="radio-row ${it.id === selectedId ? 'selected' : ''}">
              <input type="radio" name="entry" value="${it.id}" ${it.id === selectedId ? 'checked' : ''} />
              ${iconMarkup(it)}
              <span class="radio-title">${escapeHtml(it.title)}</span>
            </label>`,
        )
        .join('')}</div>`
    : ''

  const loginHtml = item?.login ? escapeHtml(item.login) : '<span class="empty">нет логина</span>'

  content.innerHTML = `
    ${selector}
    <div class="entry">
      <div class="entry-head">
        ${iconMarkup(item)}
        <div class="entry-title">${escapeHtml(item?.title ?? '')}</div>
      </div>
      ${fieldRow('Логин', loginHtml, 'login', false)}
      ${fieldRow('Пароль', '•••••••••••', 'password', true)}
    </div>
  `

  if (multiple) {
    content.querySelectorAll<HTMLInputElement>('input[name="entry"]').forEach((input) => {
      input.addEventListener('change', () => {
        selectedId = input.value
        renderItems()
      })
    })
  }

  content.querySelectorAll<HTMLButtonElement>('.copy-btn').forEach((btn) => {
    btn.addEventListener('click', () => {
      if (btn.getAttribute('data-act') === 'login') void onCopyLogin()
      else void onCopyPassword()
    })
  })
}

function selected(): LookupItem | undefined {
  return items.find((i) => i.id === selectedId)
}

async function onCopyLogin() {
  const item = selected()
  if (!item?.login) {
    showToast('У записи нет логина')
    return
  }
  await navigator.clipboard.writeText(item.login)
  showToast('Логин скопирован')
}

async function onCopyPassword() {
  const item = selected()
  if (!item) return
  try {
    // Password is fetched only on this explicit click and never stored.
    const { password } = await revealPassword(item.id, currentDomain)
    await navigator.clipboard.writeText(password)
    showToast('Пароль скопирован')
  } catch (err) {
    if (err instanceof UnauthorizedError) {
      renderNeedsAuth()
    } else {
      showToast('Не удалось получить пароль')
    }
  }
}

function renderNeedsAuth() {
  renderError(
    'Токен недействителен или истёк. Откройте <a id="toOptions">настройки</a> и вставьте новый токен из веб-приложения.',
  )
  document.getElementById('toOptions')?.addEventListener('click', () => chrome.runtime.openOptionsPage())
}

function escapeHtml(s: string): string {
  return s.replace(/[&<>"']/g, (c) =>
    ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[c] as string,
  )
}

async function init() {
  document.getElementById('openOptions')?.addEventListener('click', () => chrome.runtime.openOptionsPage())
  document.getElementById('openWeb')?.addEventListener('click', async () => {
    const s = await getSettings()
    // On localhost the web app runs on the Vite dev port; in production it is
    // served from the same origin as the API.
    const base = s.apiBaseUrl.includes('localhost')
      ? s.apiBaseUrl.replace(/:8080$/, ':5173')
      : s.apiBaseUrl
    chrome.tabs.create({ url: base || 'http://localhost:5173' })
  })

  const settings = await getSettings()
  if (!settings.extensionToken) {
    renderError(
      'Расширение не настроено. Откройте <a id="toOptions">настройки</a> и вставьте API URL и токен.',
    )
    document.getElementById('toOptions')?.addEventListener('click', () => chrome.runtime.openOptionsPage())
    return
  }

  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true })
  const domain = tab?.url ? getHostnameFromUrl(tab.url) : null
  if (!domain) {
    domainEl.textContent = '—'
    renderMessage('Не удалось определить домен вкладки')
    return
  }
  currentDomain = domain
  domainEl.textContent = domain

  try {
    const res = await lookup(domain)
    items = res.items
    if (items.length === 0) {
      renderMessage(`Для <b>${escapeHtml(domain)}</b> записей нет`)
      return
    }
    selectedId = items[0].id
    renderItems()
  } catch (err) {
    if (err instanceof NotConfiguredError) {
      renderError('Расширение не настроено. Откройте настройки.')
    } else if (err instanceof UnauthorizedError) {
      renderNeedsAuth()
    } else {
      renderError('Ошибка соединения с сервером. Проверьте API URL в настройках.')
    }
  }
}

void init()
