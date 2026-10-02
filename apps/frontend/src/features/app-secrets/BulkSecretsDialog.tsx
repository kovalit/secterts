import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { ClipboardPaste, X } from 'lucide-react'
import { Spinner } from '../../components/ui/Spinner'
import { useToast } from '../../components/ui/Toast'
import { appSecretsApi } from '../../api/appSecrets'
import { ApiError } from '../../api/client'
import { parseEnv } from '../../lib/parseEnv'
import type { AppProject } from '../../types'

const ENV_OPTIONS = ['dev', 'staging', 'prod']

export function BulkSecretsDialog({
  open,
  onClose,
  initialText,
  projects,
  fixedProject,
  defaultEnv,
  onDone,
}: {
  open: boolean
  onClose: () => void
  initialText: string
  projects: AppProject[]
  fixedProject?: { id: string; name: string } | null
  defaultEnv?: string
  onDone?: () => void
}) {
  const qc = useQueryClient()
  const toast = useToast()

  const [text, setText] = useState('')
  const [mode, setMode] = useState<'existing' | 'new'>('existing')
  const [projectId, setProjectId] = useState('')
  const [newProjectName, setNewProjectName] = useState('')
  const [env, setEnv] = useState('dev')

  useEffect(() => {
    if (!open) return
    setText(initialText)
    setEnv(defaultEnv && ENV_OPTIONS.includes(defaultEnv) ? defaultEnv : 'dev')
    if (fixedProject) {
      setMode('existing')
      setProjectId(fixedProject.id)
    } else {
      setMode(projects.length > 0 ? 'existing' : 'new')
      setProjectId(projects[0]?.id ?? '')
      setNewProjectName('')
    }
  }, [open, initialText, defaultEnv, fixedProject, projects])

  const items = useMemo(() => parseEnv(text), [text])

  const mutation = useMutation({
    mutationFn: async () => {
      let targetId = fixedProject?.id ?? projectId
      let createdName: string | null = null
      if (!fixedProject && mode === 'new') {
        const created = await appSecretsApi.createProject({ name: newProjectName.trim() })
        targetId = created.id
        createdName = created.name
      }
      const result = await appSecretsApi.bulkCreateSecrets(targetId, {
        environment: env,
        items: items.map((i) => ({ key: i.key, value: i.value })),
      })
      return { count: result.length, projectId: targetId, createdName }
    },
    onSuccess: ({ count, projectId: pid, createdName }) => {
      qc.invalidateQueries({ queryKey: ['app-projects'] })
      qc.invalidateQueries({ queryKey: ['app-project', pid] })
      qc.invalidateQueries({ queryKey: ['app-secrets', pid] })
      toast.success(
        createdName
          ? `Проект «${createdName}» создан, добавлено секретов: ${count}`
          : `Добавлено секретов: ${count}`,
      )
      onDone?.()
      onClose()
    },
    onError: (err) => toast.error(err instanceof ApiError ? err.message : 'Не удалось создать секреты'),
  })

  if (!open) return null

  const projectReady = !!fixedProject || (mode === 'existing' ? !!projectId : !!newProjectName.trim())
  const canSubmit = items.length > 0 && projectReady && !mutation.isPending

  return (
    <div className="fixed inset-0 z-[60] grid place-items-center px-4">
      <div className="absolute inset-0 bg-ink-strong/30" onClick={onClose} aria-hidden="true" />
      <div className="card relative w-full max-w-lg p-5 shadow-pop max-h-[90vh] overflow-y-auto">
        <div className="flex items-start gap-3">
          <span className="grid place-items-center w-10 h-10 rounded-xl bg-accent-50 text-accent shrink-0">
            <ClipboardPaste size={20} />
          </span>
          <div className="min-w-0 flex-1">
            <h3 className="text-[16px] font-semibold text-ink-strong">Создать секреты списком</h3>
            <p className="mt-0.5 text-[13px] text-ink-muted">
              Вставьте список в формате <code className="keychip">KEY=VALUE</code> — по одной паре в строке.
            </p>
          </div>
          <button className="icon-btn" onClick={onClose} aria-label="Закрыть">
            <X size={18} />
          </button>
        </div>

        <div className="mt-4 flex flex-col gap-4">
          {/* Project */}
          {fixedProject ? (
            <div>
              <label className="field-label">Проект</label>
              <input className="input" value={fixedProject.name} disabled />
            </div>
          ) : (
            <div>
              <label className="field-label">Проект</label>
              <div className="flex items-center gap-2 mb-2">
                <button
                  type="button"
                  className={'btn btn-sm ' + (mode === 'existing' ? 'btn-primary' : '')}
                  onClick={() => setMode('existing')}
                  disabled={projects.length === 0}
                >
                  Существующий
                </button>
                <button
                  type="button"
                  className={'btn btn-sm ' + (mode === 'new' ? 'btn-primary' : '')}
                  onClick={() => setMode('new')}
                >
                  Новый проект
                </button>
              </div>
              {mode === 'existing' ? (
                <select className="select" value={projectId} onChange={(e) => setProjectId(e.target.value)}>
                  {projects.length === 0 && <option value="">— нет проектов —</option>}
                  {projects.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name}
                    </option>
                  ))}
                </select>
              ) : (
                <input
                  className="input"
                  placeholder="Название нового проекта"
                  value={newProjectName}
                  onChange={(e) => setNewProjectName(e.target.value)}
                  autoFocus
                />
              )}
            </div>
          )}

          {/* Environment */}
          <div>
            <label className="field-label">Окружение</label>
            <select className="select" value={env} onChange={(e) => setEnv(e.target.value)}>
              {ENV_OPTIONS.map((e) => (
                <option key={e} value={e}>
                  {e}
                </option>
              ))}
            </select>
          </div>

          {/* Secrets list */}
          <div>
            <label className="field-label flex items-center justify-between">
              <span>Секреты</span>
              <span className="text-ink-faint font-normal">
                {items.length > 0 ? `распознано: ${items.length}` : 'нет пар KEY=VALUE'}
              </span>
            </label>
            <textarea
              className="textarea font-mono"
              rows={8}
              placeholder={'API_URL=https://api.example.com\nAPI_TOKEN=xxxxx'}
              value={text}
              onChange={(e) => setText(e.target.value)}
            />
            {items.length > 0 && (
              <div className="mt-2 flex flex-wrap gap-1.5">
                {items.slice(0, 12).map((i) => (
                  <span key={i.key} className="keychip">
                    {i.key}
                  </span>
                ))}
                {items.length > 12 && <span className="text-[12px] text-ink-faint">+{items.length - 12}</span>}
              </div>
            )}
            <p className="mt-2 text-[12px] text-ink-faint">
              Существующие ключи в выбранном окружении будут перезаписаны.
            </p>
          </div>
        </div>

        <div className="flex justify-end gap-2 mt-5">
          <button className="btn" onClick={onClose} disabled={mutation.isPending}>
            Отмена
          </button>
          <button className="btn btn-primary" onClick={() => canSubmit && mutation.mutate()} disabled={!canSubmit}>
            {mutation.isPending ? <Spinner size={16} /> : `Создать (${items.length})`}
          </button>
        </div>
      </div>
    </div>
  )
}
