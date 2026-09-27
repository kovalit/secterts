import { useQuery } from '@tanstack/react-query'
import { List } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader'
import { EmptyState } from '../../components/ui/EmptyState'
import { PageLoader } from '../../components/ui/Spinner'
import { apiFetch } from '../../api/client'
import { formatDate } from '../../lib/formatDate'
import type { AuditLog } from '../../types'

// Human labels for audit actions.
const ACTION_LABELS: Record<string, string> = {
  login_success: 'Успешный вход',
  login_failed: 'Неуспешный вход',
  email_2fa_sent: 'Отправлен код 2FA',
  password_changed: 'Смена пароля',
  password_entry_created: 'Создан пароль',
  password_entry_updated: 'Изменён пароль',
  password_entry_deleted: 'Удалён пароль',
  password_entry_revealed: 'Показан пароль',
  password_entry_copied_by_extension: 'Скопирован пароль (расширение)',
  app_secret_created: 'Создан секрет',
  app_secret_updated: 'Изменён секрет',
  app_secret_deleted: 'Удалён секрет',
  app_secret_revealed: 'Показан секрет',
  extension_token_created: 'Создан токен расширения',
  extension_token_revoked: 'Отозван токен расширения',
  backup_export_created: 'Создан бэкап',
}

export function AuditPage() {
  const q = useQuery({ queryKey: ['audit-logs'], queryFn: () => apiFetch<AuditLog[]>('/api/audit-logs') })
  const logs = q.data ?? []

  return (
    <div>
      <PageHeader
        eyebrow="Настройки"
        title="Журнал действий"
        description="История действий с паролями, секретами и токенами."
      />
      {q.isLoading ? (
        <PageLoader />
      ) : logs.length === 0 ? (
        <EmptyState icon={<List size={22} />} title="Событий пока нет" />
      ) : (
        <div className="card p-4 overflow-x-auto">
          <table className="data-table">
            <thead>
              <tr>
                <th>Действие</th>
                <th>Объект</th>
                <th>IP</th>
                <th>Время</th>
              </tr>
            </thead>
            <tbody>
              {logs.map((l) => (
                <tr key={l.id}>
                  <td className="text-ink-strong font-medium">{ACTION_LABELS[l.action] ?? l.action}</td>
                  <td className="text-ink-muted">{l.entity_type}</td>
                  <td className="text-ink-muted font-mono text-[13px]">{l.ip ?? '—'}</td>
                  <td className="text-ink-muted">{formatDate(l.created_at)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
