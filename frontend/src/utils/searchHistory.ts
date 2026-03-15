const STORAGE_KEY = 'scouter_search_history'
const MAX_ENTRIES = 10

export function getHistory(): string[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

export function addToHistory(query: string): void {
  const trimmed = query.trim()
  if (!trimmed) return
  const existing = getHistory().filter((q) => q.toLowerCase() !== trimmed.toLowerCase())
  const next = [trimmed, ...existing].slice(0, MAX_ENTRIES)
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(next))
  } catch {
    // localStorage may be unavailable (private browsing, quota exceeded)
  }
}

export function removeFromHistory(query: string): void {
  const next = getHistory().filter((q) => q.toLowerCase() !== query.toLowerCase())
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(next))
  } catch {
    // ignore
  }
}

export function clearHistory(): void {
  try {
    localStorage.removeItem(STORAGE_KEY)
  } catch {
    // ignore
  }
}
