import { useState, useEffect, useRef } from 'react'
import { useQuery } from '@tanstack/react-query'
import { searchOptions, getSimilarOptions } from '../api/search'
import type { SearchResult } from '../api/search'

const DEBOUNCE_MS = 300

export function useSearch(query: string, limit = 10) {
  const [debouncedQuery, setDebouncedQuery] = useState(query)

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedQuery(query), DEBOUNCE_MS)
    return () => clearTimeout(timer)
  }, [query])

  return useQuery<SearchResult[]>({
    queryKey: ['search', debouncedQuery, limit],
    queryFn: () => searchOptions(debouncedQuery, limit),
    enabled: debouncedQuery.trim().length >= 2,
    staleTime: 30_000,
  })
}

export function useSimilarOptions(optionId: string | null, limit = 5) {
  return useQuery<SearchResult[]>({
    queryKey: ['similar', optionId, limit],
    queryFn: () => getSimilarOptions(optionId!, limit),
    enabled: optionId != null,
    staleTime: 60_000,
  })
}

export function useSearchInput() {
  const [query, setQuery] = useState('')
  const [isOpen, setIsOpen] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  const open = () => setIsOpen(true)
  const close = () => setIsOpen(false)
  const clear = () => {
    setQuery('')
    setIsOpen(false)
  }

  return { query, setQuery, isOpen, open, close, clear, inputRef }
}
