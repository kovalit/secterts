import type { ReactNode } from 'react'

export function EmptyState({
  icon,
  title,
  hint,
  action,
}: {
  icon?: ReactNode
  title: string
  hint?: string
  action?: ReactNode
}) {
  return (
    <div className="card flex flex-col items-center text-center gap-3 py-14 px-6">
      {icon && (
        <div className="grid place-items-center w-12 h-12 rounded-xl bg-accent-50 text-accent">
          {icon}
        </div>
      )}
      <div>
        <p className="text-[15px] font-semibold text-ink-strong">{title}</p>
        {hint && <p className="mt-1 text-[14px] text-ink-muted max-w-sm">{hint}</p>}
      </div>
      {action}
    </div>
  )
}
