import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { KeyRound, Laptop, ShieldCheck } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader'
import { PageLoader, Spinner } from '../../components/ui/Spinner'
import { useToast } from '../../components/ui/Toast'
import { useAuth } from '../auth/AuthProvider'
import { authApi } from '../../api/auth'
import { ApiError } from '../../api/client'
import { formatDate } from '../../lib/formatDate'

export function SecurityPage() {
  const { user } = useAuth()
  const qc = useQueryClient()
  const toast = useToast()

  const [current, setCurrent] = useState('')
  const [next, setNext] = useState('')

  const sessionsQ = useQuery({ queryKey: ['sessions'], queryFn: authApi.sessions })

  const changePwMutation = useMutation({
    mutationFn: () => authApi.changePassword(current, next),
    onSuccess: () => {
      toast.success('Пароль изменён')
      setCurrent('')
      setNext('')
    },
    onError: (err) => toast.error(err instanceof ApiError ? err.message : 'Не удалось изменить пароль'),
  })

  const revokeMutation = useMutation({
    mutationFn: (id: string) => authApi.revokeSession(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['sessions'] })
      toast.success('Сессия завершена')
    },
    onError: () => toast.error('Не удалось завершить сессию'),
  })

  const sessions = (sessionsQ.data ?? []).filter((s) => !s.revoked_at)

  return (
    <div>
      <PageHeader eyebrow="Настройки" title="Безопасность" description="Аккаунт, двухфакторная защита и активные сессии." />

      {/* Account */}
      <div className="card p-5 mb-5">
        <h2 className="text-[15px] font-semibold text-ink-strong mb-4">Аккаунт</h2>
        <dl className="grid sm:grid-cols-2 gap-4 text-[14px]">
          <div>
            <dt className="text-ink-muted mb-0.5">Email</dt>
            <dd className="text-ink-strong font-medium">{user?.email}</dd>
          </div>
          <div>
            <dt className="text-ink-muted mb-0.5">Email 2FA</dt>
            <dd>
              <span className="badge badge-accent">
                <ShieldCheck size={13} />
                Включена
              </span>
            </dd>
          </div>
        </dl>
      </div>

      {/* Change password */}
      <div className="card p-5 mb-5">
        <h2 className="text-[15px] font-semibold text-ink-strong mb-4 flex items-center gap-2">
          <KeyRound size={16} className="text-accent" />
          Смена пароля
        </h2>
        <form
          className="grid sm:grid-cols-2 gap-4 max-w-xl"
          onSubmit={(e) => {
            e.preventDefault()
            if (current && next.length >= 8) changePwMutation.mutate()
          }}
        >
          <div>
            <label className="field-label">Текущий пароль</label>
            <input className="input" type="password" value={current} onChange={(e) => setCurrent(e.target.value)} />
          </div>
          <div>
            <label className="field-label">Новый пароль (мин. 8)</label>
            <input className="input" type="password" value={next} onChange={(e) => setNext(e.target.value)} />
          </div>
          <div className="sm:col-span-2">
            <button className="btn btn-primary" type="submit" disabled={changePwMutation.isPending || !current || next.length < 8}>
              {changePwMutation.isPending ? <Spinner size={16} /> : 'Изменить пароль'}
            </button>
          </div>
        </form>
      </div>

      {/* Sessions */}
      <div className="card p-5">
        <h2 className="text-[15px] font-semibold text-ink-strong mb-4 flex items-center gap-2">
          <Laptop size={16} className="text-accent" />
          Активные сессии
        </h2>
        {sessionsQ.isLoading ? (
          <PageLoader />
        ) : (
          <div className="overflow-x-auto">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Устройство</th>
                  <th>IP</th>
                  <th>Создана</th>
                  <th className="text-right">Действия</th>
                </tr>
              </thead>
              <tbody>
                {sessions.map((s) => (
                  <tr key={s.id}>
                    <td className="text-ink">
                      <span className="truncate max-w-[280px] inline-block align-middle">
                        {s.user_agent ?? 'Неизвестное устройство'}
                      </span>
                      {s.current && <span className="badge badge-accent ml-2">текущая</span>}
                    </td>
                    <td className="text-ink-muted font-mono text-[13px]">{s.ip ?? '—'}</td>
                    <td className="text-ink-muted">{formatDate(s.created_at)}</td>
                    <td className="text-right">
                      {!s.current && (
                        <button
                          className="btn btn-sm btn-danger"
                          onClick={() => revokeMutation.mutate(s.id)}
                          disabled={revokeMutation.isPending}
                        >
                          Завершить
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
