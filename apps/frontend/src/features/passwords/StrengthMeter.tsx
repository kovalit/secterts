import { STRENGTH_LABELS } from '../../lib/entryTypes'

// Common weak passwords, mirrors the backend blocklist for consistent feedback.
const COMMON = new Set([
  'password', 'passw0rd', '123456', '12345678', '123456789', 'qwerty',
  'qwerty123', '111111', '123123', 'abc123', 'letmein', 'iloveyou',
  'admin', 'welcome', 'monkey', 'dragon', 'secret', 'master', '000000', 'football',
])

// Live 0..4 strength estimate, mirroring the backend EstimateStrength logic so
// the meter matches the score stored on save.
export function estimateStrength(password: string): number {
  if (!password) return 0
  const length = password.length
  if (length < 8 || COMMON.has(password.trim().toLowerCase())) return 0

  let classes = 0
  if (/[a-z]/.test(password)) classes++
  if (/[A-Z]/.test(password)) classes++
  if (/[0-9]/.test(password)) classes++
  if (/[^a-zA-Z0-9]/.test(password)) classes++

  let score = 0
  if (length >= 16) score += 2
  else if (length >= 12) score += 1
  score += classes - 1

  return Math.max(0, Math.min(4, score))
}

const COLORS = ['bg-red-500', 'bg-red-500', 'bg-amber-500', 'bg-emerald-500', 'bg-emerald-600']

export function StrengthMeter({
  password,
  className = '',
}: {
  password: string
  className?: string
}) {
  const score = estimateStrength(password)
  const label = STRENGTH_LABELS[score]

  return (
    <div className={className}>
      <div className="flex gap-1">
        {[0, 1, 2, 3, 4].map((i) => (
          <span
            key={i}
            className={
              'h-1.5 flex-1 rounded-full ' + (i <= score ? COLORS[score] : 'bg-ink-strong/10')
            }
          />
        ))}
      </div>
      <p className="mt-1 text-[12px] text-ink-muted">
        Надёжность: <span className="font-medium text-ink">{label}</span>
      </p>
    </div>
  )
}
