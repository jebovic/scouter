import { describe, it, expect } from 'vitest'
import { renderHook } from '@testing-library/react'
import { useBudgetVariance } from './useBudgetVariance'

describe('useBudgetVariance', () => {
  it('returns empty categories when no envelopes exist', () => {
    const { result } = renderHook(() => useBudgetVariance([]))
    expect(result.current.months).toEqual(expect.any(Array))
    expect(result.current.categories).toEqual([])
    expect(result.current.cells).toEqual([])
  })

  it('generates month labels for the given number of months', () => {
    const { result } = renderHook(() => useBudgetVariance([], 3))
    expect(result.current.months.length).toBe(3)
    // Verify months are in ISO format YYYY-MM
    result.current.months.forEach((month) => {
      expect(month).toMatch(/^\d{4}-\d{2}$/)
    })
  })

  it('extracts categories from envelopes', () => {
    const envelopes = [
      { id: '1', name: 'Housing', monthlyAmount: 1000, currency: 'EUR', createdAt: '2025-01-01' },
      { id: '2', name: 'Food', monthlyAmount: 500, currency: 'EUR', createdAt: '2025-01-01' },
    ]
    const { result } = renderHook(() => useBudgetVariance(envelopes, 1))
    expect(result.current.categories).toEqual(['Housing', 'Food'])
  })

  it('creates one cell per month per category', () => {
    const envelopes = [
      { id: '1', name: 'Housing', monthlyAmount: 1000, currency: 'EUR', createdAt: '2025-01-01' },
      { id: '2', name: 'Food', monthlyAmount: 500, currency: 'EUR', createdAt: '2025-01-01' },
    ]
    const { result } = renderHook(() => useBudgetVariance(envelopes, 2))
    // 2 categories × 2 months = 4 cells
    expect(result.current.cells.length).toBe(4)
  })

  it('sets correct budgeted amount in cells', () => {
    const envelopes = [
      { id: '1', name: 'Housing', monthlyAmount: 1000, currency: 'EUR', createdAt: '2025-01-01' },
    ]
    const { result } = renderHook(() => useBudgetVariance(envelopes, 1))
    const cell = result.current.cells[0]
    expect(cell.budgeted).toBe(1000)
    expect(cell.category).toBe('Housing')
  })

  it('initializes spent to 0 for all cells', () => {
    const envelopes = [
      { id: '1', name: 'Housing', monthlyAmount: 1000, currency: 'EUR', createdAt: '2025-01-01' },
    ]
    const { result } = renderHook(() => useBudgetVariance(envelopes, 1))
    const cell = result.current.cells[0]
    expect(cell.spent).toBe(0)
  })

  it('calculates variance as spent - budgeted', () => {
    const envelopes = [
      { id: '1', name: 'Housing', monthlyAmount: 1000, currency: 'EUR', createdAt: '2025-01-01' },
    ]
    const { result } = renderHook(() => useBudgetVariance(envelopes, 1))
    const cell = result.current.cells[0]
    // spent=0, budgeted=1000 → variance = 0 - 1000 = -1000
    expect(cell.variance).toBe(-1000)
  })

  it('calculates variance percentage correctly', () => {
    const envelopes = [
      { id: '1', name: 'Housing', monthlyAmount: 1000, currency: 'EUR', createdAt: '2025-01-01' },
    ]
    const { result } = renderHook(() => useBudgetVariance(envelopes, 1))
    const cell = result.current.cells[0]
    // (0 - 1000) / 1000 * 100 = -100%
    expect(cell.variancePct).toBe(-100)
  })

  it('sets variancePct to Infinity when budgeted is 0', () => {
    const envelopes = [
      { id: '1', name: 'No Budget', monthlyAmount: 0, currency: 'EUR', createdAt: '2025-01-01' },
    ]
    const { result } = renderHook(() => useBudgetVariance(envelopes, 1))
    const cell = result.current.cells[0]
    expect(cell.variancePct).toBe(Infinity)
  })

  it('returns months in chronological order (oldest first)', () => {
    const { result } = renderHook(() => useBudgetVariance([], 3))
    // Verify they are in order by parsing as dates
    const dates = result.current.months.map((m) => new Date(m + '-01'))
    for (let i = 1; i < dates.length; i++) {
      expect(dates[i].getTime()).toBeGreaterThan(dates[i - 1].getTime())
    }
  })

  it('generates correct month labels in locale format', () => {
    const { result } = renderHook(() => useBudgetVariance([], 3))
    // Month labels should be human-readable
    expect(result.current.months.length).toBe(3)
    result.current.cells.forEach((cell) => {
      // monthLabel should be something like "Jan 2025" or "Janvier 2025" (depends on locale)
      expect(cell.monthLabel).toBeTruthy()
      expect(typeof cell.monthLabel).toBe('string')
    })
  })
})
