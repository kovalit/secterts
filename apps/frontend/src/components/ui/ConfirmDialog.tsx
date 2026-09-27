import { TriangleAlert } from 'lucide-react'
import type { ReactNode } from 'react'

// A small confirm modal used before destructive actions.
export function ConfirmDialog({
  open,
  title,
  message,
  confirmLabel = 'Удалить',
  cancelLabel = 'Отмена',
  onConfirm,
  onCancel,
  busy,
}: {
  open: boolean
  title: string
  message: ReactNode
  confirmLabel?: string
  cancelLabel?: string
  onConfirm: () => void
  onCancel: () => void
  busy?: boolean
}) {
  if (!open) return null
  return (
    <div className="fixed inset-0 z-[60] grid place-items-center px-4">
      <div className="absolute inset-0 bg-ink-strong/30" onClick={onCancel} aria-hidden="true" />
      <div className="card relative w-full max-w-md p-5 shadow-pop">
        <div className="flex items-start gap-3">
          <span className="grid place-items-center w-10 h-10 rounded-xl bg-danger-bg text-danger-icon shrink-0">
            <TriangleAlert size={20} />
          </span>
          <div className="min-w-0">
            <h3 className="text-[16px] font-semibold text-ink-strong">{title}</h3>
            <div className="mt-1 text-[14px] text-ink-muted leading-relaxed">{message}</div>
          </div>
        </div>
        <div className="flex justify-end gap-2 mt-5">
          <button className="btn" onClick={onCancel} disabled={busy}>
            {cancelLabel}
          </button>
          <button className="btn btn-danger" onClick={onConfirm} disabled={busy}>
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  )
}
