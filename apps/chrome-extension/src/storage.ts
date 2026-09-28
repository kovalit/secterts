import type { Settings } from './types'

const DEFAULTS: Settings = {
  apiBaseUrl: 'https://locker.ttspace.ru',
  extensionToken: '',
}

// normalizeBase turns whatever the user typed into a clean origin (no trailing
// slash and no trailing "/api"). The backend mounts everything under "/api",
// and our request paths already include that prefix, so storing the bare origin
// avoids a doubled "/api/api/..." path.
export function normalizeBase(raw: string): string {
  let base = (raw || '').trim().replace(/\/+$/, '')
  base = base.replace(/\/api$/i, '')
  return base
}

// Read extension settings from chrome.storage.local.
export async function getSettings(): Promise<Settings> {
  const stored = await chrome.storage.local.get(['apiBaseUrl', 'extensionToken'])
  const base = (stored.apiBaseUrl as string) || DEFAULTS.apiBaseUrl
  return {
    // Normalize on read too, so any previously saved "/api" value keeps working.
    apiBaseUrl: normalizeBase(base) || normalizeBase(DEFAULTS.apiBaseUrl),
    extensionToken: (stored.extensionToken as string) || DEFAULTS.extensionToken,
  }
}

// Persist settings. The token is never logged.
export async function saveSettings(settings: Settings): Promise<void> {
  await chrome.storage.local.set({
    apiBaseUrl: normalizeBase(settings.apiBaseUrl),
    extensionToken: settings.extensionToken.trim(),
  })
}
