import { useEffect, useRef, useState } from 'react'
import { Copy, Eye, EyeOff } from 'lucide-react'
import { copyToClipboard } from '../../lib/clipboard'
import { useToast } from '../../components/ui/Toast'
import { Spinner } from '../../components/ui/Spinner'

// Reusable reveal/copy control. Fetches the secret only on demand, hides it
// again after 30s and never keeps it in state longer than necessary.
export function RevealValue({
  fetchValue,
  copiedMessage = 'Скопировано',
}: {
  fetchValue: () => Promise<string>
  copiedMessage?: string
}) {
  const toast = useToast()
  const [value, setValue] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const timer = useRef<number | null>(null)

  const clearTimer = () => {
    if (timer.current) window.clearTimeout(timer.current)
    timer.current = null
  }

  // Auto-hide 30s after reveal; always clear value on unmount.
  useEffect(() => () => clearTimer(), [])

  const scheduleHide = () => {
    clearTimer()
    timer.current = window.setTimeout(() => setValue(null), 30_000)
  }

  const onReveal = async () => {
    if (value !== null) {
      setValue(null)
      clearTimer()
      return
    }
    setLoading(true)
    try {
      const v = await fetchValue()
      setValue(v)
      scheduleHide()
    } catch {
      toast.error('Не удалось получить значение')
    } finally {
      setLoading(false)
    }
  }

  const onCopy = async () => {
    setLoading(true)
    try {
      // Fetch transiently when hidden; the value is not stored in that case.
      const v = value ?? (await fetchValue())
      const ok = await copyToClipboard(v)
      if (ok) toast.success(copiedMessage)
      else toast.error('Не удалось скопировать')
    } catch {
      toast.error('Не удалось получить значение')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="inline-flex items-center gap-1.5">
      {value !== null ? (
        <code className="keychip max-w-[220px] truncate" title={value}>
          {value}
        </code>
      ) : (
        <span className="font-mono text-ink-faint tracking-widest select-none">••••••••</span>
      )}
      <button className="icon-btn w-8 h-8" onClick={onReveal} disabled={loading} title="Показать / скрыть">
        {loading ? <Spinner size={14} /> : value !== null ? <EyeOff size={15} /> : <Eye size={15} />}
      </button>
      <button className="icon-btn w-8 h-8" onClick={onCopy} disabled={loading} title="Скопировать">
        <Copy size={15} />
      </button>
    </div>
  )
}
