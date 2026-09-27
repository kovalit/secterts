import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Drawer } from '../../components/ui/Drawer'
import { Spinner } from '../../components/ui/Spinner'
import { useToast } from '../../components/ui/Toast'
import { passwordsApi, type PasswordWrite } from '../../api/passwords'
import { ApiError } from '../../api/client'
import type { Company, PasswordEntry, PasswordGroup } from '../../types'

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
    } else {
      reset({ title: '', scope: 'personal', group_id: groups[0]?.id ?? '', password: '' })
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

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title={isEdit ? 'Редактировать запись' : 'Новая запись'}
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
          <input className="input" placeholder="https://github.com" {...register('site_url')} />
        </div>

        <div>
          <label className="field-label">Логин</label>
          <input className="input" placeholder="user@example.com" autoComplete="off" {...register('login')} />
        </div>

        <div>
          <label className="field-label">
            Пароль {isEdit && <span className="text-ink-faint font-normal">(оставьте пустым, чтобы не менять)</span>}
          </label>
          <input className="input" type="password" autoComplete="new-password" {...register('password')} />
        </div>

        <div>
          <label className="field-label">Комментарий</label>
          <textarea className="textarea" rows={3} placeholder="Хранится в зашифрованном виде" {...register('comment')} />
        </div>
      </form>
    </Drawer>
  )
}
