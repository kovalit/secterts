import { getSettings, saveSettings } from '../storage'
import { testConnection, NotConfiguredError, UnauthorizedError } from '../api'

const apiInput = document.getElementById('apiBaseUrl') as HTMLInputElement
const tokenInput = document.getElementById('token') as HTMLInputElement
const saveBtn = document.getElementById('save') as HTMLButtonElement
const testBtn = document.getElementById('test') as HTMLButtonElement
const statusEl = document.getElementById('status') as HTMLElement

function setStatus(kind: 'ok' | 'err', message: string) {
  statusEl.hidden = false
  statusEl.className = `status ${kind}`
  statusEl.textContent = message
}

async function load() {
  const s = await getSettings()
  apiInput.value = s.apiBaseUrl
  tokenInput.value = s.extensionToken
}

saveBtn.addEventListener('click', async () => {
  await saveSettings({ apiBaseUrl: apiInput.value, extensionToken: tokenInput.value })
  setStatus('ok', 'Настройки сохранены')
})

testBtn.addEventListener('click', async () => {
  // Save first so the test uses the current values.
  await saveSettings({ apiBaseUrl: apiInput.value, extensionToken: tokenInput.value })
  setStatus('ok', 'Проверяем…')
  try {
    await testConnection()
    setStatus('ok', 'Соединение успешно — токен действителен')
  } catch (err) {
    if (err instanceof NotConfiguredError) {
      setStatus('err', 'Заполните API URL и токен')
    } else if (err instanceof UnauthorizedError) {
      setStatus('err', 'Токен недействителен или истёк')
    } else {
      setStatus('err', 'Не удалось соединиться с сервером')
    }
  }
})

void load()
