import { apiFetch } from './client'
import type { ExtensionToken } from '../types'

export const extensionApi = {
  list: () => apiFetch<ExtensionToken[]>('/api/extension/tokens'),
  create: (name: string) =>
    apiFetch<{ token: string; token_info: ExtensionToken }>('/api/extension/tokens', {
      method: 'POST',
      body: { name },
    }),
  revoke: (id: string) => apiFetch<void>(`/api/extension/tokens/${id}`, { method: 'DELETE' }),
}
