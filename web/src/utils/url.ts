export function normalizeBookmarkUrl(raw: string): string {
  try {
    const parsed = new URL(raw.trim())
    parsed.hash = ''
    if (parsed.pathname !== '/' && parsed.pathname.endsWith('/')) {
      parsed.pathname = parsed.pathname.slice(0, -1)
    }
    return parsed.toString()
  } catch {
    return ''
  }
}

export function bookmarkUrlsMatch(a: string, b: string): boolean {
  const left = normalizeBookmarkUrl(a)
  const right = normalizeBookmarkUrl(b)
  if (!left || !right) return a.trim() === b.trim()
  return left === right
}
