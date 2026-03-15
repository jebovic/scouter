import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook } from '@testing-library/react'
import { useBudgetRollup } from './useBudgetRollup'
import type { Mission } from '../types'

// Mock useMissions
vi.mock('./useMission', () => ({
  useMissions: vi.fn(),
}))

import { useMissions } from './useMission'

const mockUseMissions = useMissions as ReturnType<typeof vi.fn>

function makeMission(overrides: Partial<Mission>): Mission {
  return {
    id: 'uuid-1',
    slug: 'test-mission',
    name: 'Test Mission',
    icon: '🎯',
    category: 'electronics',
    budget: 1000,
    currency: 'EUR',
    locale: 'fr-FR',
    phase: 'researching',
    constraints: [],
    costCategories: [],
    timeline: [],
    weightProfile: { price: 0, quality: 0, feature: 0 },
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  }
}

describe('useBudgetRollup', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns isLoading true when missions are loading', () => {
    mockUseMissions.mockReturnValue({ missions: [], isLoading: true, error: null })
    const { result } = renderHook(() => useBudgetRollup())
    expect(result.current.isLoading).toBe(true)
  })

  it('returns empty rollup when no missions', () => {
    mockUseMissions.mockReturnValue({ missions: [], isLoading: false, error: null })
    const { result } = renderHook(() => useBudgetRollup())
    const { rollup } = result.current
    expect(rollup.totalBudget).toBe(0)
    expect(rollup.entries).toHaveLength(0)
    expect(rollup.activeMissions).toBe(0)
    expect(rollup.purchasedTotal).toBe(0)
    expect(rollup.byCategory).toEqual({})
    expect(rollup.byPhase).toEqual({})
  })

  it('filters out missions with budget = 0', () => {
    mockUseMissions.mockReturnValue({
      missions: [
        makeMission({ id: 'uuid-1', budget: 0 }),
        makeMission({ id: 'uuid-2', budget: 500 }),
      ],
      isLoading: false,
      error: null,
    })
    const { result } = renderHook(() => useBudgetRollup())
    expect(result.current.rollup.entries).toHaveLength(1)
    expect(result.current.rollup.entries[0].budget).toBe(500)
  })

  it('computes totalBudget as sum of non-done mission budgets', () => {
    mockUseMissions.mockReturnValue({
      missions: [
        makeMission({ id: 'uuid-1', budget: 1000, phase: 'researching' }),
        makeMission({ id: 'uuid-2', budget: 2000, phase: 'comparing' }),
        makeMission({ id: 'uuid-3', budget: 500, phase: 'done' }),
      ],
      isLoading: false,
      error: null,
    })
    const { result } = renderHook(() => useBudgetRollup())
    // totalBudget excludes 'done' missions
    expect(result.current.rollup.totalBudget).toBe(3000)
  })

  it('computes purchasedTotal as sum of done mission budgets', () => {
    mockUseMissions.mockReturnValue({
      missions: [
        makeMission({ id: 'uuid-1', budget: 1000, phase: 'researching' }),
        makeMission({ id: 'uuid-2', budget: 500, phase: 'done' }),
        makeMission({ id: 'uuid-3', budget: 300, phase: 'done' }),
      ],
      isLoading: false,
      error: null,
    })
    const { result } = renderHook(() => useBudgetRollup())
    expect(result.current.rollup.purchasedTotal).toBe(800)
  })

  it('computes activeMissions as count of non-done missions with budget > 0', () => {
    mockUseMissions.mockReturnValue({
      missions: [
        makeMission({ id: 'uuid-1', budget: 1000, phase: 'researching' }),
        makeMission({ id: 'uuid-2', budget: 2000, phase: 'comparing' }),
        makeMission({ id: 'uuid-3', budget: 500, phase: 'done' }),
        makeMission({ id: 'uuid-4', budget: 0, phase: 'researching' }),
      ],
      isLoading: false,
      error: null,
    })
    const { result } = renderHook(() => useBudgetRollup())
    expect(result.current.rollup.activeMissions).toBe(2)
  })

  it('groups byCategory correctly', () => {
    mockUseMissions.mockReturnValue({
      missions: [
        makeMission({ id: 'uuid-1', budget: 1000, category: 'electronics', phase: 'researching' }),
        makeMission({ id: 'uuid-2', budget: 2000, category: 'electronics', phase: 'comparing' }),
        makeMission({ id: 'uuid-3', budget: 500, category: 'travel', phase: 'researching' }),
      ],
      isLoading: false,
      error: null,
    })
    const { result } = renderHook(() => useBudgetRollup())
    expect(result.current.rollup.byCategory['electronics']).toBe(3000)
    expect(result.current.rollup.byCategory['travel']).toBe(500)
  })

  it('groups byPhase correctly', () => {
    mockUseMissions.mockReturnValue({
      missions: [
        makeMission({ id: 'uuid-1', budget: 1000, phase: 'researching' }),
        makeMission({ id: 'uuid-2', budget: 2000, phase: 'researching' }),
        makeMission({ id: 'uuid-3', budget: 500, phase: 'comparing' }),
        makeMission({ id: 'uuid-4', budget: 300, phase: 'done' }),
      ],
      isLoading: false,
      error: null,
    })
    const { result } = renderHook(() => useBudgetRollup())
    expect(result.current.rollup.byPhase['researching']).toBe(3000)
    expect(result.current.rollup.byPhase['comparing']).toBe(500)
    expect(result.current.rollup.byPhase['done']).toBe(300)
  })

  it('includes done missions in byCategory grouping', () => {
    mockUseMissions.mockReturnValue({
      missions: [
        makeMission({ id: 'uuid-1', budget: 1000, category: 'electronics', phase: 'researching' }),
        makeMission({ id: 'uuid-2', budget: 500, category: 'electronics', phase: 'done' }),
      ],
      isLoading: false,
      error: null,
    })
    const { result } = renderHook(() => useBudgetRollup())
    // byCategory includes ALL missions (both active and done) with budget > 0
    expect(result.current.rollup.byCategory['electronics']).toBe(1500)
  })

  it('sets correct MissionBudgetEntry fields', () => {
    const mission = makeMission({
      id: 'uuid-abc',
      slug: 'my-mission',
      name: 'My Mission',
      category: 'travel',
      budget: 1500,
      phase: 'comparing',
      currency: 'EUR',
    })
    mockUseMissions.mockReturnValue({
      missions: [mission],
      isLoading: false,
      error: null,
    })
    const { result } = renderHook(() => useBudgetRollup())
    const entry = result.current.rollup.entries[0]
    expect(entry.missionId).toBe('uuid-abc')
    expect(entry.missionSlug).toBe('my-mission')
    expect(entry.missionName).toBe('My Mission')
    expect(entry.category).toBe('travel')
    expect(entry.budget).toBe(1500)
    expect(entry.phase).toBe('comparing')
    expect(entry.currency).toBe('EUR')
  })

  it('handles all missions with budget > 0 in entries list', () => {
    mockUseMissions.mockReturnValue({
      missions: [
        makeMission({ id: 'uuid-1', budget: 100 }),
        makeMission({ id: 'uuid-2', budget: 200 }),
        makeMission({ id: 'uuid-3', budget: 0 }),
        makeMission({ id: 'uuid-4', budget: 300 }),
      ],
      isLoading: false,
      error: null,
    })
    const { result } = renderHook(() => useBudgetRollup())
    expect(result.current.rollup.entries).toHaveLength(3)
  })
})
