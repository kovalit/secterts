import { apiFetch } from './client'
import type { Session, User } from '../types'

export type LoginResponse = {
  challenge_id: string
  requires_2fa: boolean
  email_2fa: boolean
  message: string
}

export const authApi = {
  register: (email: string, password: string) =>
    apiFetch<User>('/api/auth/register', { method: 'POST', body: { email, password } }),

  login: (email: string, password: string) =>
    apiFetch<LoginResponse>('/api/auth/login', { method: 'POST', body: { email, password } }),

  verify: (challenge_id: string, code: string) =>
    apiFetch<User>('/api/auth/verify-email-code', { method: 'POST', body: { challenge_id, code } }),

  logout: () => apiFetch<void>('/api/auth/logout', { method: 'POST' }),

  me: () => apiFetch<User>('/api/auth/me'),

  changePassword: (current_password: string, new_password: string) =>
    apiFetch<void>('/api/auth/change-password', {
      method: 'POST',
      body: { current_password, new_password },
    }),

  sessions: () => apiFetch<Session[]>('/api/auth/sessions'),

  revokeSession: (id: string) => apiFetch<void>(`/api/auth/sessions/${id}`, { method: 'DELETE' }),

  passwordResetRequest: (email: string) =>
    apiFetch<{ challenge_id: string }>('/api/auth/password-reset/request', {
      method: 'POST',
      body: { email },
    }),

  passwordResetConfirm: (challenge_id: string, code: string, new_password: string) =>
    apiFetch<void>('/api/auth/password-reset/confirm', {
      method: 'POST',
      body: { challenge_id, code, new_password },
    }),
}
