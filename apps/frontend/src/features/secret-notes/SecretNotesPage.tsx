import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { FileLock2, Pencil, Plus, Search, Trash2 } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader'
import { EmptyState } from '../../components/ui/EmptyState'
import { PageLoader } from '../../components/ui/Spinner'
import { ConfirmDialog } from '../../components/ui/ConfirmDialog'
import { useToast } from '../../components/ui/Toast'
import { SecretNoteEditorDrawer } from './SecretNoteEditorDrawer'
import { NOTE_TYPES, noteTypeMeta } from './noteTypes'
import { secretNotesApi } from '../../api/secretNotes'
import { formatDate } from '../../lib/formatDate'
import type { SecretNote } from '../../types'

export function SecretNotesPage() {
  const qc = useQueryClient()
  const toast = useToast()

  const [search, setSearch] = useState('')
  const [type, setType] = useState('')
  const [editorOpen, setEditorOpen] = useState(false)
  const [editing, setEditing] = useState<SecretNote | null>(null)
  const [toDelete, setToDelete] = useState<SecretNote | null>(null)

  const filters = { q: search, type }
  const listQ = useQuery({
    queryKey: ['secret-notes', filters],
    queryFn: () => secretNotesApi.list(filters),
  })

  const removeMutation = useMutation({
    mutationFn: (id: string) => secretNotesApi.remove(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['secret-notes'] })
      toast.success('Запись удалена')
      setToDelete(null)
    },
    onError: () => toast.error('Не удалось удалить'),
  })

  const openCreate = () => {
    setEditing(null)
    setEditorOpen(true)
  }
  const openEdit = (n: SecretNote) => {
    setEditing(n)
    setEditorOpen(true)
  }

  const notes = listQ.data ?? []

  return (
    <div>
      <PageHeader
        eyebrow="Секретные записи"
        title="Секретные записи"
        description="Произвольные секреты: название, тип и текст. Текст хранится в зашифрованном виде и открывается только по запросу."
        action={
          <button className="btn btn-primary" onClick={openCreate}>
            <Plus size={16} />
            Добавить
          </button>
        }
      />

      {/* Filters */}
      <div className="card p-3 mb-5 flex flex-wrap items-center gap-2">
        <div className="relative flex-1 min-w-[180px]">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-ink-faint" />
          <input
            className="input pl-9"
            placeholder="Поиск по названию"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <select className="select w-auto" value={type} onChange={(e) => setType(e.target.value)}>
          <option value="">Все типы</option>
          {NOTE_TYPES.map((t) => (
            <option key={t.type} value={t.type}>
              {t.label}
            </option>
          ))}
        </select>
      </div>

      {listQ.isLoading ? (
        <PageLoader />
      ) : notes.length === 0 ? (
        <EmptyState
          icon={<FileLock2 size={22} />}
          title="Записей нет"
          hint="Добавьте первую секретную запись — текст будет храниться в зашифрованном виде."
          action={
            <button className="btn btn-primary" onClick={openCreate}>
              <Plus size={16} />
              Добавить запись
            </button>
          }
        />
      ) : (
        <>
          {/* Desktop table */}
          <div className="card p-4 hidden md:block overflow-x-auto">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Запись</th>
                  <th>Тип</th>
                  <th>Обновлена</th>
                  <th className="text-right">Действия</th>
                </tr>
              </thead>
              <tbody>
                {notes.map((n) => {
                  const meta = noteTypeMeta(n.type)
                  const Icon = meta.icon
                  return (
                    <tr key={n.id}>
                      <td>
                        <button
                          type="button"
                          className="flex items-center gap-3 text-left group"
                          onClick={() => openEdit(n)}
                          title="Открыть запись"
                        >
                          <span className="grid place-items-center w-8 h-8 rounded-md bg-accent-50 text-accent shrink-0">
                            <Icon size={16} />
                          </span>
                          <span className="font-medium text-ink-strong group-hover:text-accent truncate">
                            {n.title}
                          </span>
                        </button>
                      </td>
                      <td>
                        <span className="badge badge-neutral">{meta.label}</span>
                      </td>
                      <td className="text-ink-muted">{formatDate(n.updated_at)}</td>
                      <td>
                        <div className="flex items-center justify-end gap-1">
                          <button className="icon-btn w-8 h-8" onClick={() => openEdit(n)} title="Редактировать">
                            <Pencil size={15} />
                          </button>
                          <button
                            className="icon-btn w-8 h-8 hover:text-danger-icon"
                            onClick={() => setToDelete(n)}
                            title="Удалить"
                          >
                            <Trash2 size={15} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>

          {/* Mobile cards */}
          <div className="grid gap-3 md:hidden">
            {notes.map((n) => {
              const meta = noteTypeMeta(n.type)
              const Icon = meta.icon
              return (
                <div key={n.id} className="card p-4">
                  <div className="flex items-center gap-3">
                    <span className="grid place-items-center w-9 h-9 rounded-md bg-accent-50 text-accent shrink-0">
                      <Icon size={18} />
                    </span>
                    <button type="button" className="min-w-0 flex-1 text-left" onClick={() => openEdit(n)}>
                      <div className="font-medium text-ink-strong truncate">{n.title}</div>
                      <div className="text-[12.5px] text-ink-muted truncate">{meta.label}</div>
                    </button>
                    <div className="flex gap-1">
                      <button className="icon-btn w-8 h-8" onClick={() => openEdit(n)}>
                        <Pencil size={15} />
                      </button>
                      <button className="icon-btn w-8 h-8 hover:text-danger-icon" onClick={() => setToDelete(n)}>
                        <Trash2 size={15} />
                      </button>
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        </>
      )}

      <SecretNoteEditorDrawer open={editorOpen} onClose={() => setEditorOpen(false)} note={editing} />

      <ConfirmDialog
        open={!!toDelete}
        title="Удалить запись?"
        message={
          <>
            Запись <strong>{toDelete?.title}</strong> будет удалена без возможности восстановления.
          </>
        }
        busy={removeMutation.isPending}
        onCancel={() => setToDelete(null)}
        onConfirm={() => toDelete && removeMutation.mutate(toDelete.id)}
      />
    </div>
  )
}
