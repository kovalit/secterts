import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from 'react'
import { authApi } from '../../api/auth'
import { ApiError, UNAUTHORIZED_EVENT } from '../../api/client'
import type { User } from '../../types'

type AuthContextValue = {
  user: User | null
  loading: boolean
  setUser: (u: User | null) => void
  refresh: () => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  const refresh = useCallback(async () => {
    try {
      const me = await authApi.me()
      setUser(me)
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        setUser(null)
      }
    } finally {
      setLoading(false)
    }
  }, [])

  const logout = useCallback(async () => {
    try {
      await authApi.logout()
    } finally {
      setUser(null)
    }
  }, [])

  useEffect(() => {
    void refresh()
  }, [refresh])

  // Global 401 → drop the session so ProtectedRoute redirects to /login.
  useEffect(() => {
    const onUnauthorized = () => setUser(null)
    window.addEventListener(UNAUTHORIZED_EVENT, onUnauthorized)
    return () => window.removeEventListener(UNAUTHORIZED_EVENT, onUnauthorized)
  }, [])

  return (
    <AuthContext.Provider value={{ user, loading, setUser, refresh, logout }}>
      {children}
    </AuthContext.Provider>
  )
}
