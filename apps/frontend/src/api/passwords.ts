import { apiFetch } from './client'
import type { PasswordEntry, PasswordGroup } from '../types'

export type PasswordFilters = {
  q?: string
  scope?: string
  company_id?: string
  group_id?: string
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

  create: (body: PasswordWrite) =>
    apiFetch<PasswordEntry>('/api/passwords', { method: 'POST', body }),

  update: (id: string, body: PasswordWrite) =>
    apiFetch<PasswordEntry>(`/api/passwords/${id}`, { method: 'PUT', body }),

  remove: (id: string) => apiFetch<void>(`/api/passwords/${id}`, { method: 'DELETE' }),

  reveal: (id: string) =>
    apiFetch<{ password: string }>(`/api/passwords/${id}/reveal`, { method: 'POST' }),
}
