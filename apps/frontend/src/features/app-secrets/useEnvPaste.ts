import { useEffect } from 'react'
import { looksLikeEnv } from '../../lib/parseEnv'

function isEditableTarget(el: EventTarget | null): boolean {
  const node = el as HTMLElement | null
  if (!node || !node.tagName) return false
  const tag = node.tagName.toUpperCase()
  if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return true
  return node.isContentEditable
}

// useEnvPaste watches for a clipboard paste of a KEY=VALUE block anywhere on the
// page (except inside form fields) and hands the raw text to onDetected. It is
// disabled while `enabled` is false — e.g. when the dialog is already open.
export function useEnvPaste(enabled: boolean, onDetected: (text: string) => void) {
  useEffect(() => {
    if (!enabled) return
    const handler = (e: ClipboardEvent) => {
      if (isEditableTarget(e.target)) return
      const text = e.clipboardData?.getData('text/plain') ?? ''
      if (!looksLikeEnv(text)) return
      e.preventDefault()
      onDetected(text)
    }
    window.addEventListener('paste', handler)
    return () => window.removeEventListener('paste', handler)
  }, [enabled, onDetected])
}
