import { useState } from 'react'

const MAX_SELECTION = 4

export interface ComparisonModeState {
  isCompareMode: boolean
  setIsCompareMode: (value: boolean) => void
  selectedIds: Set<string>
  toggle: (id: string) => void
  clearSelection: () => void
}

export function useComparisonMode(): ComparisonModeState {
  const [isCompareMode, setIsCompareMode] = useState(false)
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())

  const toggle = (id: string) => {
    setSelectedIds((prev) => {
      if (prev.has(id)) {
        const next = new Set(prev)
        next.delete(id)
        return next
      }
      if (prev.size >= MAX_SELECTION) {
        return prev
      }
      return new Set(prev).add(id)
    })
  }

  const clearSelection = () => {
    setSelectedIds(new Set())
  }

  return { isCompareMode, setIsCompareMode, selectedIds, toggle, clearSelection }
}
