import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Boxes, Pencil, Plus, Trash2 } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader'
import { EmptyState } from '../../components/ui/EmptyState'
import { PageLoader, Spinner } from '../../components/ui/Spinner'
import { ConfirmDialog } from '../../components/ui/ConfirmDialog'
import { Drawer } from '../../components/ui/Drawer'
import { useToast } from '../../components/ui/Toast'
import { appSecretsApi } from '../../api/appSecrets'
import { companiesApi } from '../../api/companies'
import { ApiError } from '../../api/client'
import { formatDate } from '../../lib/formatDate'
import type { AppProject } from '../../types'

export function ProjectsPage() {
  const qc = useQueryClient()
  const toast = useToast()
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<AppProject | null>(null)
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [companyId, setCompanyId] = useState('')
  const [toDelete, setToDelete] = useState<AppProject | null>(null)

  const listQ = useQuery({ queryKey: ['app-projects'], queryFn: appSecretsApi.listProjects })
  const companiesQ = useQuery({ queryKey: ['companies'], queryFn: companiesApi.list })
  const companies = companiesQ.data ?? []

  const saveMutation = useMutation({
    mutationFn: () => {
      const body = { name, description: description || null, company_id: companyId || null }
      return editing ? appSecretsApi.updateProject(editing.id, body) : appSecretsApi.createProject(body)
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['app-projects'] })
      toast.success(editing ? 'Проект обновлён' : 'Проект создан')
      setOpen(false)
    },
    onError: (err) => toast.error(err instanceof ApiError ? err.message : 'Ошибка сохранения'),
  })

  const removeMutation = useMutation({
    mutationFn: (id: string) => appSecretsApi.removeProject(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['app-projects'] })
      toast.success('Проект удалён')
      setToDelete(null)
    },
    onError: () => toast.error('Не удалось удалить'),
  })

  const openCreate = () => {
    setEditing(null)
    setName('')
    setDescription('')
    setCompanyId('')
    setOpen(true)
  }
  const openEdit = (p: AppProject) => {
    setEditing(p)
    setName(p.name)
    setDescription(p.description ?? '')
    setCompanyId(p.company_id ?? '')
    setOpen(true)
  }

  const projects = listQ.data ?? []

  return (
    <div>
      <PageHeader
        eyebrow="Секреты приложений"
        title="Проекты"
        description="Каждый проект содержит окружения (dev / staging / prod) и зашифрованные секреты."
        action={
          <button className="btn btn-primary" onClick={openCreate}>
            <Plus size={16} />
            Новый проект
          </button>
        }
      />

      {listQ.isLoading ? (
        <PageLoader />
      ) : projects.length === 0 ? (
        <EmptyState
          icon={<Boxes size={22} />}
          title="Проектов пока нет"
          hint="Создайте проект — окружения dev, staging и prod добавятся автоматически."
          action={
            <button className="btn btn-primary" onClick={openCreate}>
              <Plus size={16} />
              Новый проект
            </button>
          }
        />
      ) : (
        <div className="card p-4 overflow-x-auto">
          <table className="data-table">
            <thead>
              <tr>
                <th>Проект</th>
                <th>Окружения</th>
                <th>Секреты</th>
                <th>Обновлён</th>
                <th className="text-right">Действия</th>
              </tr>
            </thead>
            <tbody>
              {projects.map((p) => (
                <tr key={p.id}>
                  <td>
                    <Link to={`/app-secrets/${p.id}`} className="flex items-center gap-3 group">
                      <span className="grid place-items-center w-8 h-8 rounded-md bg-accent-50 text-accent">
                        <Boxes size={16} />
                      </span>
                      <div className="min-w-0">
                        <div className="font-medium text-ink-strong group-hover:text-accent truncate">
                          {p.name}
                        </div>
                        {p.description && (
                          <div className="text-[12.5px] text-ink-muted truncate">{p.description}</div>
                        )}
                      </div>
                    </Link>
                  </td>
                  <td className="text-ink-muted">{p.env_count}</td>
                  <td className="text-ink-muted">{p.secret_count}</td>
                  <td className="text-ink-muted">{formatDate(p.updated_at)}</td>
                  <td>
                    <div className="flex items-center justify-end gap-1">
                      <button className="icon-btn w-8 h-8" onClick={() => openEdit(p)} title="Редактировать">
                        <Pencil size={15} />
                      </button>
                      <button
                        className="icon-btn w-8 h-8 hover:text-danger-icon"
                        onClick={() => setToDelete(p)}
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
        title={editing ? 'Редактировать проект' : 'Новый проект'}
        footer={
          <>
            <button className="btn" onClick={() => setOpen(false)}>
              Отмена
            </button>
            <button
              className="btn btn-primary"
              onClick={() => name.trim() && saveMutation.mutate()}
              disabled={saveMutation.isPending || !name.trim()}
            >
              {saveMutation.isPending ? <Spinner size={16} /> : 'Сохранить'}
            </button>
          </>
        }
      >
        <div className="flex flex-col gap-4">
          <div>
            <label className="field-label">Название</label>
            <input className="input" value={name} onChange={(e) => setName(e.target.value)} autoFocus />
          </div>
          <div>
            <label className="field-label">Описание</label>
            <textarea className="textarea" rows={3} value={description} onChange={(e) => setDescription(e.target.value)} />
          </div>
          {companies.length > 0 && (
            <div>
              <label className="field-label">Компания (необязательно)</label>
              <select className="select" value={companyId} onChange={(e) => setCompanyId(e.target.value)}>
                <option value="">— личный проект —</option>
                {companies.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </div>
          )}
        </div>
      </Drawer>

      <ConfirmDialog
        open={!!toDelete}
        title="Удалить проект?"
        message={
          <>
            Проект <strong>{toDelete?.name}</strong>, его окружения и все секреты будут удалены.
          </>
        }
        busy={removeMutation.isPending}
        onCancel={() => setToDelete(null)}
        onConfirm={() => toDelete && removeMutation.mutate(toDelete.id)}
      />
    </div>
  )
}
