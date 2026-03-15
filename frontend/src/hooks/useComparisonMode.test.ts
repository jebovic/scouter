import { describe, it, expect } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useComparisonMode } from './useComparisonMode'

describe('useComparisonMode', () => {
  it('starts with compareMode off and empty selection', () => {
    const { result } = renderHook(() => useComparisonMode())
    expect(result.current.isCompareMode).toBe(false)
    expect(result.current.selectedIds.size).toBe(0)
  })

  it('toggles compare mode on and off', () => {
    const { result } = renderHook(() => useComparisonMode())
    act(() => result.current.setIsCompareMode(true))
    expect(result.current.isCompareMode).toBe(true)
    act(() => result.current.setIsCompareMode(false))
    expect(result.current.isCompareMode).toBe(false)
  })

  it('adds an id when toggle is called with a new id', () => {
    const { result } = renderHook(() => useComparisonMode())
    act(() => result.current.toggle('opt-1'))
    expect(result.current.selectedIds.has('opt-1')).toBe(true)
  })

  it('removes an id when toggle is called with an already-selected id', () => {
    const { result } = renderHook(() => useComparisonMode())
    act(() => result.current.toggle('opt-1'))
    act(() => result.current.toggle('opt-1'))
    expect(result.current.selectedIds.has('opt-1')).toBe(false)
  })

  it('enforces max 4 selected ids', () => {
    const { result } = renderHook(() => useComparisonMode())
    act(() => {
      result.current.toggle('opt-1')
      result.current.toggle('opt-2')
      result.current.toggle('opt-3')
      result.current.toggle('opt-4')
    })
    expect(result.current.selectedIds.size).toBe(4)
    // 5th toggle should be ignored
    act(() => result.current.toggle('opt-5'))
    expect(result.current.selectedIds.size).toBe(4)
    expect(result.current.selectedIds.has('opt-5')).toBe(false)
  })

  it('allows toggling off when at max 4', () => {
    const { result } = renderHook(() => useComparisonMode())
    act(() => {
      result.current.toggle('opt-1')
      result.current.toggle('opt-2')
      result.current.toggle('opt-3')
      result.current.toggle('opt-4')
    })
    // Remove one of the selected
    act(() => result.current.toggle('opt-2'))
    expect(result.current.selectedIds.size).toBe(3)
    expect(result.current.selectedIds.has('opt-2')).toBe(false)
  })

  it('clearSelection resets the selection to empty', () => {
    const { result } = renderHook(() => useComparisonMode())
    act(() => {
      result.current.toggle('opt-1')
      result.current.toggle('opt-2')
    })
    expect(result.current.selectedIds.size).toBe(2)
    act(() => result.current.clearSelection())
    expect(result.current.selectedIds.size).toBe(0)
  })

  it('does not mutate the set — each toggle returns a new Set', () => {
    const { result } = renderHook(() => useComparisonMode())
    let firstSet: Set<string>
    act(() => result.current.toggle('opt-1'))
    firstSet = result.current.selectedIds
    act(() => result.current.toggle('opt-2'))
    expect(result.current.selectedIds).not.toBe(firstSet)
  })
})
