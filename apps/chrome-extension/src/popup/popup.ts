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

function iconMarkup(item: LookupItem): string {
  if (item.favicon_url) {
    return `<span class="item-icon"><img src="${item.favicon_url}" alt="" onerror="this.style.display='none'" /></span>`
  }
  return `<span class="item-icon">🔑</span>`
}

function renderItems() {
  const multiple = items.length > 1
  const list = items
    .map((it) => {
      const selected = it.id === selectedId
      if (multiple) {
        return `
          <label class="item ${selected ? 'selected' : ''}">
            <div class="radio-row">
              <input type="radio" name="entry" value="${it.id}" ${selected ? 'checked' : ''} />
              ${iconMarkup(it)}
              <div style="min-width:0">
                <div class="item-title">${escapeHtml(it.title)}</div>
                <div class="item-login">${escapeHtml(it.login ?? '')}</div>
              </div>
            </div>
          </label>`
      }
      return `
        <div class="item selected">
          <div class="item-head">
            ${iconMarkup(it)}
            <div style="min-width:0">
              <div class="item-title">${escapeHtml(it.title)}</div>
              <div class="item-login">${escapeHtml(it.login ?? '')}</div>
            </div>
          </div>
        </div>`
    })
    .join('')

  content.innerHTML = `
    ${list}
    <div class="item-actions">
      <button class="btn" id="copyLogin">Копировать логин</button>
      <button class="btn btn-primary" id="copyPassword">Копировать пароль</button>
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

  document.getElementById('copyLogin')?.addEventListener('click', onCopyLogin)
  document.getElementById('copyPassword')?.addEventListener('click', onCopyPassword)
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
    const base = s.apiBaseUrl.replace(/:8080$/, ':5173')
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
