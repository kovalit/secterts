import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, KeySquare, Pencil, Plus, Trash2 } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader'
import { EmptyState } from '../../components/ui/EmptyState'
import { PageLoader, Spinner } from '../../components/ui/Spinner'
import { ConfirmDialog } from '../../components/ui/ConfirmDialog'
import { Drawer } from '../../components/ui/Drawer'
import { RevealValue } from '../passwords/RevealValue'
import { useToast } from '../../components/ui/Toast'
import { appSecretsApi } from '../../api/appSecrets'
import { ApiError } from '../../api/client'
import { formatDate } from '../../lib/formatDate'
import type { AppSecret } from '../../types'

export function ProjectDetailPage() {
  const { projectId = '' } = useParams()
  const qc = useQueryClient()
  const toast = useToast()

  const [env, setEnv] = useState('')
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<AppSecret | null>(null)
  const [key, setKey] = useState('')
  const [value, setValue] = useState('')
  const [comment, setComment] = useState('')
  const [toDelete, setToDelete] = useState<AppSecret | null>(null)

  const projectQ = useQuery({ queryKey: ['app-project', projectId], queryFn: () => appSecretsApi.getProject(projectId) })
  const envsQ = useQuery({
    queryKey: ['app-envs', projectId],
    queryFn: () => appSecretsApi.listEnvironments(projectId),
  })

  const envs = envsQ.data ?? []
  useEffect(() => {
    if (!env && envs.length > 0) setEnv(envs[0].name)
  }, [envs, env])

  const secretsQ = useQuery({
    queryKey: ['app-secrets', projectId, env],
    queryFn: () => appSecretsApi.listSecrets(projectId, env),
    enabled: !!env,
  })

  const currentEnv = envs.find((e) => e.name === env)

  const saveMutation = useMutation({
    mutationFn: () => {
      if (editing) {
        return appSecretsApi.updateSecret(editing.id, {
          key,
          value: value || undefined,
          comment: comment ? comment : null,
        })
      }
      return appSecretsApi.createSecret(projectId, {
        environment_id: currentEnv?.id,
        key,
        value,
        comment: comment ? comment : null,
      })
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['app-secrets', projectId] })
      qc.invalidateQueries({ queryKey: ['app-project', projectId] })
      toast.success(editing ? 'Секрет обновлён' : 'Секрет создан')
      setOpen(false)
    },
    onError: (err) => toast.error(err instanceof ApiError ? err.message : 'Ошибка сохранения'),
  })

  const removeMutation = useMutation({
    mutationFn: (id: string) => appSecretsApi.removeSecret(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['app-secrets', projectId] })
      qc.invalidateQueries({ queryKey: ['app-project', projectId] })
      toast.success('Секрет удалён')
      setToDelete(null)
    },
    onError: () => toast.error('Не удалось удалить'),
  })

  const openCreate = () => {
    setEditing(null)
    setKey('')
    setValue('')
    setComment('')
    setOpen(true)
  }
  const openEdit = (s: AppSecret) => {
    setEditing(s)
    setKey(s.key)
    setValue('')
    setComment('')
    setOpen(true)
  }

  if (projectQ.isLoading || envsQ.isLoading) return <PageLoader />

  const secrets = secretsQ.data ?? []

  return (
    <div>
      <Link to="/app-secrets" className="inline-flex items-center gap-1.5 text-[13px] text-ink-muted hover:text-accent mb-3">
        <ArrowLeft size={15} />
        Все проекты
      </Link>

      <PageHeader
        eyebrow="Секреты приложений"
        title={projectQ.data?.name ?? 'Проект'}
        description={projectQ.data?.description ?? 'Значения секретов зашифрованы и не показываются в списке.'}
        action={
          <button className="btn btn-primary" onClick={openCreate} disabled={!currentEnv}>
            <Plus size={16} />
            Добавить секрет
          </button>
        }
      />

      {/* Environment tabs */}
      <div className="flex items-center gap-1 mb-5 border-b border-border">
        {envs.map((e) => (
          <button
            key={e.id}
            onClick={() => setEnv(e.name)}
            className={
              'relative px-4 py-2.5 text-[14px] font-medium transition-colors ' +
              (e.name === env ? 'text-accent-text' : 'text-ink-muted hover:text-ink-strong')
            }
          >
            {e.name}
            {e.name === env && (
              <span className="absolute left-3 right-3 -bottom-px h-0.5 rounded bg-accent" />
            )}
          </button>
        ))}
      </div>

      {secretsQ.isLoading ? (
        <PageLoader />
      ) : secrets.length === 0 ? (
        <EmptyState
          icon={<KeySquare size={22} />}
          title="В этом окружении нет секретов"
          hint={`Добавьте первый секрет для окружения «${env}».`}
          action={
            <button className="btn btn-primary" onClick={openCreate} disabled={!currentEnv}>
              <Plus size={16} />
              Добавить секрет
            </button>
          }
        />
      ) : (
        <div className="card p-4 overflow-x-auto">
          <table className="data-table">
            <thead>
              <tr>
                <th>Ключ</th>
                <th>Значение</th>
                <th>Обновлён</th>
                <th className="text-right">Действия</th>
              </tr>
            </thead>
            <tbody>
              {secrets.map((s) => (
                <tr key={s.id}>
                  <td>
                    <span className="keychip">{s.key}</span>
                  </td>
                  <td>
                    <RevealValue
                      fetchValue={async () => (await appSecretsApi.revealSecret(s.id)).value}
                      copiedMessage="Значение скопировано"
                    />
                  </td>
                  <td className="text-ink-muted">{formatDate(s.updated_at)}</td>
                  <td>
                    <div className="flex items-center justify-end gap-1">
                      <button className="icon-btn w-8 h-8" onClick={() => openEdit(s)} title="Редактировать">
                        <Pencil size={15} />
                      </button>
                      <button
                        className="icon-btn w-8 h-8 hover:text-danger-icon"
                        onClick={() => setToDelete(s)}
                        title="Удалить"
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

      <Drawer
        open={open}
        onClose={() => setOpen(false)}
        title={editing ? 'Редактировать секрет' : 'Новый секрет'}
        subtitle={`Окружение: ${env}`}
        footer={
          <>
            <button className="btn" onClick={() => setOpen(false)}>
              Отмена
            </button>
            <button
              className="btn btn-primary"
              onClick={() => key.trim() && (editing || value) && saveMutation.mutate()}
              disabled={saveMutation.isPending || !key.trim() || (!editing && !value)}
            >
              {saveMutation.isPending ? <Spinner size={16} /> : 'Сохранить'}
            </button>
          </>
        }
      >
        <div className="flex flex-col gap-4">
          <div>
            <label className="field-label">Ключ</label>
            <input
              className="input font-mono"
              placeholder="DATABASE_URL"
              value={key}
              onChange={(e) => setKey(e.target.value)}
              autoFocus
            />
          </div>
          <div>
            <label className="field-label">
              Значение {editing && <span className="text-ink-faint font-normal">(пусто = без изменений)</span>}
            </label>
            <textarea
              className="textarea font-mono"
              rows={3}
              placeholder="postgres://..."
              value={value}
              onChange={(e) => setValue(e.target.value)}
            />
          </div>
          <div>
            <label className="field-label">Комментарий</label>
            <input className="input" value={comment} onChange={(e) => setComment(e.target.value)} />
          </div>
        </div>
      </Drawer>

      <ConfirmDialog
        open={!!toDelete}
        title="Удалить секрет?"
        message={
          <>
            Секрет <strong>{toDelete?.key}</strong> будет удалён.
          </>
        }
        busy={removeMutation.isPending}
        onCancel={() => setToDelete(null)}
        onConfirm={() => toDelete && removeMutation.mutate(toDelete.id)}
      />
    </div>
  )
}
