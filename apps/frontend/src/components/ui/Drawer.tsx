import { useEffect, type ReactNode } from 'react'
import { X } from 'lucide-react'

// A right-hand drawer used for create/edit forms.
export function Drawer({
  open,
  onClose,
  title,
  subtitle,
  children,
  footer,
}: {
  open: boolean
  onClose: () => void
  title: string
  subtitle?: string
  children: ReactNode
  footer?: ReactNode
}) {
  useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [open, onClose])

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex justify-end">
      <div
        className="absolute inset-0 bg-ink-strong/30"
        onClick={onClose}
        aria-hidden="true"
      />
      <aside className="relative w-[min(480px,100vw)] h-full bg-white border-l border-border shadow-pop flex flex-col animate-[slideIn_.2s_ease]">
        <header className="flex items-start justify-between gap-4 px-5 py-4 border-b border-border">
          <div className="min-w-0">
            <h2 className="text-[17px] font-semibold text-ink-strong truncate">{title}</h2>
            {subtitle && <p className="text-[13px] text-ink-muted mt-0.5">{subtitle}</p>}
          </div>
          <button className="icon-btn" onClick={onClose} aria-label="Закрыть">
            <X size={18} />
          </button>
        </header>
        <div className="flex-1 overflow-y-auto px-5 py-5">{children}</div>
        {footer && (
          <footer className="flex items-center justify-end gap-2 px-5 py-4 border-t border-border bg-surface-muted">
            {footer}
          </footer>
        )}
      </aside>
    </div>
  )
}
