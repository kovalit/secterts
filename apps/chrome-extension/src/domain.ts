// Extract a bare hostname (without a leading www.) from a tab URL.
export function getHostnameFromUrl(url: string): string | null {
  try {
    const parsed = new URL(url)
    if (!['http:', 'https:'].includes(parsed.protocol)) return null
    return parsed.hostname.replace(/^www\./, '').toLowerCase()
  } catch {
    return null
  }
}
