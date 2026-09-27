import type { ReactNode } from 'react'

export function PageHeader({
  eyebrow,
  title,
  description,
  action,
}: {
  eyebrow?: string
  title: string
  description?: string
  action?: ReactNode
}) {
  return (
    <div className="flex flex-wrap items-start justify-between gap-4 mb-6">
      <div className="min-w-0">
        {eyebrow && (
          <div className="text-[13px] font-semibold text-accent mb-1.5">{eyebrow}</div>
        )}
        <h1 className="text-[28px] leading-tight font-bold text-ink-strong tracking-tight">
          {title}
        </h1>
        {description && <p className="mt-2 text-[15px] text-ink-muted max-w-2xl">{description}</p>}
      </div>
      {action && <div className="shrink-0">{action}</div>}
    </div>
  )
}
