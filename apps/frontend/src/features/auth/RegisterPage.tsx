import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Link, useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { AuthShell } from './AuthShell'
import { authApi } from '../../api/auth'
import { ApiError } from '../../api/client'
import { useToast } from '../../components/ui/Toast'
import { Spinner } from '../../components/ui/Spinner'

const schema = z
  .object({
    email: z.string().email('Введите корректный email'),
    password: z.string().min(8, 'Минимум 8 символов'),
    confirm: z.string(),
  })
  .refine((v) => v.password === v.confirm, {
    message: 'Пароли не совпадают',
    path: ['confirm'],
  })
type FormValues = z.infer<typeof schema>

export function RegisterPage() {
  const navigate = useNavigate()
  const toast = useToast()
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormValues>({ resolver: zodResolver(schema) })

  const mutation = useMutation({
    mutationFn: (v: FormValues) => authApi.register(v.email, v.password),
    onSuccess: () => {
      toast.success('Аккаунт создан. Войдите с помощью email-кода.')
      navigate('/login')
    },
    onError: (err) => {
      toast.error(err instanceof ApiError ? err.message : 'Не удалось зарегистрироваться')
    },
  })

  return (
    <AuthShell
      title="Регистрация"
      subtitle="Создайте аккаунт. Вход всегда защищён кодом на email."
      footer={
        <>
          Уже есть аккаунт?{' '}
          <Link to="/login" className="text-accent font-medium">
            Войти
          </Link>
        </>
      }
    >
      <form className="flex flex-col gap-4" onSubmit={handleSubmit((v) => mutation.mutate(v))}>
        <div>
          <label className="field-label" htmlFor="email">
            Email
          </label>
          <input id="email" type="email" autoComplete="email" className="input" {...register('email')} />
          {errors.email && <p className="field-error">{errors.email.message}</p>}
        </div>
        <div>
          <label className="field-label" htmlFor="password">
            Пароль
          </label>
          <input
            id="password"
            type="password"
            autoComplete="new-password"
            className="input"
            {...register('password')}
          />
          {errors.password && <p className="field-error">{errors.password.message}</p>}
        </div>
        <div>
          <label className="field-label" htmlFor="confirm">
            Повторите пароль
          </label>
          <input
            id="confirm"
            type="password"
            autoComplete="new-password"
            className="input"
            {...register('confirm')}
          />
          {errors.confirm && <p className="field-error">{errors.confirm.message}</p>}
        </div>
        <button className="btn btn-primary w-full" type="submit" disabled={mutation.isPending}>
          {mutation.isPending ? <Spinner size={16} /> : 'Создать аккаунт'}
        </button>
      </form>
    </AuthShell>
  )
}
