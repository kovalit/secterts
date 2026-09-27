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

const schema = z.object({
  email: z.string().email('Введите корректный email'),
  password: z.string().min(1, 'Введите пароль'),
})
type FormValues = z.infer<typeof schema>

export function LoginPage() {
  const navigate = useNavigate()
  const toast = useToast()
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormValues>({ resolver: zodResolver(schema) })

  const mutation = useMutation({
    mutationFn: (v: FormValues) => authApi.login(v.email, v.password),
    onSuccess: (res, vars) => {
      navigate('/verify-email-code', { state: { challengeId: res.challenge_id, email: vars.email } })
    },
    onError: (err) => {
      toast.error(err instanceof ApiError ? err.message : 'Не удалось войти')
    },
  })

  return (
    <AuthShell
      title="Вход"
      subtitle="Введите email и пароль. Затем мы отправим код подтверждения на почту."
      footer={
        <>
          Нет аккаунта?{' '}
          <Link to="/register" className="text-accent font-medium">
            Зарегистрироваться
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
            autoComplete="current-password"
            className="input"
            {...register('password')}
          />
          {errors.password && <p className="field-error">{errors.password.message}</p>}
        </div>
        <button className="btn btn-primary w-full" type="submit" disabled={mutation.isPending}>
          {mutation.isPending ? <Spinner size={16} /> : 'Продолжить'}
        </button>
      </form>
    </AuthShell>
  )
}
