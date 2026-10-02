import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import {
  AlertTriangle,
  Boxes,
  CalendarClock,
  Clock,
  KeyRound,
  ShieldAlert,
  UserX,
} from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader'
import { PageLoader } from '../../components/ui/Spinner'
import { EmptyState } from '../../components/ui/EmptyState'
import { passwordsApi } from '../../api/passwords'
import { entryTypePlural } from '../../lib/entryTypes'

function StatTile({
  icon,
  value,
  label,
  tone = 'neutral',
}: {
  icon: React.ReactNode
  value: number
  label: string
  tone?: 'neutral' | 'warn' | 'danger'
}) {
  const toneClass =
    tone === 'danger'
      ? 'text-danger-icon'
      : tone === 'warn'
        ? 'text-warn-icon'
        : 'text-accent'
  return (
    <div className="card p-4 flex items-center gap-3">
      <span className={'shrink-0 ' + toneClass}>{icon}</span>
      <div className="min-w-0">
        <div className="text-[22px] leading-none font-bold text-ink-strong">{value}</div>
        <div className="text-[13px] text-ink-muted mt-1">{label}</div>
      </div>
    </div>
  )
}

export function InventoryPage() {
  const navigate = useNavigate()
  const invQ = useQuery({ queryKey: ['passwords-inventory'], queryFn: passwordsApi.inventory })

  if (invQ.isLoading) return <PageLoader />

  const inv = invQ.data

  return (
    <div>
      <PageHeader
        eyebrow="Инвентаризация"
        title="Инвентаризация секретов"
        description="Вся инфраструктура по типам записей, со сроками истечения и состоянием здоровья."
      />

      {!inv || inv.total === 0 ? (
        <EmptyState
          icon={<Boxes size={22} />}
          title="Записей пока нет"
          hint="Добавьте секреты — здесь появится инвентаризация по типам и срокам."
        />
      ) : (
        <div className="flex flex-col gap-6">
          {/* By type */}
          <div>
            <h2 className="text-[15px] font-semibold text-ink-strong mb-3">По типам ({inv.total})</h2>
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {inv.by_type.map((it) => (
                <button
                  key={it.type}
                  type="button"
                  onClick={() => navigate(`/passwords?entry_type=${it.type}`)}
                  className="card p-4 flex items-center gap-3 text-left hover:border-accent-200 transition-colors"
                >
                  <span className="shrink-0 text-accent">
                    <KeyRound size={20} />
                  </span>
                  <div className="min-w-0">
                    <div className="text-[22px] leading-none font-bold text-ink-strong">{it.count}</div>
                    <div className="text-[13px] text-ink-muted mt-1 truncate">{entryTypePlural(it.type)}</div>
                  </div>
                </button>
              ))}
            </div>
          </div>

          {/* Expiration + health roll-up */}
          <div>
            <h2 className="text-[15px] font-semibold text-ink-strong mb-3">Сроки и здоровье</h2>
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              <StatTile
                icon={<CalendarClock size={20} />}
                value={inv.expiring_this_month}
                label="истекают в этом месяце"
                tone="warn"
              />
              <StatTile
                icon={<CalendarClock size={20} />}
                value={inv.expiring_soon}
                label="истекают в ближайшие 30 дней"
                tone="warn"
              />
              <StatTile
                icon={<AlertTriangle size={20} />}
                value={inv.expired}
                label="уже истекли"
                tone="danger"
              />
              <StatTile
                icon={<ShieldAlert size={20} />}
                value={inv.weak_passwords}
                label="слабые пароли"
                tone="danger"
              />
              <StatTile
                icon={<UserX size={20} />}
                value={inv.without_owner}
                label="без владельца"
                tone="neutral"
              />
              <StatTile
                icon={<Clock size={20} />}
                value={inv.unused}
                label="давно не использовались"
                tone="neutral"
              />
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
