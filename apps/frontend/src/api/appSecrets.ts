import { apiFetch } from './client'
import type { AppEnvironment, AppProject, AppSecret } from '../types'

export type ProjectWrite = {
  name: string
  description?: string | null
  company_id?: string | null
}

export type SecretWrite = {
  environment_id?: string
  key?: string
  value?: string
  comment?: string | null
}

export type BulkSecretItem = {
  key: string
  value: string
  comment?: string | null
}

export type BulkSecretsWrite = {
  environment: string
  items: BulkSecretItem[]
}

export const appSecretsApi = {
  listProjects: () => apiFetch<AppProject[]>('/api/app-projects'),
  getProject: (id: string) => apiFetch<AppProject>(`/api/app-projects/${id}`),
  createProject: (body: ProjectWrite) =>
    apiFetch<AppProject>('/api/app-projects', { method: 'POST', body }),
  updateProject: (id: string, body: ProjectWrite) =>
    apiFetch<AppProject>(`/api/app-projects/${id}`, { method: 'PUT', body }),
  removeProject: (id: string) => apiFetch<void>(`/api/app-projects/${id}`, { method: 'DELETE' }),

  listEnvironments: (projectId: string) =>
    apiFetch<AppEnvironment[]>(`/api/app-projects/${projectId}/environments`),

  listSecrets: (projectId: string, env: string) =>
    apiFetch<AppSecret[]>(`/api/app-projects/${projectId}/secrets?env=${encodeURIComponent(env)}`),
  createSecret: (projectId: string, body: SecretWrite) =>
    apiFetch<AppSecret>(`/api/app-projects/${projectId}/secrets`, { method: 'POST', body }),
  bulkCreateSecrets: (projectId: string, body: BulkSecretsWrite) =>
    apiFetch<AppSecret[]>(`/api/app-projects/${projectId}/secrets/bulk`, { method: 'POST', body }),
  updateSecret: (id: string, body: SecretWrite) =>
    apiFetch<AppSecret>(`/api/app-secrets/${id}`, { method: 'PUT', body }),
  removeSecret: (id: string) => apiFetch<void>(`/api/app-secrets/${id}`, { method: 'DELETE' }),
  revealSecret: (id: string) =>
    apiFetch<{ value: string }>(`/api/app-secrets/${id}/reveal`, { method: 'POST' }),
}
