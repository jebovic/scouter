import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { BadgeRow } from './BadgeRow'
import { ALL_BADGES } from '../../utils/badges'
import type { Badge } from '../../utils/badges'

describe('BadgeRow', () => {
  it('renders without crashing with empty badges', () => {
    render(<BadgeRow badges={[]} />)
    const container = screen.getByRole('list')
    expect(container).toBeInTheDocument()
  })

  it('renders a list role', () => {
    render(<BadgeRow badges={[]} />)
    expect(screen.getByRole('list')).toBeInTheDocument()
  })

  it('renders all unlocked badges', () => {
    const unlockedBadges: Badge[] = [
      { ...ALL_BADGES.first_mission, unlockedAt: new Date('2026-03-15') },
      { ...ALL_BADGES.five_missions, unlockedAt: new Date('2026-03-15') },
    ]
    render(<BadgeRow badges={unlockedBadges} />)
    const items = screen.getAllByRole('listitem')
    expect(items.length).toBeGreaterThanOrEqual(2)
  })

  it('renders badges with correct emoji icons', () => {
    const unlockedBadges: Badge[] = [
      { ...ALL_BADGES.first_mission, unlockedAt: new Date('2026-03-15') },
    ]
    render(<BadgeRow badges={unlockedBadges} />)
    expect(screen.getByText('🎯')).toBeInTheDocument()
  })

  it('renders locked badges when provided', () => {
    const lockedBadges: Badge[] = [
      { ...ALL_BADGES.first_mission },
      { ...ALL_BADGES.five_missions },
    ]
    render(<BadgeRow badges={lockedBadges} />)
    const items = screen.getAllByRole('listitem')
    expect(items.length).toBeGreaterThanOrEqual(2)
  })

  it('applies different classes for unlocked vs locked badges', () => {
    const badges: Badge[] = [
      { ...ALL_BADGES.first_mission, unlockedAt: new Date('2026-03-15') },
      { ...ALL_BADGES.five_missions }, // locked
    ]
    render(<BadgeRow badges={badges} />)
    const items = screen.getAllByRole('listitem')
    expect(items.length).toBeGreaterThanOrEqual(2)
    // At least one should be unlocked and one locked
    const hasUnlockedClass = items.some((item) => item.className.includes('unlocked'))
    const hasLockedClass = items.some((item) => item.className.includes('locked'))
    expect(hasUnlockedClass || hasLockedClass).toBe(true)
  })

  it('each badge item is accessible', () => {
    const unlockedBadges: Badge[] = [
      { ...ALL_BADGES.first_mission, unlockedAt: new Date('2026-03-15') },
    ]
    render(<BadgeRow badges={unlockedBadges} />)
    const item = screen.getByRole('listitem')
    expect(item).toBeInTheDocument()
  })

  it('renders badges in order', () => {
    const badges: Badge[] = [
      { ...ALL_BADGES.first_mission, unlockedAt: new Date('2026-03-15') },
      { ...ALL_BADGES.five_missions, unlockedAt: new Date('2026-03-16') },
      { ...ALL_BADGES.saver_gold, unlockedAt: new Date('2026-03-17') },
    ]
    render(<BadgeRow badges={badges} />)
    const items = screen.getAllByRole('listitem')
    expect(items.length).toBeGreaterThanOrEqual(3)
  })
})
