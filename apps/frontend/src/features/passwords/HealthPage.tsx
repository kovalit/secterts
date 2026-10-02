import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  AlertTriangle,
  CalendarClock,
  Clock,
  HeartPulse,
  ShieldAlert,
  ShieldCheck,
  UserX,
} from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader'
import { PageLoader } from '../../components/ui/Spinner'
import { EmptyState } from '../../components/ui/EmptyState'
import { EntryIcon } from '../../components/ui/GroupIcon'
import { PasswordEditorDrawer } from './PasswordEditorDrawer'
import { passwordsApi } from '../../api/passwords'
import { companiesApi } from '../../api/companies'
import { formatDate } from '../../lib/formatDate'
import { entryTypeLabel, strengthLabel } from '../../lib/entryTypes'
import type { PasswordEntry, PasswordGroup } from '../../types'

type Section = {
  key: string
  title: string
  hint: string
  icon: React.ReactNode
  tone: 'warn' | 'danger' | 'neutral'
  entries: PasswordEntry[]
  meta: (e: PasswordEntry) => string
}

export function HealthPage() {
  const healthQ = useQuery({ queryKey: ['passwords-health'], queryFn: passwordsApi.health })
  const groupsQ = useQuery({ queryKey: ['password-groups'], queryFn: passwordsApi.groups })
  const companiesQ = useQuery({ queryKey: ['companies'], queryFn: companiesApi.list })

  const [editing, setEditing] = useState<PasswordEntry | null>(null)
  const [editorOpen, setEditorOpen] = useState(false)

  const groups: PasswordGroup[] = groupsQ.data ?? []
  const companies = companiesQ.data ?? []

  const openEdit = (e: PasswordEntry) => {
    setEditing(e)
    setEditorOpen(true)
  }

  if (healthQ.isLoading) return <PageLoader />

  const report = healthQ.data
  const summary = report?.summary
  const issueTotal = summary
    ? summary.weak + summary.no_owner + summary.unused + summary.expiring + summary.expired
    : 0

  const allSections: Section[] = report
    ? [
        {
          key: 'expired',
          title: 'Истёкшие',
          hint: 'Срок действия уже прошёл — обновите или удалите.',
          icon: <AlertTriangle size={18} />,
          tone: 'danger',
          entries: report.expired,
          meta: (e) => `Истёк ${formatDate(e.expires_at)}`,
        },
        {
          key: 'weak',
          title: 'Слабые пароли',
          hint: 'Пароли легко подобрать — сгенерируйте новые.',
          icon: <ShieldAlert size={18} />,
          tone: 'danger',
          entries: report.weak,
          meta: (e) => `Надёжность: ${strengthLabel(e.password_strength)}`,
        },
        {
          key: 'expiring',
          title: 'Скоро истекают',
          hint: 'Срок действия закончится в ближайшие 30 дней.',
          icon: <CalendarClock size={18} />,
          tone: 'warn',
          entries: report.expiring,
          meta: (e) => `Истекает ${formatDate(e.expires_at)}`,
        },
        {
          key: 'no_owner',
          title: 'Без владельца',
          hint: 'Не назначен ответственный за запись.',
          icon: <UserX size={18} />,
          tone: 'neutral',
          entries: report.no_owner,
          meta: () => 'Владелец не указан',
        },
        {
          key: 'unused',
          title: 'Давно не использовались',
          hint: 'Записи, к которым давно не обращались (90+ дней).',
          icon: <Clock size={18} />,
          tone: 'neutral',
          entries: report.unused,
          meta: (e) =>
            e.last_used_at ? `Использован ${formatDate(e.last_used_at)}` : 'Ни разу не использовался',
        },
      ]
    : []
  const sections = allSections.filter((s) => s.entries.length > 0)

  return (
    <div>
      <PageHeader
        eyebrow="Здоровье секретов"
        title="Проверка здоровья"
        description="Слабые пароли, записи без владельца, давно неиспользуемые и истекающие секреты."
      />

      {summary && (
        <div className="grid gap-3 grid-cols-2 lg:grid-cols-5 mb-6">
          <SummaryTile icon={<AlertTriangle size={18} />} value={summary.expired} label="истекли" tone="danger" />
          <SummaryTile icon={<ShieldAlert size={18} />} value={summary.weak} label="слабые" tone="danger" />
          <SummaryTile icon={<CalendarClock size={18} />} value={summary.expiring} label="скоро истекут" tone="warn" />
          <SummaryTile icon={<UserX size={18} />} value={summary.no_owner} label="без владельца" tone="neutral" />
          <SummaryTile icon={<Clock size={18} />} value={summary.unused} label="не используются" tone="neutral" />
        </div>
      )}

      {issueTotal === 0 ? (
        <EmptyState
          icon={<ShieldCheck size={22} />}
          title="Проблем не найдено"
          hint="Все секреты надёжны, с владельцами и актуальными сроками. Отличная работа!"
        />
      ) : (
        <div className="flex flex-col gap-6">
          {sections.map((s) => (
            <div key={s.key} className="card p-4">
              <div className="flex items-center gap-2 mb-3">
                <span
                  className={
                    s.tone === 'danger'
                      ? 'text-danger-icon'
                      : s.tone === 'warn'
                        ? 'text-warn-icon'
                        : 'text-ink-muted'
                  }
                >
                  {s.icon}
                </span>
                <h2 className="text-[15px] font-semibold text-ink-strong">{s.title}</h2>
                <span className="badge badge-neutral">{s.entries.length}</span>
              </div>
              <p className="text-[13px] text-ink-muted mb-3">{s.hint}</p>
              <div className="flex flex-col divide-y divide-border">
                {s.entries.map((e) => (
                  <button
                    key={e.id}
                    type="button"
                    onClick={() => openEdit(e)}
                    className="flex items-center gap-3 py-2.5 text-left hover:bg-ink-strong/[.02] -mx-2 px-2 rounded-lg transition-colors"
                  >
                    <EntryIcon
                      iconUrl={e.custom_icon}
                      faviconUrl={e.favicon_url}
                      groupIcon={groups.find((g) => g.id === e.group_id)?.icon}
                    />
                    <div className="min-w-0 flex-1">
                      <div className="font-medium text-ink-strong truncate">{e.title}</div>
                      <div className="text-[12.5px] text-ink-muted truncate">
                        {entryTypeLabel(e.entry_type)}
                        {e.domain ? ` · ${e.domain}` : e.login ? ` · ${e.login}` : ''}
                      </div>
                    </div>
                    <span className="text-[12.5px] text-ink-muted shrink-0 text-right">{s.meta(e)}</span>
                  </button>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      <PasswordEditorDrawer
        open={editorOpen}
        onClose={() => setEditorOpen(false)}
        entry={editing}
        groups={groups}
        companies={companies}
      />
    </div>
  )
}

function SummaryTile({
  icon,
  value,
  label,
  tone,
}: {
  icon: React.ReactNode
  value: number
  label: string
  tone: 'warn' | 'danger' | 'neutral'
}) {
  const toneClass =
    tone === 'danger' ? 'text-danger-icon' : tone === 'warn' ? 'text-warn-icon' : 'text-ink-muted'
  return (
    <div className="card p-4 flex items-center gap-3">
      <span className={'shrink-0 ' + toneClass}>
        {value > 0 ? icon : <HeartPulse size={18} className="text-accent" />}
      </span>
      <div className="min-w-0">
        <div className="text-[20px] leading-none font-bold text-ink-strong">{value}</div>
        <div className="text-[12.5px] text-ink-muted mt-1 truncate">{label}</div>
      </div>
    </div>
  )
}
