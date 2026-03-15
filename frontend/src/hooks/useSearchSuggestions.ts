import { useState, useEffect, useCallback } from 'react'
import { getHistory, clearHistory as clearHistoryUtil } from '../utils/searchHistory'

const POPULAR_TERMS: string[] = [
  'iPhone',
  'MacBook',
  'PS5',
  'TV 4K',
  'aspirateur robot',
  'vélo électrique',
  'AirPods',
  'iPad',
  'Samsung Galaxy',
  'Nintendo Switch',
  'casque audio',
  'machine à café',
  'lave-linge',
  'réfrigérateur',
  'drone',
  'imprimante',
  'écran PC',
  'clavier mécanique',
  'trottinette électrique',
  'robot cuiseur',
]

export interface SearchSuggestions {
  suggestions: string[]
  history: string[]
  clearHistory: () => void
  refreshHistory: () => void
}

export function useSearchSuggestions(query: string): SearchSuggestions {
  const [history, setHistory] = useState<string[]>(() => getHistory())

  const refreshHistory = useCallback(() => {
    setHistory(getHistory())
  }, [])

  const clearHistory = useCallback(() => {
    clearHistoryUtil()
    setHistory([])
  }, [])

  // Re-read history on query change (user may have navigated away and back)
  useEffect(() => {
    setHistory(getHistory())
  }, [query])

  const trimmed = query.trim().toLowerCase()

  const suggestions = trimmed
    ? POPULAR_TERMS.filter(
        (term) =>
          term.toLowerCase().includes(trimmed) &&
          !history.some((h) => h.toLowerCase() === term.toLowerCase()),
      )
    : []

  return { suggestions, history, clearHistory, refreshHistory }
}
