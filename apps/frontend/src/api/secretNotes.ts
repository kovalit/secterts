import { apiFetch } from './client'
import type { SecretNote, SecretNoteType } from '../types'

export type SecretNoteFilters = {
  q?: string
  type?: string
}

export type SecretNoteWrite = {
  title: string
  type: SecretNoteType
  text?: string
}

function toQuery(filters: SecretNoteFilters): string {
  const params = new URLSearchParams()
  Object.entries(filters).forEach(([k, v]) => {
    if (v) params.set(k, v)
  })
  const s = params.toString()
  return s ? `?${s}` : ''
}

export const secretNotesApi = {
  list: (filters: SecretNoteFilters = {}) =>
    apiFetch<SecretNote[]>(`/api/secret-notes${toQuery(filters)}`),

  create: (body: SecretNoteWrite) =>
    apiFetch<SecretNote>('/api/secret-notes', { method: 'POST', body }),

  update: (id: string, body: SecretNoteWrite) =>
    apiFetch<SecretNote>(`/api/secret-notes/${id}`, { method: 'PUT', body }),

  remove: (id: string) => apiFetch<void>(`/api/secret-notes/${id}`, { method: 'DELETE' }),

  reveal: (id: string) =>
    apiFetch<{ text: string }>(`/api/secret-notes/${id}/reveal`, { method: 'POST' }),
}
