const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

// ApiError carries the backend error code + message.
export class ApiError extends Error {
  code: string
  status: number
  constructor(status: number, code: string, message: string) {
    super(message)
    this.status = status
    this.code = code
  }
}

// Event name dispatched on 401 so the app can force a logout.
export const UNAUTHORIZED_EVENT = 'sc:unauthorized'

type FetchOptions = Omit<RequestInit, 'body'> & { body?: unknown }

export async function apiFetch<T>(path: string, options: FetchOptions = {}): Promise<T> {
  const { body, headers, ...rest } = options
  const res = await fetch(`${API_URL}${path}`, {
    ...rest,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...(headers ?? {}),
    },
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })

  if (res.status === 401) {
    // Let listeners (auth provider) react, but still throw for the caller.
    window.dispatchEvent(new CustomEvent(UNAUTHORIZED_EVENT))
  }

  if (res.status === 204) {
    return undefined as T
  }

  const isJson = res.headers.get('content-type')?.includes('application/json')
  const payload = isJson ? await res.json().catch(() => null) : null

  if (!res.ok) {
    const code = (payload && payload.code) || 'error'
    const message = (payload && payload.message) || `Request failed (${res.status})`
    throw new ApiError(res.status, code, message)
  }

  return payload as T
}

export { API_URL }
