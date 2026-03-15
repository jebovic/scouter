import { useState } from 'react'

const MAX_COMPARE = 4

export function useMissionComparison() {
  const [selectedIds, setSelectedIds] = useState<string[]>([])

  const toggle = (id: string) => {
    setSelectedIds(prev =>
      prev.includes(id)
        ? prev.filter(x => x !== id)
        : prev.length < MAX_COMPARE
          ? [...prev, id]
          : prev,
    )
  }

  const clear = () => setSelectedIds([])

  const isSelected = (id: string) => selectedIds.includes(id)

  const canAdd = (id: string) =>
    selectedIds.includes(id) || selectedIds.length < MAX_COMPARE

  return { selectedIds, toggle, clear, isSelected, canAdd }
}
