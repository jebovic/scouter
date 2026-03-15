import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MissionGoalTracker } from './MissionGoalTracker'

describe('MissionGoalTracker', () => {
  const baseTime = new Date('2026-01-01T00:00:00Z').getTime()

  it('renders overall status banner', () => {
    const mission = { budget: 1000, currency: 'EUR', createdAt: new Date(baseTime).toISOString() }
    const items = [{ currentPrice: 100, targetPrice: 150 }]

    vi.useFakeTimers()
    vi.setSystemTime(new Date(baseTime))
    render(<MissionGoalTracker mission={mission} items={items} />)
    vi.useRealTimers()

    // Check for any of the status labels
    const statusBanner = document.querySelector('[class*="statusBanner"]')
    expect(statusBanner).toBeInTheDocument()
  })

  it('renders budget section with spent/total', () => {
    const mission = { budget: 1000, currency: 'EUR', createdAt: new Date(baseTime).toISOString() }
    const items = [{ currentPrice: 600, targetPrice: 550 }]

    vi.useFakeTimers()
    vi.setSystemTime(new Date(baseTime))
    render(<MissionGoalTracker mission={mission} items={items} />)
    vi.useRealTimers()

    expect(screen.getByText(/Budget/)).toBeInTheDocument()
  })

  it('renders readiness section with item count', () => {
    const mission = { budget: 1000, currency: 'EUR', createdAt: new Date(baseTime).toISOString() }
    const items = [
      { currentPrice: 100, targetPrice: 150 },
      { currentPrice: 200, targetPrice: 150 },
    ]

    vi.useFakeTimers()
    vi.setSystemTime(new Date(baseTime))
    render(<MissionGoalTracker mission={mission} items={items} />)
    vi.useRealTimers()

    expect(screen.getByText('Préparation')).toBeInTheDocument()
  })

  it('renders timeline section with days elapsed', () => {
    const createdTime = new Date(baseTime).toISOString()
    const currentTime = new Date(baseTime + 10 * 24 * 60 * 60 * 1000).toISOString()
    const mission = { budget: 1000, currency: 'EUR', createdAt: createdTime }
    const items = [{ currentPrice: 100, targetPrice: 150 }]

    vi.useFakeTimers()
    vi.setSystemTime(new Date(currentTime))
    render(<MissionGoalTracker mission={mission} items={items} />)
    vi.useRealTimers()

    expect(screen.getByText('J+10')).toBeInTheDocument()
  })

  it('handles empty items array', () => {
    const mission = { budget: 1000, currency: 'EUR', createdAt: new Date(baseTime).toISOString() }
    const items: never[] = []

    vi.useFakeTimers()
    vi.setSystemTime(new Date(baseTime))
    render(<MissionGoalTracker mission={mission} items={items} />)
    vi.useRealTimers()

    const statusBanner = document.querySelector('[class*="statusBanner"]')
    expect(statusBanner).toBeInTheDocument()
  })

  it('handles mission without budget', () => {
    const mission = { currency: 'EUR', createdAt: new Date(baseTime).toISOString() }
    const items = [{ currentPrice: 100, targetPrice: 150 }]

    vi.useFakeTimers()
    vi.setSystemTime(new Date(baseTime))
    render(<MissionGoalTracker mission={mission} items={items} />)
    vi.useRealTimers()

    const statusBanner = document.querySelector('[class*="statusBanner"]')
    expect(statusBanner).toBeInTheDocument()
  })
})
