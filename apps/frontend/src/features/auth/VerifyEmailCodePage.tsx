import { useState } from 'react'
import { Navigate, useLocation, useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { AuthShell } from './AuthShell'
import { authApi } from '../../api/auth'
import { ApiError } from '../../api/client'
import { useAuth } from './AuthProvider'
import { useToast } from '../../components/ui/Toast'
import { Spinner } from '../../components/ui/Spinner'

type NavState = { challengeId?: string; email?: string }

export function VerifyEmailCodePage() {
  const location = useLocation()
  const navigate = useNavigate()
  const toast = useToast()
  const { setUser } = useAuth()
  const state = (location.state as NavState) ?? {}
  const [code, setCode] = useState('')

  const mutation = useMutation({
    mutationFn: () => authApi.verify(state.challengeId as string, code),
    onSuccess: (user) => {
      setUser(user)
      toast.success('Вход выполнен')
      navigate('/passwords', { replace: true })
    },
    onError: (err) => {
      toast.error(err instanceof ApiError ? err.message : 'Неверный код')
    },
  })

  if (!state.challengeId) {
    return <Navigate to="/login" replace />
  }

  return (
    <AuthShell
      title="Подтверждение входа"
      subtitle={
        state.email
          ? `Мы отправили 6-значный код на ${state.email}`
          : 'Введите код из письма'
      }
    >
      <form
        className="flex flex-col gap-4"
        onSubmit={(e) => {
          e.preventDefault()
          if (code.length >= 4) mutation.mutate()
        }}
      >
        <div>
          <label className="field-label" htmlFor="code">
            Код подтверждения
          </label>
          <input
            id="code"
            inputMode="numeric"
            autoComplete="one-time-code"
            maxLength={6}
            className="input text-center tracking-[0.5em] text-lg font-mono"
            placeholder="000000"
            value={code}
            onChange={(e) => setCode(e.target.value.replace(/\D/g, ''))}
            autoFocus
          />
        </div>
        <button
          className="btn btn-primary w-full"
          type="submit"
          disabled={mutation.isPending || code.length < 4}
        >
          {mutation.isPending ? <Spinner size={16} /> : 'Войти'}
        </button>
        <p className="text-[13px] text-ink-muted text-center">
          Код действует 10 минут. В dev-режиме он выводится в логи backend.
        </p>
      </form>
    </AuthShell>
  )
}
