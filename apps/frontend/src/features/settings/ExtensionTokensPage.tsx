import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Copy, Puzzle, ShieldAlert, Trash2 } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader'
import { EmptyState } from '../../components/ui/EmptyState'
import { PageLoader, Spinner } from '../../components/ui/Spinner'
import { ConfirmDialog } from '../../components/ui/ConfirmDialog'
import { useToast } from '../../components/ui/Toast'
import { extensionApi } from '../../api/extension'
import { copyToClipboard } from '../../lib/clipboard'
import { formatDate } from '../../lib/formatDate'
import type { ExtensionToken } from '../../types'

export function ExtensionTokensPage() {
  const qc = useQueryClient()
  const toast = useToast()
  const [name, setName] = useState('')
  const [freshToken, setFreshToken] = useState<string | null>(null)
  const [toRevoke, setToRevoke] = useState<ExtensionToken | null>(null)

  const listQ = useQuery({ queryKey: ['extension-tokens'], queryFn: extensionApi.list })

  const createMutation = useMutation({
    mutationFn: () => extensionApi.create(name.trim() || 'Chrome Extension'),
    onSuccess: (res) => {
      qc.invalidateQueries({ queryKey: ['extension-tokens'] })
      setFreshToken(res.token)
      setName('')
      toast.success('Токен создан — скопируйте его сейчас')
    },
    onError: () => toast.error('Не удалось создать токен'),
  })

  const revokeMutation = useMutation({
    mutationFn: (id: string) => extensionApi.revoke(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['extension-tokens'] })
      toast.success('Токен отозван')
      setToRevoke(null)
    },
    onError: () => toast.error('Не удалось отозвать токен'),
  })

  const copyToken = async () => {
    if (!freshToken) return
    const ok = await copyToClipboard(freshToken)
    if (ok) toast.success('Токен скопирован')
  }

  const tokens = (listQ.data ?? []).filter((t) => !t.revoked_at)

  return (
    <div>
      <PageHeader
        eyebrow="Настройки"
        title="Chrome extension"
        description="Токен нужен расширению для чтения паролей по домену текущей вкладки."
      />

      <div className="callout callout-warn mb-5">
        <ShieldAlert size={20} className="text-warn-icon" />
        <div>
          <p className="callout-title">Токен даёт доступ на чтение паролей</p>
          <p>
            Он позволяет расширению искать записи по домену и раскрывать пароль по явному клику. Храните
            токен как секрет и отзывайте, если устройство потеряно.
          </p>
        </div>
      </div>

      {/* Create */}
      <div className="card p-4 mb-5">
        <div className="flex flex-wrap items-end gap-3">
          <div className="flex-1 min-w-[200px]">
            <label className="field-label">Название токена</label>
            <input
              className="input"
              placeholder="Например, Рабочий ноутбук"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <button className="btn btn-primary" onClick={() => createMutation.mutate()} disabled={createMutation.isPending}>
            {createMutation.isPending ? <Spinner size={16} /> : 'Создать токен'}
          </button>
        </div>

        {freshToken && (
          <div className="callout callout-info mt-4">
            <Copy size={20} className="text-accent" />
            <div className="min-w-0">
              <p className="callout-title">Скопируйте токен — он показывается только один раз</p>
              <div className="flex items-center gap-2 mt-2">
                <code className="keychip flex-1 truncate">{freshToken}</code>
                <button className="btn btn-sm" onClick={copyToken}>
                  <Copy size={14} />
                  Копировать
                </button>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* List */}
      {listQ.isLoading ? (
        <PageLoader />
      ) : tokens.length === 0 ? (
        <EmptyState icon={<Puzzle size={22} />} title="Активных токенов нет" hint="Создайте токен для расширения выше." />
      ) : (
        <div className="card p-4 overflow-x-auto">
          <table className="data-table">
            <thead>
              <tr>
                <th>Название</th>
                <th>Создан</th>
                <th>Использован</th>
                <th>Истекает</th>
                <th className="text-right">Действия</th>
              </tr>
            </thead>
            <tbody>
              {tokens.map((t) => (
                <tr key={t.id}>
                  <td>
                    <div className="flex items-center gap-3">
                      <span className="grid place-items-center w-8 h-8 rounded-md bg-accent-50 text-accent">
                        <Puzzle size={16} />
                      </span>
                      <span className="font-medium text-ink-strong">{t.name}</span>
                    </div>
                  </td>
                  <td className="text-ink-muted">{formatDate(t.created_at)}</td>
                  <td className="text-ink-muted">{t.last_used_at ? formatDate(t.last_used_at) : 'ещё нет'}</td>
                  <td className="text-ink-muted">{formatDate(t.expires_at)}</td>
                  <td>
                    <div className="flex justify-end">
                      <button
                        className="icon-btn w-8 h-8 hover:text-danger-icon"
                        onClick={() => setToRevoke(t)}
                        title="Отозвать"
                      >
                        <Trash2 size={15} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <ConfirmDialog
        open={!!toRevoke}
        title="Отозвать токен?"
        confirmLabel="Отозвать"
        message={
          <>
            Токен <strong>{toRevoke?.name}</strong> перестанет работать во всех расширениях, где он используется.
          </>
        }
        busy={revokeMutation.isPending}
        onCancel={() => setToRevoke(null)}
        onConfirm={() => toRevoke && revokeMutation.mutate(toRevoke.id)}
      />
    </div>
  )
}
