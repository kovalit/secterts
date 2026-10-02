import { useEffect, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Copy } from 'lucide-react'
import { Drawer } from '../../components/ui/Drawer'
import { Spinner } from '../../components/ui/Spinner'
import { useToast } from '../../components/ui/Toast'
import { secretNotesApi } from '../../api/secretNotes'
import { ApiError } from '../../api/client'
import { copyToClipboard } from '../../lib/clipboard'
import { NOTE_TYPES, noteTypeMeta } from './noteTypes'
import type { SecretNote, SecretNoteType } from '../../types'

export function SecretNoteEditorDrawer({
  open,
  onClose,
  note,
}: {
  open: boolean
  onClose: () => void
  note: SecretNote | null
}) {
  const qc = useQueryClient()
  const toast = useToast()
  const isEdit = !!note

  const [title, setTitle] = useState('')
  const [type, setType] = useState<SecretNoteType>('generic')
  const [text, setText] = useState('')
  const [loadingText, setLoadingText] = useState(false)

  // Load the note's encrypted text when opening an existing entry; reset the
  // form when opening a new one. The reveal is audited on the backend.
  useEffect(() => {
    if (!open) return
    setTitle(note?.title ?? '')
    setType(note?.type ?? 'generic')
    setText('')
    if (note) {
      setLoadingText(true)
      secretNotesApi
        .reveal(note.id)
        .then((r) => setText(r.text))
        .catch(() => toast.error('Не удалось расшифровать текст'))
        .finally(() => setLoadingText(false))
    }
  }, [open, note, toast])

  const meta = noteTypeMeta(type)
  const TypeIcon = meta.icon

  const mutation = useMutation({
    mutationFn: () => {
      const body = { title: title.trim(), type, text: text ? text : undefined }
      return isEdit ? secretNotesApi.update(note!.id, body) : secretNotesApi.create(body)
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['secret-notes'] })
      toast.success(isEdit ? 'Запись обновлена' : 'Запись создана')
      onClose()
    },
    onError: (err) => toast.error(err instanceof ApiError ? err.message : 'Ошибка сохранения'),
  })

  const canSave = title.trim().length > 0 && (isEdit || text.length > 0) && !loadingText

  const onCopyText = async () => {
    if (!text) return
    const ok = await copyToClipboard(text)
    if (ok) toast.success('Текст скопирован')
    else toast.error('Не удалось скопировать')
  }

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title={isEdit ? 'Секретная запись' : 'Новая запись'}
      subtitle={isEdit ? note?.title : 'Текст хранится в зашифрованном виде'}
      footer={
        <>
          <button className="btn" onClick={onClose} type="button">
            Отмена
          </button>
          <button
            className="btn btn-primary"
            onClick={() => canSave && mutation.mutate()}
            disabled={mutation.isPending || !canSave}
          >
            {mutation.isPending ? <Spinner size={16} /> : 'Сохранить'}
          </button>
        </>
      }
    >
      <div className="flex flex-col gap-4">
        <div>
          <label className="field-label">Название</label>
          <input
            className="input"
            placeholder="Например, Продакшн API-ключ"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            autoFocus
          />
        </div>

        <div>
          <label className="field-label">Тип</label>
          <div className="flex items-center gap-2">
            <span className="grid place-items-center w-9 h-9 rounded-md bg-accent-50 text-accent shrink-0">
              <TypeIcon size={18} />
            </span>
            <select
              className="select flex-1"
              value={type}
              onChange={(e) => setType(e.target.value as SecretNoteType)}
            >
              {NOTE_TYPES.map((t) => (
                <option key={t.type} value={t.type}>
                  {t.label}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div>
          <label className="field-label flex items-center justify-between">
            <span>
              Текст{' '}
              {isEdit && <span className="text-ink-faint font-normal">(расшифровано)</span>}
            </span>
            {isEdit && text && (
              <button
                type="button"
                className="inline-flex items-center gap-1 text-[12.5px] text-ink-muted hover:text-accent"
                onClick={onCopyText}
              >
                <Copy size={13} />
                Скопировать
              </button>
            )}
          </label>
          {loadingText ? (
            <div className="flex items-center gap-2 text-ink-muted text-[13px] py-3">
              <Spinner size={16} />
              Расшифровка…
            </div>
          ) : (
            <textarea
              className="textarea font-mono"
              rows={8}
              placeholder="Секретный текст — хранится в зашифрованном виде"
              value={text}
              onChange={(e) => setText(e.target.value)}
            />
          )}
        </div>
      </div>
    </Drawer>
  )
}
