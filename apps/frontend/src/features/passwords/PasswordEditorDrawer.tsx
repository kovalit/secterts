import { useEffect, useRef, useState, type ChangeEvent } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Copy, ExternalLink, Upload, X } from 'lucide-react'
import { Drawer } from '../../components/ui/Drawer'
import { Spinner } from '../../components/ui/Spinner'
import { useToast } from '../../components/ui/Toast'
import { EntryIcon } from '../../components/ui/GroupIcon'
import { RevealValue } from './RevealValue'
import { passwordsApi, type PasswordWrite } from '../../api/passwords'
import { ApiError } from '../../api/client'
import { copyToClipboard } from '../../lib/clipboard'
import { domainFromUrl } from '../../lib/domain'
import type { Company, PasswordEntry, PasswordGroup } from '../../types'

type IconSource = 'favicon' | 'group' | 'custom'

// Max size for an uploaded custom icon; kept small since it is stored inline.
const MAX_ICON_BYTES = 128 * 1024

const schema = z
  .object({
    title: z.string().min(1, 'Введите название'),
    scope: z.enum(['personal', 'commercial']),
    company_id: z.string().optional(),
    group_id: z.string().min(1, 'Выберите группу'),
    site_url: z.string().optional(),
    login: z.string().optional(),
    password: z.string().optional(),
    comment: z.string().optional(),
  })
  .refine((v) => v.scope !== 'commercial' || !!v.company_id, {
    message: 'Для коммерческой записи выберите компанию',
    path: ['company_id'],
  })
type FormValues = z.infer<typeof schema>

export function PasswordEditorDrawer({
  open,
  onClose,
  entry,
  groups,
  companies,
}: {
  open: boolean
  onClose: () => void
  entry: PasswordEntry | null
  groups: PasswordGroup[]
  companies: Company[]
}) {
  const qc = useQueryClient()
  const toast = useToast()
  const isEdit = !!entry

  const [iconSource, setIconSource] = useState<IconSource>('favicon')
  const [customIcon, setCustomIcon] = useState<string | null>(null)
  const fileRef = useRef<HTMLInputElement>(null)

  const {
    register,
    handleSubmit,
    reset,
    watch,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { scope: 'personal', group_id: groups[0]?.id ?? '' },
  })

  const scope = watch('scope')
  const siteUrl = watch('site_url')
  const login = watch('login')
  const groupId = watch('group_id')

  const groupIcon = groups.find((g) => g.id === groupId)?.icon
  const previewDomain = domainFromUrl(siteUrl)

  // Reset the form each time the drawer opens; clears password from state on close.
  useEffect(() => {
    if (open) {
      reset({
        title: entry?.title ?? '',
        scope: entry?.scope ?? 'personal',
        company_id: entry?.company_id ?? '',
        group_id: entry?.group_id ?? groups[0]?.id ?? '',
        site_url: entry?.site_url ?? '',
        login: entry?.login ?? '',
        password: '',
        comment: '',
      })
      setIconSource((entry?.icon_source as IconSource) ?? 'favicon')
      setCustomIcon(entry?.custom_icon ?? null)
    } else {
      reset({ title: '', scope: 'personal', group_id: groups[0]?.id ?? '', password: '' })
      setIconSource('favicon')
      setCustomIcon(null)
    }
  }, [open, entry, groups, reset])

  const mutation = useMutation({
    mutationFn: (v: FormValues) => {
      const body: PasswordWrite = {
        scope: v.scope,
        company_id: v.scope === 'commercial' ? v.company_id || null : null,
        group_id: v.group_id,
        title: v.title,
        site_url: v.site_url || null,
        login: v.login || null,
        comment: v.comment ? v.comment : null,
        icon_source: iconSource,
        custom_icon: iconSource === 'custom' ? customIcon : null,
      }
      if (v.password) body.password = v.password
      return isEdit ? passwordsApi.update(entry!.id, body) : passwordsApi.create(body)
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['passwords'] })
      toast.success(isEdit ? 'Запись обновлена' : 'Запись создана')
      onClose()
    },
    onError: (err) => toast.error(err instanceof ApiError ? err.message : 'Ошибка сохранения'),
  })

  const onCopyLogin = async () => {
    if (!login) return
    const ok = await copyToClipboard(login)
    if (ok) toast.success('Логин скопирован')
    else toast.error('Не удалось скопировать')
  }

  const openSite = () => {
    if (!siteUrl) return
    const href = siteUrl.includes('://') ? siteUrl : `https://${siteUrl}`
    window.open(href, '_blank', 'noopener,noreferrer')
  }

  const onPickFile = (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    e.target.value = '' // allow re-picking the same file
    if (!file) return
    if (!file.type.startsWith('image/')) {
      toast.error('Выберите файл изображения')
      return
    }
    if (file.size > MAX_ICON_BYTES) {
      toast.error('Иконка слишком большая (макс. 128 КБ)')
      return
    }
    const reader = new FileReader()
    reader.onload = () => {
      setCustomIcon(reader.result as string)
      setIconSource('custom')
    }
    reader.onerror = () => toast.error('Не удалось прочитать файл')
    reader.readAsDataURL(file)
  }

  // Live preview: custom icon, favicon (via a preview URL), or the group icon.
  const previewIconUrl =
    iconSource === 'custom'
      ? customIcon
      : iconSource === 'favicon' && previewDomain
        ? `https://www.google.com/s2/favicons?domain=${previewDomain}&sz=64`
        : null

  const iconBtn = (src: IconSource, label: string) => (
    <button
      type="button"
      onClick={() => setIconSource(src)}
      className={
        'btn btn-sm ' + (iconSource === src ? 'btn-primary' : '')
      }
    >
      {label}
    </button>
  )

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title={isEdit ? 'Запись пароля' : 'Новая запись'}
      subtitle={isEdit ? entry?.title : 'Пароль хранится в зашифрованном виде'}
      footer={
        <>
          <button className="btn" onClick={onClose} type="button">
            Отмена
          </button>
          <button
            className="btn btn-primary"
            type="submit"
            form="password-form"
            disabled={mutation.isPending}
          >
            {mutation.isPending ? <Spinner size={16} /> : 'Сохранить'}
          </button>
        </>
      }
    >
      <form id="password-form" className="flex flex-col gap-4" onSubmit={handleSubmit((v) => mutation.mutate(v))}>
        {/* Icon */}
        <div>
          <label className="field-label">Иконка</label>
          <div className="flex items-center gap-3">
            <EntryIcon iconUrl={previewIconUrl} groupIcon={groupIcon} size={28} />
            <div className="flex flex-wrap items-center gap-1.5">
              {iconBtn('favicon', 'Иконка сайта')}
              <button type="button" className="btn btn-sm" onClick={() => fileRef.current?.click()}>
                <Upload size={14} />
                Своя
              </button>
              {iconBtn('group', 'Без иконки')}
              {iconSource === 'custom' && customIcon && (
                <button
                  type="button"
                  className="icon-btn w-8 h-8"
                  title="Убрать свою иконку"
                  onClick={() => {
                    setCustomIcon(null)
                    setIconSource('favicon')
                  }}
                >
                  <X size={15} />
                </button>
              )}
            </div>
            <input ref={fileRef} type="file" accept="image/*" className="hidden" onChange={onPickFile} />
          </div>
          {iconSource === 'favicon' && !previewDomain && (
            <p className="mt-1 text-[12px] text-ink-faint">
              Укажите адрес сайта, чтобы получить иконку автоматически.
            </p>
          )}
        </div>

        <div>
          <label className="field-label">Название</label>
          <input className="input" placeholder="Например, GitHub" {...register('title')} />
          {errors.title && <p className="field-error">{errors.title.message}</p>}
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="field-label">Тип</label>
            <select className="select" {...register('scope')}>
              <option value="personal">Личное</option>
              <option value="commercial">Коммерческое</option>
            </select>
          </div>
          <div>
            <label className="field-label">Группа</label>
            <select className="select" {...register('group_id')}>
              {groups.map((g) => (
                <option key={g.id} value={g.id}>
                  {g.label}
                </option>
              ))}
            </select>
            {errors.group_id && <p className="field-error">{errors.group_id.message}</p>}
          </div>
        </div>

        {scope === 'commercial' && (
          <div>
            <label className="field-label">Компания</label>
            <select className="select" {...register('company_id')}>
              <option value="">— выберите компанию —</option>
              {companies.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
            {errors.company_id && <p className="field-error">{errors.company_id.message}</p>}
          </div>
        )}

        <div>
          <label className="field-label">Адрес сайта</label>
          <div className="flex items-center gap-2">
            <input className="input flex-1" placeholder="https://github.com" {...register('site_url')} />
            <button
              type="button"
              className="icon-btn w-9 h-9 shrink-0"
              onClick={openSite}
              disabled={!siteUrl}
              title="Открыть сайт в новой вкладке"
            >
              <ExternalLink size={16} />
            </button>
          </div>
        </div>

        <div>
          <label className="field-label">Логин</label>
          <div className="flex items-center gap-2">
            <input className="input flex-1" placeholder="user@example.com" autoComplete="off" {...register('login')} />
            <button
              type="button"
              className="icon-btn w-9 h-9 shrink-0"
              onClick={onCopyLogin}
              disabled={!login}
              title="Скопировать логин"
            >
              <Copy size={16} />
            </button>
          </div>
        </div>

        <div>
          <label className="field-label">
            Пароль {isEdit && <span className="text-ink-faint font-normal">(оставьте пустым, чтобы не менять)</span>}
          </label>
          <input className="input" type="password" autoComplete="new-password" {...register('password')} />
          {isEdit && entry?.has_password && (
            <div className="mt-2 flex items-center gap-2">
              <span className="text-[12.5px] text-ink-muted">Текущий пароль:</span>
              <RevealValue
                fetchValue={async () => (await passwordsApi.reveal(entry.id)).password}
                copiedMessage="Пароль скопирован"
              />
            </div>
          )}
        </div>

        <div>
          <label className="field-label">Комментарий</label>
          <textarea className="textarea" rows={3} placeholder="Хранится в зашифрованном виде" {...register('comment')} />
        </div>
      </form>
    </Drawer>
  )
}
