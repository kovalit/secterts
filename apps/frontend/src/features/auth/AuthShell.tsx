import type { ReactNode } from 'react'
import { Lock } from 'lucide-react'

export function AuthShell({
  title,
  subtitle,
  children,
  footer,
}: {
  title: string
  subtitle?: string
  children: ReactNode
  footer?: ReactNode
}) {
  return (
    <div className="min-h-screen grid place-items-center px-4 py-10">
      <div className="w-full max-w-[420px]">
        <div className="flex items-center gap-2.5 mb-6 justify-center">
          <span
            className="grid place-items-center w-9 h-9 rounded-lg text-white"
            style={{
              background: 'linear-gradient(135deg,#3B82F6,#1D4ED8)',
              boxShadow: 'inset 0 1px 0 rgba(255,255,255,.25), 0 2px 6px rgba(37,99,235,.35)',
            }}
          >
            <Lock size={18} />
          </span>
          <span className="text-[18px] font-bold text-ink-strong">Secrets Center</span>
        </div>

        <div className="card p-6 shadow-shell">
          <h1 className="text-[22px] font-bold text-ink-strong">{title}</h1>
          {subtitle && <p className="mt-1.5 text-[14px] text-ink-muted">{subtitle}</p>}
          <div className="mt-5">{children}</div>
        </div>

        {footer && <div className="mt-4 text-center text-[14px] text-ink-muted">{footer}</div>}
      </div>
    </div>
  )
}
