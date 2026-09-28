import { useState } from 'react'
import { Outlet } from 'react-router-dom'
import { Lock, LogOut, Menu, X } from 'lucide-react'
import { Sidebar } from './Sidebar'
import { useAuth } from '../../features/auth/AuthProvider'
import { useToast } from '../ui/Toast'

export function AppLayout() {
  const { user, logout } = useAuth()
  const toast = useToast()
  const [mobileOpen, setMobileOpen] = useState(false)

  const onLogout = async () => {
    await logout()
    toast.info('Вы вышли из системы')
  }

  return (
    <div className="min-h-screen">
      <div className="mx-auto max-w-[1440px] min-h-screen bg-surface md:border-x border-border shadow-shell">
        {/* Header */}
        <header className="sticky top-0 z-30 h-16 flex items-center gap-3 px-4 md:px-6 bg-white/90 backdrop-blur border-b border-border">
          <button
            className="icon-btn md:hidden"
            onClick={() => setMobileOpen((v) => !v)}
            aria-label="Меню"
          >
            {mobileOpen ? <X size={20} /> : <Menu size={20} />}
          </button>

          {/* Brand */}
          <div className="flex items-center gap-2.5">
            <span
              className="grid place-items-center w-8 h-8 rounded-lg text-white shrink-0"
              style={{
                background: 'linear-gradient(135deg,#3B82F6,#1D4ED8)',
                boxShadow: 'inset 0 1px 0 rgba(255,255,255,.25), 0 2px 6px rgba(37,99,235,.35)',
              }}
            >
              <Lock size={16} />
            </span>
            <div className="leading-tight">
              <div className="text-[15px] font-bold text-ink-strong">Secrets Center</div>
              <div className="text-[11px] text-ink-muted hidden sm:block">Хранилище секретов</div>
            </div>
          </div>

          <div className="flex-1" />
          <div className="flex items-center gap-3">
            {user && (
              <span className="hidden sm:inline text-[13px] text-ink-muted">{user.email}</span>
            )}
            <button className="btn btn-sm" onClick={onLogout}>
              <LogOut size={15} />
              Выйти
            </button>
          </div>
        </header>

        <div className="grid md:grid-cols-[264px_minmax(0,1fr)]">
          {/* Desktop sidebar */}
          <aside className="hidden md:block bg-surface-muted border-r border-border">
            <div className="sticky top-16 max-h-[calc(100vh-4rem)] overflow-y-auto">
              <Sidebar />
            </div>
          </aside>

          {/* Mobile sidebar */}
          {mobileOpen && (
            <>
              <div
                className="fixed inset-0 top-16 z-40 bg-ink-strong/30 md:hidden"
                onClick={() => setMobileOpen(false)}
              />
              <aside className="fixed top-16 left-0 bottom-0 z-40 w-[min(300px,85vw)] bg-surface-muted border-r border-border shadow-pop md:hidden overflow-y-auto">
                <Sidebar onNavigate={() => setMobileOpen(false)} />
              </aside>
            </>
          )}

          {/* Content */}
          <main className="min-w-0 px-4 md:px-8 py-8">
            <Outlet />
          </main>
        </div>
      </div>
    </div>
  )
}
