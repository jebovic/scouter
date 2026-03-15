import { describe, it, expect } from 'vitest'
import { computeMissionProgress } from './missionProgress'

describe('missionProgress', () => {
  const baseTime = new Date('2026-01-01T00:00:00Z').getTime()

  // Budget calculations
  it('calculates budget spent as sum of item prices', () => {
    const mission = { budget: 1000, createdAt: new Date(baseTime).toISOString() }
    const items = [
      { currentPrice: 200, targetPrice: 150 },
      { currentPrice: 300, targetPrice: 300 },
      { currentPrice: 150, targetPrice: 100 },
    ]
    const result = computeMissionProgress(mission, items)
    expect(result.budgetSpent).toBe(650)
  })

  it('calculates budgetPercent 0-100', () => {
    const mission = { budget: 1000, createdAt: new Date(baseTime).toISOString() }
    const items = [{ currentPrice: 600, targetPrice: 500 }]
    const result = computeMissionProgress(mission, items)
    expect(result.budgetPercent).toBe(60)
  })

  it('marks budget as on_track when spent is 25-75% of budget', () => {
    const mission = { budget: 1000, createdAt: new Date(baseTime).toISOString() }
    const items = [{ currentPrice: 500, targetPrice: 450 }]
    const result = computeMissionProgress(mission, items)
    expect(result.budgetStatus).toBe('on_track')
  })

  it('marks budget as under_budget when spent is less than 50% of budget', () => {
    const mission = { budget: 1000, createdAt: new Date(baseTime).toISOString() }
    const items = [{ currentPrice: 400, targetPrice: 350 }]
    const result = computeMissionProgress(mission, items)
    expect(result.budgetStatus).toBe('under_budget')
  })

  it('marks budget as over_budget when spent exceeds budget', () => {
    const mission = { budget: 1000, createdAt: new Date(baseTime).toISOString() }
    const items = [{ currentPrice: 1200, targetPrice: 1000 }]
    const result = computeMissionProgress(mission, items)
    expect(result.budgetStatus).toBe('over_budget')
  })

  // Readiness calculations
  it('counts items as ready when currentPrice <= targetPrice', () => {
    const mission = { budget: 1000, createdAt: new Date(baseTime).toISOString() }
    const items = [
      { currentPrice: 100, targetPrice: 150 }, // ready (100 <= 150)
      { currentPrice: 200, targetPrice: 150 }, // not ready (200 > 150)
      { currentPrice: 50, targetPrice: 50 },   // ready (50 <= 50)
    ]
    const result = computeMissionProgress(mission, items)
    expect(result.readyItems).toBe(2)
    expect(result.totalItems).toBe(3)
  })

  it('counts items as ready when no targetPrice is set', () => {
    const mission = { budget: 1000, createdAt: new Date(baseTime).toISOString() }
    const items = [
      { currentPrice: 100 }, // ready (no target)
      { currentPrice: 200, targetPrice: 150 }, // not ready
      { currentPrice: 50, targetPrice: undefined }, // ready (no target)
    ]
    const result = computeMissionProgress(mission, items)
    expect(result.readyItems).toBe(2)
  })

  it('calculates readinessPercent 0-100', () => {
    const mission = { budget: 1000, createdAt: new Date(baseTime).toISOString() }
    const items = [
      { currentPrice: 100, targetPrice: 150 },
      { currentPrice: 200, targetPrice: 150 },
      { currentPrice: 50, targetPrice: 50 },
      { currentPrice: 75, targetPrice: 100 },
    ]
    const result = computeMissionProgress(mission, items)
    expect(result.readinessPercent).toBe(75) // 3 out of 4
  })

  // Time calculations
  it('calculates daysElapsed from createdAt', () => {
    const createdTime = new Date(baseTime).toISOString()
    const currentTime = new Date(baseTime + 5 * 24 * 60 * 60 * 1000).toISOString()
    const mission = { budget: 1000, createdAt: createdTime }
    const items: never[] = []

    // Mock current time for test
    vi.useFakeTimers()
    vi.setSystemTime(new Date(currentTime))
    const result = computeMissionProgress(mission, items)
    vi.useRealTimers()

    expect(result.daysElapsed).toBe(5)
  })

  it('calculates timePercent based on target days', () => {
    const createdTime = new Date(baseTime).toISOString()
    const currentTime = new Date(baseTime + 45 * 24 * 60 * 60 * 1000).toISOString()
    const mission = { budget: 1000, createdAt: createdTime }
    const items: never[] = []

    vi.useFakeTimers()
    vi.setSystemTime(new Date(currentTime))
    const result = computeMissionProgress(mission, items, 90)
    vi.useRealTimers()

    expect(result.timePercent).toBe(50) // 45/90
  })

  it('defaults to 90 days target when not specified', () => {
    const createdTime = new Date(baseTime).toISOString()
    const currentTime = new Date(baseTime + 30 * 24 * 60 * 60 * 1000).toISOString()
    const mission = { budget: 1000, createdAt: createdTime }
    const items: never[] = []

    vi.useFakeTimers()
    vi.setSystemTime(new Date(currentTime))
    const result = computeMissionProgress(mission, items)
    vi.useRealTimers()

    expect(result.daysTotal).toBe(90)
  })

  // Urgency
  it('marks urgency as high when timePercent > 75 and readinessPercent < 50', () => {
    const createdTime = new Date(baseTime).toISOString()
    const currentTime = new Date(baseTime + 80 * 24 * 60 * 60 * 1000).toISOString()
    const mission = { budget: 1000, createdAt: createdTime }
    const items = [
      { currentPrice: 200, targetPrice: 150 }, // not ready
      { currentPrice: 300, targetPrice: 250 }, // not ready
    ]

    vi.useFakeTimers()
    vi.setSystemTime(new Date(currentTime))
    const result = computeMissionProgress(mission, items, 90)
    vi.useRealTimers()

    expect(result.urgency).toBe('high')
  })

  it('marks urgency as medium when timePercent > 50 and readinessPercent < 75', () => {
    const createdTime = new Date(baseTime).toISOString()
    const currentTime = new Date(baseTime + 50 * 24 * 60 * 60 * 1000).toISOString()
    const mission = { budget: 1000, createdAt: createdTime }
    const items = [
      { currentPrice: 100, targetPrice: 150 }, // ready
      { currentPrice: 200, targetPrice: 150 }, // not ready (200 > 150)
    ]

    vi.useFakeTimers()
    vi.setSystemTime(new Date(currentTime))
    const result = computeMissionProgress(mission, items, 90)
    vi.useRealTimers()

    expect(result.urgency).toBe('medium')
  })

  it('marks urgency as low otherwise', () => {
    const createdTime = new Date(baseTime).toISOString()
    const currentTime = new Date(baseTime + 10 * 24 * 60 * 60 * 1000).toISOString()
    const mission = { budget: 1000, createdAt: createdTime }
    const items = [{ currentPrice: 100, targetPrice: 150 }]

    vi.useFakeTimers()
    vi.setSystemTime(new Date(currentTime))
    const result = computeMissionProgress(mission, items, 90)
    vi.useRealTimers()

    expect(result.urgency).toBe('low')
  })

  // Overall status
  it('marks overall status as overdue when timePercent > 100', () => {
    const createdTime = new Date(baseTime).toISOString()
    const currentTime = new Date(baseTime + 95 * 24 * 60 * 60 * 1000).toISOString()
    const mission = { budget: 1000, createdAt: createdTime }
    const items = [{ currentPrice: 100, targetPrice: 150 }]

    vi.useFakeTimers()
    vi.setSystemTime(new Date(currentTime))
    const result = computeMissionProgress(mission, items, 90)
    vi.useRealTimers()

    expect(result.overallStatus).toBe('overdue')
  })

  it('marks overall status as at_risk when timePercent > 75 and readinessPercent < 50', () => {
    const createdTime = new Date(baseTime).toISOString()
    const currentTime = new Date(baseTime + 80 * 24 * 60 * 60 * 1000).toISOString()
    const mission = { budget: 1000, createdAt: createdTime }
    const items = [
      { currentPrice: 200, targetPrice: 150 },
      { currentPrice: 300, targetPrice: 250 },
    ]

    vi.useFakeTimers()
    vi.setSystemTime(new Date(currentTime))
    const result = computeMissionProgress(mission, items, 90)
    vi.useRealTimers()

    expect(result.overallStatus).toBe('at_risk')
  })

  it('marks overall status as ahead when readinessPercent > 75', () => {
    const createdTime = new Date(baseTime).toISOString()
    const currentTime = new Date(baseTime + 30 * 24 * 60 * 60 * 1000).toISOString()
    const mission = { budget: 1000, createdAt: createdTime }
    const items = [
      { currentPrice: 100, targetPrice: 150 },
      { currentPrice: 100, targetPrice: 150 },
      { currentPrice: 100, targetPrice: 150 },
    ]

    vi.useFakeTimers()
    vi.setSystemTime(new Date(currentTime))
    const result = computeMissionProgress(mission, items, 90)
    vi.useRealTimers()

    expect(result.overallStatus).toBe('ahead')
  })

  it('marks overall status as on_track for normal cases', () => {
    const createdTime = new Date(baseTime).toISOString()
    const currentTime = new Date(baseTime + 30 * 24 * 60 * 60 * 1000).toISOString()
    const mission = { budget: 1000, createdAt: createdTime }
    const items = [
      { currentPrice: 100, targetPrice: 150 }, // ready
      { currentPrice: 200, targetPrice: 150 }, // not ready (200 > 150)
    ]

    vi.useFakeTimers()
    vi.setSystemTime(new Date(currentTime))
    const result = computeMissionProgress(mission, items, 90)
    vi.useRealTimers()

    expect(result.overallStatus).toBe('on_track')
  })

  // Edge cases
  it('handles no items', () => {
    const mission = { budget: 1000, createdAt: new Date(baseTime).toISOString() }
    const items: never[] = []
    const result = computeMissionProgress(mission, items)
    expect(result.totalItems).toBe(0)
    expect(result.readyItems).toBe(0)
    expect(result.readinessPercent).toBe(100) // 0/0 = ready
  })

  it('handles no budget', () => {
    const mission = { budget: 0, createdAt: new Date(baseTime).toISOString() }
    const items = [{ currentPrice: 100, targetPrice: 150 }]
    const result = computeMissionProgress(mission, items)
    expect(result.budgetTotal).toBe(0)
    expect(result.budgetPercent).toBe(100) // over infinite budget treated as 100%
  })

  it('handles null/undefined budget', () => {
    const mission = { budget: undefined, createdAt: new Date(baseTime).toISOString() }
    const items = [{ currentPrice: 100, targetPrice: 150 }]
    const result = computeMissionProgress(mission, items)
    expect(result.budgetTotal).toBe(0)
  })
})

// Import vitest for vi.useFakeTimers
import { vi } from 'vitest'
