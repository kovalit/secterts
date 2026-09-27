import type { Settings } from './types'

const DEFAULTS: Settings = {
  apiBaseUrl: 'http://localhost:8080',
  extensionToken: '',
}

// Read extension settings from chrome.storage.local.
export async function getSettings(): Promise<Settings> {
  const stored = await chrome.storage.local.get(['apiBaseUrl', 'extensionToken'])
  return {
    apiBaseUrl: (stored.apiBaseUrl as string) || DEFAULTS.apiBaseUrl,
    extensionToken: (stored.extensionToken as string) || DEFAULTS.extensionToken,
  }
}

// Persist settings. The token is never logged.
export async function saveSettings(settings: Settings): Promise<void> {
  await chrome.storage.local.set({
    apiBaseUrl: settings.apiBaseUrl.trim().replace(/\/+$/, ''),
    extensionToken: settings.extensionToken.trim(),
  })
}
