import { apiFetch } from './client'
import type { EntryType, PasswordEntry, PasswordGroup } from '../types'

export type PasswordFilters = {
  q?: string
  scope?: string
  company_id?: string
  group_id?: string
  entry_type?: string
  domain?: string
}

export type PasswordWrite = {
  scope: string
  company_id?: string | null
  group_id: string
  title: string
  site_url?: string | null
  login?: string | null
  password?: string
  comment?: string | null
  icon_source?: 'favicon' | 'group' | 'custom'
  custom_icon?: string | null
  entry_type?: EntryType
  expires_at?: string | null
  owner?: string | null
}

export type InventoryItem = { type: EntryType; count: number }

export type Inventory = {
  total: number
  by_type: InventoryItem[]
  expiring_this_month: number
  expiring_soon: number
  expired: number
  weak_passwords: number
  without_owner: number
  unused: number
}

export type HealthSummary = {
  total: number
  weak: number
  no_owner: number
  unused: number
  expiring: number
  expired: number
}

export type HealthReport = {
  summary: HealthSummary
  weak: PasswordEntry[]
  no_owner: PasswordEntry[]
  unused: PasswordEntry[]
  expiring: PasswordEntry[]
  expired: PasswordEntry[]
}

function toQuery(filters: PasswordFilters): string {
  const params = new URLSearchParams()
  Object.entries(filters).forEach(([k, v]) => {
    if (v) params.set(k, v)
  })
  const s = params.toString()
  return s ? `?${s}` : ''
}

export const passwordsApi = {
  groups: () => apiFetch<PasswordGroup[]>('/api/password-groups'),

  list: (filters: PasswordFilters = {}) =>
    apiFetch<PasswordEntry[]>(`/api/passwords${toQuery(filters)}`),

  inventory: () => apiFetch<Inventory>('/api/passwords/inventory'),

  health: () => apiFetch<HealthReport>('/api/passwords/health'),

  create: (body: PasswordWrite) =>
    apiFetch<PasswordEntry>('/api/passwords', { method: 'POST', body }),

  update: (id: string, body: PasswordWrite) =>
    apiFetch<PasswordEntry>(`/api/passwords/${id}`, { method: 'PUT', body }),

  remove: (id: string) => apiFetch<void>(`/api/passwords/${id}`, { method: 'DELETE' }),

  reveal: (id: string) =>
    apiFetch<{ password: string }>(`/api/passwords/${id}/reveal`, { method: 'POST' }),
}
