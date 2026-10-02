// Cryptographically-secure password generator used by the entry editor.

export type GenerateOptions = {
  length?: number
  lowercase?: boolean
  uppercase?: boolean
  digits?: boolean
  symbols?: boolean
}

const LOWER = 'abcdefghijkmnpqrstuvwxyz' // no l/o to avoid lookalikes
const UPPER = 'ABCDEFGHJKLMNPQRSTUVWXYZ' // no I/O
const DIGITS = '23456789' // no 0/1
const SYMBOLS = '!@#$%^&*()-_=+[]{};:,.?'

// Pick a uniformly random character from a set using crypto randomness,
// rejecting values that would bias the modulo.
function randomChar(set: string): string {
  const max = Math.floor(256 / set.length) * set.length
  const buf = new Uint8Array(1)
  let v = 0
  do {
    crypto.getRandomValues(buf)
    v = buf[0]
  } while (v >= max)
  return set[v % set.length]
}

export function generatePassword(opts: GenerateOptions = {}): string {
  const {
    length = 20,
    lowercase = true,
    uppercase = true,
    digits = true,
    symbols = true,
  } = opts

  const pools: string[] = []
  if (lowercase) pools.push(LOWER)
  if (uppercase) pools.push(UPPER)
  if (digits) pools.push(DIGITS)
  if (symbols) pools.push(SYMBOLS)
  if (pools.length === 0) pools.push(LOWER)

  const all = pools.join('')
  const len = Math.max(8, Math.min(128, length))

  // Guarantee at least one character from each selected pool.
  const chars: string[] = pools.map((p) => randomChar(p))
  while (chars.length < len) chars.push(randomChar(all))

  // Fisher–Yates shuffle so the guaranteed characters are not positional.
  for (let i = chars.length - 1; i > 0; i--) {
    const buf = new Uint8Array(1)
    let j = 0
    const max = Math.floor(256 / (i + 1)) * (i + 1)
    do {
      crypto.getRandomValues(buf)
      j = buf[0]
    } while (j >= max)
    j = j % (i + 1)
    ;[chars[i], chars[j]] = [chars[j], chars[i]]
  }

  return chars.join('')
}
