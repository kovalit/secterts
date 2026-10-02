// Parsing helpers for pasted `.env`-style KEY=VALUE blocks.

export type ParsedEnvItem = { key: string; value: string }

// Strip a single pair of matching surrounding quotes from a value.
function unquote(v: string): string {
  if (v.length >= 2) {
    const first = v[0]
    const last = v[v.length - 1]
    if ((first === '"' && last === '"') || (first === "'" && last === "'")) {
      return v.slice(1, -1)
    }
  }
  return v
}

// parseEnv turns a pasted block into key/value pairs. It ignores blank lines
// and `#` comments, tolerates an `export ` prefix, splits on the first `=`
// (so values may contain `=`), and strips matching surrounding quotes.
export function parseEnv(text: string): ParsedEnvItem[] {
  const out: ParsedEnvItem[] = []
  for (const rawLine of text.split(/\r?\n/)) {
    let line = rawLine.trim()
    if (!line || line.startsWith('#')) continue
    if (line.startsWith('export ')) line = line.slice('export '.length).trim()
    const eq = line.indexOf('=')
    if (eq <= 0) continue
    const key = line.slice(0, eq).trim()
    const value = unquote(line.slice(eq + 1).trim())
    if (!key) continue
    out.push({ key, value })
  }
  return out
}

// looksLikeEnv reports whether a pasted block resembles a KEY=VALUE list, so a
// stray clipboard paste does not pop the bulk dialog open.
export function looksLikeEnv(text: string): boolean {
  if (!text || !text.includes('=')) return false
  return parseEnv(text).length > 0
}
