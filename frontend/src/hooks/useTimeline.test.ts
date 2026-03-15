import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook } from '@testing-library/react'
import { useTimeline } from './useTimeline'
import * as missionHook from './useMission'
import * as optionsHook from './useOptions'
import * as shoppingHook from './useShopping'

describe('useTimeline', () => {
  const missionId = 'test-mission-id'
  const missionSlug = 'test-mission'

  const mockMission = {
    id: missionId,
    slug: missionSlug,
    name: 'Test Mission',
    icon: '🚀',
    category: 'electronics' as const,
    budget: 1000,
    currency: 'EUR',
    locale: 'fr-FR',
    phase: 'researching' as const,
    constraints: [],
    costCategories: [],
    timeline: [],
    weightProfile: { price: 0.5, quality: 0.3, feature: 0.2 },
    createdAt: '2026-03-01T10:00:00Z',
    updatedAt: '2026-03-01T10:00:00Z',
  }

  const mockOptions = [
    {
      id: 'option-1',
      missionId,
      name: 'Option 1',
      category: 'laptop',
      badge: 'recommended' as const,
      attributes: [],
      warnings: [],
      pinned: false,
      rejected: false,
      createdAt: '2026-03-02T11:00:00Z',
    },
    {
      id: 'option-2',
      missionId,
      name: 'Option 2',
      category: 'laptop',
      badge: 'alternative' as const,
      attributes: [],
      warnings: [],
      pinned: false,
      rejected: false,
      createdAt: '2026-03-03T12:00:00Z',
    },
  ]

  const mockShoppingItems = [
    {
      id: 'item-1',
      missionId,
      name: 'Laptop Item',
      merchant: 'Amazon',
      costCategory: 'electronics',
      price: 800,
      status: 'watch' as const,
      pinned: false,
      createdAt: '2026-03-04T13:00:00Z',
    },
    {
      id: 'item-2',
      missionId,
      name: 'Keyboard',
      merchant: 'Local Store',
      costCategory: 'accessories',
      price: 50,
      status: 'buy' as const,
      pinned: false,
      createdAt: '2026-03-05T14:00:00Z',
    },
  ]

  beforeEach(() => {
    vi.spyOn(missionHook, 'useMission').mockReturnValue({
      mission: mockMission,
      isLoading: false,
      error: null,
    })
    vi.spyOn(optionsHook, 'useOptions').mockReturnValue({
      options: mockOptions,
      isLoading: false,
      error: null,
    })
    vi.spyOn(shoppingHook, 'useShopping').mockReturnValue({
      items: mockShoppingItems,
      isLoading: false,
      error: null,
    })
  })

  it('creates mission_created event from mission.createdAt', () => {
    const { result } = renderHook(() => useTimeline(missionId, missionSlug))

    const missionEvent = result.current.events.find((e) => e.type === 'mission_created')
    expect(missionEvent).toBeDefined()
    expect(missionEvent?.title).toBe('Mission créée')
    expect(missionEvent?.icon).toBe('🚀')
    expect(missionEvent?.date.getTime()).toBe(new Date(mockMission.createdAt).getTime())
  })

  it('creates option_added events from options.createdAt', () => {
    const { result } = renderHook(() => useTimeline(missionId, missionSlug))

    const optionEvents = result.current.events.filter((e) => e.type === 'option_added')
    expect(optionEvents.length).toBe(2)
    expect(optionEvents[0].title).toContain('Option ajoutée')
    expect(optionEvents[0].icon).toBe('📦')
  })

  it('creates item_added events from shopping items', () => {
    const { result } = renderHook(() => useTimeline(missionId, missionSlug))

    const itemEvents = result.current.events.filter((e) => e.type === 'item_added')
    expect(itemEvents.length).toBe(2)
    expect(itemEvents[0].icon).toBe('🛒')
  })

  it('creates item_purchased events for purchased items', () => {
    const { result } = renderHook(() => useTimeline(missionId, missionSlug))

    const purchasedEvents = result.current.events.filter((e) => e.type === 'item_purchased')
    expect(purchasedEvents.length).toBe(1)
    expect(purchasedEvents[0].icon).toBe('✅')
  })

  it('sorts events newest first', () => {
    const { result } = renderHook(() => useTimeline(missionId, missionSlug))

    const dates = result.current.events.map((e) => new Date(e.date).getTime())
    for (let i = 0; i < dates.length - 1; i++) {
      expect(dates[i]).toBeGreaterThanOrEqual(dates[i + 1])
    }
  })

  it('generates stable event IDs', () => {
    const { result: result1 } = renderHook(() => useTimeline(missionId, missionSlug))
    const { result: result2 } = renderHook(() => useTimeline(missionId, missionSlug))

    const ids1 = result1.current.events.map((e) => e.id)
    const ids2 = result2.current.events.map((e) => e.id)

    expect(ids1).toEqual(ids2)
  })

  it('generates unique event IDs', () => {
    const { result } = renderHook(() => useTimeline(missionId, missionSlug))

    const ids = result.current.events.map((e) => e.id)
    const uniqueIds = new Set(ids)

    expect(uniqueIds.size).toBe(ids.length)
  })

  it('returns isLoading true when any data is loading', () => {
    vi.spyOn(missionHook, 'useMission').mockReturnValue({
      mission: mockMission,
      isLoading: true,
      error: null,
    })

    const { result } = renderHook(() => useTimeline(missionId, missionSlug))

    expect(result.current.isLoading).toBe(true)
  })

  it('returns isLoading false when all data is loaded', () => {
    const { result } = renderHook(() => useTimeline(missionId, missionSlug))

    expect(result.current.isLoading).toBe(false)
  })

  it('handles missing mission data gracefully', () => {
    vi.spyOn(missionHook, 'useMission').mockReturnValue({
      mission: undefined,
      isLoading: false,
      error: null,
    })

    const { result } = renderHook(() => useTimeline(missionId, missionSlug))

    expect(result.current.events).toBeDefined()
    expect(Array.isArray(result.current.events)).toBe(true)
  })
})
