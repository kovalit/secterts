import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Building2, Pencil, Plus, Trash2 } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader'
import { EmptyState } from '../../components/ui/EmptyState'
import { PageLoader, Spinner } from '../../components/ui/Spinner'
import { ConfirmDialog } from '../../components/ui/ConfirmDialog'
import { Drawer } from '../../components/ui/Drawer'
import { useToast } from '../../components/ui/Toast'
import { companiesApi } from '../../api/companies'
import { ApiError } from '../../api/client'
import { formatDate } from '../../lib/formatDate'
import type { Company } from '../../types'

export function CompaniesPage() {
  const qc = useQueryClient()
  const toast = useToast()
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Company | null>(null)
  const [name, setName] = useState('')
  const [toDelete, setToDelete] = useState<Company | null>(null)

  const listQ = useQuery({ queryKey: ['companies'], queryFn: companiesApi.list })

  const saveMutation = useMutation({
    mutationFn: () => (editing ? companiesApi.update(editing.id, name) : companiesApi.create(name)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['companies'] })
      toast.success(editing ? 'Компания обновлена' : 'Компания создана')
      setOpen(false)
    },
    onError: (err) => toast.error(err instanceof ApiError ? err.message : 'Ошибка сохранения'),
  })

  const removeMutation = useMutation({
    mutationFn: (id: string) => companiesApi.remove(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['companies'] })
      toast.success('Компания удалена')
      setToDelete(null)
    },
    onError: () => toast.error('Не удалось удалить'),
  })

  const openCreate = () => {
    setEditing(null)
    setName('')
    setOpen(true)
  }
  const openEdit = (c: Company) => {
    setEditing(c)
    setName(c.name)
    setOpen(true)
  }

  const companies = listQ.data ?? []

  return (
    <div>
      <PageHeader
        eyebrow="Пароли"
        title="Компании"
        description="Компании используются для коммерческих паролей и проектов."
        action={
          <button className="btn btn-primary" onClick={openCreate}>
            <Plus size={16} />
            Добавить компанию
          </button>
        }
      />

      {listQ.isLoading ? (
        <PageLoader />
      ) : companies.length === 0 ? (
        <EmptyState
          icon={<Building2 size={22} />}
          title="Компаний пока нет"
          hint="Создайте компанию, чтобы группировать коммерческие пароли."
          action={
            <button className="btn btn-primary" onClick={openCreate}>
              <Plus size={16} />
              Добавить компанию
            </button>
          }
        />
      ) : (
        <div className="card p-4 overflow-x-auto">
          <table className="data-table">
            <thead>
              <tr>
                <th>Компания</th>
                <th>Создана</th>
                <th className="text-right">Действия</th>
              </tr>
            </thead>
            <tbody>
              {companies.map((c) => (
                <tr key={c.id}>
                  <td>
                    <div className="flex items-center gap-3">
                      <span className="grid place-items-center w-8 h-8 rounded-md bg-accent-50 text-accent">
                        <Building2 size={16} />
                      </span>
                      <span className="font-medium text-ink-strong">{c.name}</span>
                    </div>
                  </td>
                  <td className="text-ink-muted">{formatDate(c.created_at)}</td>
                  <td>
                    <div className="flex items-center justify-end gap-1">
                      <button className="icon-btn w-8 h-8" onClick={() => openEdit(c)} title="Переименовать">
                        <Pencil size={15} />
                      </button>
                      <button
                        className="icon-btn w-8 h-8 hover:text-danger-icon"
                        onClick={() => setToDelete(c)}
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
        title={editing ? 'Переименовать компанию' : 'Новая компания'}
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
        <div>
          <label className="field-label">Название компании</label>
          <input
            className="input"
            placeholder="Например, Acme Inc."
            value={name}
            onChange={(e) => setName(e.target.value)}
            autoFocus
          />
        </div>
      </Drawer>

      <ConfirmDialog
        open={!!toDelete}
        title="Удалить компанию?"
        message={
          <>
            Компания <strong>{toDelete?.name}</strong> и связанные с ней коммерческие пароли будут удалены.
          </>
        }
        busy={removeMutation.isPending}
        onCancel={() => setToDelete(null)}
        onConfirm={() => toDelete && removeMutation.mutate(toDelete.id)}
      />
    </div>
  )
}
