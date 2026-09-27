import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from 'react'
import { Check, Info, TriangleAlert, X } from 'lucide-react'

type ToastKind = 'success' | 'error' | 'info'
type Toast = { id: number; kind: ToastKind; message: string }

type ToastContextValue = {
  success: (message: string) => void
  error: (message: string) => void
  info: (message: string) => void
}

const ToastContext = createContext<ToastContextValue | null>(null)

export function useToast(): ToastContextValue {
  const ctx = useContext(ToastContext)
  if (!ctx) throw new Error('useToast must be used within ToastProvider')
  return ctx
}

let nextId = 1

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([])

  const remove = useCallback((id: number) => {
    setToasts((t) => t.filter((x) => x.id !== id))
  }, [])

  const push = useCallback(
    (kind: ToastKind, message: string) => {
      const id = nextId++
      setToasts((t) => [...t, { id, kind, message }])
      window.setTimeout(() => remove(id), 3500)
    },
    [remove],
  )

  const value = useMemo<ToastContextValue>(
    () => ({
      success: (m) => push('success', m),
      error: (m) => push('error', m),
      info: (m) => push('info', m),
    }),
    [push],
  )

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div className="fixed bottom-4 right-4 z-[100] flex flex-col gap-2 w-[min(360px,calc(100vw-2rem))]">
        {toasts.map((t) => (
          <div
            key={t.id}
            className="card flex items-start gap-3 p-3.5 shadow-pop animate-[fadeIn_.15s_ease]"
            role="status"
          >
            <span
              className={
                'grid place-items-center w-6 h-6 rounded-full shrink-0 ' +
                (t.kind === 'success'
                  ? 'bg-accent-50 text-accent'
                  : t.kind === 'error'
                    ? 'bg-danger-bg text-danger-icon'
                    : 'bg-surface-sunken text-ink-muted')
              }
            >
              {t.kind === 'success' ? (
                <Check size={15} />
              ) : t.kind === 'error' ? (
                <TriangleAlert size={15} />
              ) : (
                <Info size={15} />
              )}
            </span>
            <p className="text-[14px] text-ink flex-1 leading-snug">{t.message}</p>
            <button className="icon-btn w-6 h-6" onClick={() => remove(t.id)} aria-label="Закрыть">
              <X size={15} />
            </button>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  )
}
