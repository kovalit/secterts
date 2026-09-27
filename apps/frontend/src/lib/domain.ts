// Extract a bare hostname (without www.) from a URL string, for display.
export function domainFromUrl(url?: string | null): string | null {
  if (!url) return null
  let value = url.trim()
  if (!value.includes('://')) value = 'https://' + value
  try {
    const parsed = new URL(value)
    if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') return null
    return parsed.hostname.replace(/^www\./, '').toLowerCase()
  } catch {
    return null
  }
}
