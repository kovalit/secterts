import { apiFetch } from './client'
import type { Company } from '../types'

export const companiesApi = {
  list: () => apiFetch<Company[]>('/api/companies'),
  create: (name: string) => apiFetch<Company>('/api/companies', { method: 'POST', body: { name } }),
  update: (id: string, name: string) =>
    apiFetch<Company>(`/api/companies/${id}`, { method: 'PUT', body: { name } }),
  remove: (id: string) => apiFetch<void>(`/api/companies/${id}`, { method: 'DELETE' }),
}
