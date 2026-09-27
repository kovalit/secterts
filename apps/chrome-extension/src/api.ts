import { getSettings } from './storage'
import type { LookupResponse } from './types'

// Error thrown when the extension is missing its API URL or token.
export class NotConfiguredError extends Error {
  constructor() {
    super('Extension is not configured')
    this.name = 'NotConfiguredError'
  }
}

// Error thrown on an invalid/expired token (HTTP 401).
export class UnauthorizedError extends Error {
  constructor() {
    super('Unauthorized')
    this.name = 'UnauthorizedError'
  }
}

async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
  const settings = await getSettings()
  if (!settings.apiBaseUrl || !settings.extensionToken) {
    throw new NotConfiguredError()
  }

  const res = await fetch(`${settings.apiBaseUrl}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${settings.extensionToken}`,
      ...(options.headers ?? {}),
    },
  })

  if (res.status === 401) throw new UnauthorizedError()
  if (!res.ok) throw new Error(`Request failed (${res.status})`)
  return (await res.json()) as T
}

// Look up entries for a domain.
export function lookup(domain: string): Promise<LookupResponse> {
  return apiFetch<LookupResponse>(`/api/extension/lookup?domain=${encodeURIComponent(domain)}`)
}

// Reveal a password (only after an explicit user click); backend checks domain.
export function revealPassword(entryId: string, domain: string): Promise<{ password: string }> {
  return apiFetch<{ password: string }>(`/api/extension/passwords/${entryId}/reveal`, {
    method: 'POST',
    body: JSON.stringify({ domain }),
  })
}

// Lightweight connectivity check used by the options page.
export function testConnection(): Promise<LookupResponse> {
  return lookup('example.com')
}
