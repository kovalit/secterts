import { useMemo, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Copy, KeyRound, Pencil, Plus, Search, Trash2 } from 'lucide-react'
import { PageHeader } from '../../components/layout/PageHeader'
import { EmptyState } from '../../components/ui/EmptyState'
import { PageLoader } from '../../components/ui/Spinner'
import { ConfirmDialog } from '../../components/ui/ConfirmDialog'
import { EntryIcon } from '../../components/ui/GroupIcon'
import { RevealValue } from './RevealValue'
import { PasswordEditorDrawer } from './PasswordEditorDrawer'
import { passwordsApi } from '../../api/passwords'
import { companiesApi } from '../../api/companies'
import { copyToClipboard } from '../../lib/clipboard'
import { useToast } from '../../components/ui/Toast'
import type { PasswordEntry } from '../../types'

export function PasswordsPage() {
  const [params, setParams] = useSearchParams()
  const scope = params.get('scope') ?? ''
  const qc = useQueryClient()
  const toast = useToast()

  const [search, setSearch] = useState('')
  const [groupId, setGroupId] = useState('')
  const [companyId, setCompanyId] = useState('')
  const [editorOpen, setEditorOpen] = useState(false)
  const [editing, setEditing] = useState<PasswordEntry | null>(null)
  const [toDelete, setToDelete] = useState<PasswordEntry | null>(null)

  const groupsQ = useQuery({ queryKey: ['password-groups'], queryFn: passwordsApi.groups })
  const companiesQ = useQuery({ queryKey: ['companies'], queryFn: companiesApi.list })

  const filters = { scope, q: search, group_id: groupId, company_id: companyId }
  const listQ = useQuery({
    queryKey: ['passwords', filters],
    queryFn: () => passwordsApi.list(filters),
  })

  const groups = groupsQ.data ?? []
  const companies = companiesQ.data ?? []
  const groupLabel = useMemo(() => {
    const m = new Map(groups.map((g) => [g.id, g.label]))
    return (id: string) => m.get(id) ?? '—'
  }, [groups])
  const companyName = useMemo(() => {
    const m = new Map(companies.map((c) => [c.id, c.name]))
    return (id?: string | null) => (id ? m.get(id) ?? '—' : '—')
  }, [companies])

  const removeMutation = useMutation({
    mutationFn: (id: string) => passwordsApi.remove(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['passwords'] })
      toast.success('Запись удалена')
      setToDelete(null)
    },
    onError: () => toast.error('Не удалось удалить'),
  })

  const scopeTitle =
    scope === 'personal' ? 'Личные пароли' : scope === 'commercial' ? 'Коммерческие пароли' : 'Все пароли'

  const openCreate = () => {
    setEditing(null)
    setEditorOpen(true)
  }
  const openEdit = (e: PasswordEntry) => {
    setEditing(e)
    setEditorOpen(true)
  }

  const copyLogin = async (login?: string | null) => {
    if (!login) return
    const ok = await copyToClipboard(login)
    if (ok) toast.success('Логин скопирован')
  }

  const setScope = (next: string) => {
    const p = new URLSearchParams(params)
    if (next) p.set('scope', next)
    else p.delete('scope')
    setParams(p)
  }

  const entries = listQ.data ?? []

  return (
    <div>
      <PageHeader
        eyebrow="Пароли"
        title={scopeTitle}
        description="Пароли не показываются в списке — раскрывайте или копируйте их только по кнопке."
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
            placeholder="Поиск по названию, логину, домену"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <select className="select w-auto" value={scope} onChange={(e) => setScope(e.target.value)}>
          <option value="">Все типы</option>
          <option value="personal">Личные</option>
          <option value="commercial">Коммерческие</option>
        </select>
        <select className="select w-auto" value={groupId} onChange={(e) => setGroupId(e.target.value)}>
          <option value="">Все группы</option>
          {groups.map((g) => (
            <option key={g.id} value={g.id}>
              {g.label}
            </option>
          ))}
        </select>
        {companies.length > 0 && (
          <select className="select w-auto" value={companyId} onChange={(e) => setCompanyId(e.target.value)}>
            <option value="">Все компании</option>
            {companies.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
        )}
      </div>

      {listQ.isLoading ? (
        <PageLoader />
      ) : entries.length === 0 ? (
        <EmptyState
          icon={<KeyRound size={22} />}
          title="Записей нет"
          hint="Добавьте первую запись пароля — она будет храниться в зашифрованном виде."
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
                  <th>Логин</th>
                  <th>Группа</th>
                  <th>Тип</th>
                  <th>Пароль</th>
                  <th className="text-right">Действия</th>
                </tr>
              </thead>
              <tbody>
                {entries.map((e) => (
                  <tr key={e.id}>
                    <td>
                      <div className="flex items-center gap-3">
                        <EntryIcon
                          iconUrl={e.custom_icon}
                          faviconUrl={e.favicon_url}
                          groupIcon={groups.find((g) => g.id === e.group_id)?.icon}
                        />
                        <div className="min-w-0">
                          <button
                            type="button"
                            className="block font-medium text-ink-strong truncate text-left hover:text-accent transition-colors"
                            onClick={() => openEdit(e)}
                            title="Открыть запись"
                          >
                            {e.title}
                          </button>
                          {e.domain && <div className="text-[12.5px] text-ink-muted truncate">{e.domain}</div>}
                        </div>
                      </div>
                    </td>
                    <td className="text-ink">
                      {e.login ? (
                        <button
                          className="inline-flex items-center gap-1.5 hover:text-accent"
                          onClick={() => copyLogin(e.login)}
                          title="Скопировать логин"
                        >
                          <span className="truncate max-w-[180px]">{e.login}</span>
                          <Copy size={13} className="text-ink-faint" />
                        </button>
                      ) : (
                        <span className="text-ink-faint">—</span>
                      )}
                    </td>
                    <td>
                      <span className="badge badge-neutral">{groupLabel(e.group_id)}</span>
                    </td>
                    <td>
                      {e.scope === 'commercial' ? (
                        <span className="badge badge-accent">{companyName(e.company_id)}</span>
                      ) : (
                        <span className="badge badge-neutral">Личное</span>
                      )}
                    </td>
                    <td>
                      <RevealValue
                        fetchValue={async () => (await passwordsApi.reveal(e.id)).password}
                        copiedMessage="Пароль скопирован"
                      />
                    </td>
                    <td>
                      <div className="flex items-center justify-end gap-1">
                        <button className="icon-btn w-8 h-8" onClick={() => openEdit(e)} title="Редактировать">
                          <Pencil size={15} />
                        </button>
                        <button
                          className="icon-btn w-8 h-8 hover:text-danger-icon"
                          onClick={() => setToDelete(e)}
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

          {/* Mobile cards */}
          <div className="grid gap-3 md:hidden">
            {entries.map((e) => (
              <div key={e.id} className="card p-4">
                <div className="flex items-center gap-3">
                  <EntryIcon
                    iconUrl={e.custom_icon}
                    faviconUrl={e.favicon_url}
                    groupIcon={groups.find((g) => g.id === e.group_id)?.icon}
                  />
                  <button
                    type="button"
                    className="min-w-0 flex-1 text-left"
                    onClick={() => openEdit(e)}
                  >
                    <div className="font-medium text-ink-strong truncate">{e.title}</div>
                    <div className="text-[12.5px] text-ink-muted truncate">{e.login ?? e.domain ?? '—'}</div>
                  </button>
                  <div className="flex gap-1">
                    <button className="icon-btn w-8 h-8" onClick={() => openEdit(e)}>
                      <Pencil size={15} />
                    </button>
                    <button className="icon-btn w-8 h-8 hover:text-danger-icon" onClick={() => setToDelete(e)}>
                      <Trash2 size={15} />
                    </button>
                  </div>
                </div>
                <div className="flex items-center justify-between mt-3 pt-3 border-t border-border">
                  <div className="flex gap-2">
                    <span className="badge badge-neutral">{groupLabel(e.group_id)}</span>
                    {e.scope === 'commercial' && <span className="badge badge-accent">{companyName(e.company_id)}</span>}
                  </div>
                  <RevealValue
                    fetchValue={async () => (await passwordsApi.reveal(e.id)).password}
                    copiedMessage="Пароль скопирован"
                  />
                </div>
              </div>
            ))}
          </div>
        </>
      )}

      <PasswordEditorDrawer
        open={editorOpen}
        onClose={() => setEditorOpen(false)}
        entry={editing}
        groups={groups}
        companies={companies}
      />

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
