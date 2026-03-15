import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useBadges } from './useBadges'
import type { BadgeId } from '../utils/badges'

const STORAGE_KEY = 'scouter_badges'

describe('useBadges', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('returns empty badges array when localStorage is empty', () => {
    const { result } = renderHook(() => useBadges())
    expect(result.current.badges).toEqual([])
  })

  it('reads badges from localStorage on mount', () => {
    const badgeData = [
      { id: 'first_mission' as BadgeId, unlockedAt: '2026-03-15T10:00:00Z' },
    ]
    localStorage.setItem(STORAGE_KEY, JSON.stringify(badgeData))
    const { result } = renderHook(() => useBadges())
    expect(result.current.badges).toHaveLength(1)
    expect(result.current.badges[0].id).toBe('first_mission')
  })

  it('unlock adds new badge to localStorage', () => {
    const { result } = renderHook(() => useBadges())
    act(() => {
      result.current.unlock('first_mission')
    })
    const stored = localStorage.getItem(STORAGE_KEY)
    expect(stored).toBeTruthy()
    const parsed = JSON.parse(stored!)
    expect(parsed.some((b: any) => b.id === 'first_mission')).toBe(true)
  })

  it('unlock returns new array (immutable)', () => {
    const { result: result1 } = renderHook(() => useBadges())
    const firstBadges = result1.current.badges
    act(() => {
      result1.current.unlock('first_mission')
    })
    const secondBadges = result1.current.badges
    expect(firstBadges).not.toBe(secondBadges)
  })

  it('unlock idempotency: unlocking same badge twice does not duplicate', () => {
    const { result } = renderHook(() => useBadges())
    act(() => {
      result.current.unlock('first_mission')
    })
    const firstCount = result.current.badges.length
    act(() => {
      result.current.unlock('first_mission')
    })
    const secondCount = result.current.badges.length
    expect(firstCount).toBe(secondCount)
  })

  it('unlock sets unlockedAt timestamp', () => {
    const { result } = renderHook(() => useBadges())
    const before = new Date().getTime()
    act(() => {
      result.current.unlock('first_mission')
    })
    const after = new Date().getTime()
    const badge = result.current.badges.find((b) => b.id === 'first_mission')
    expect(badge?.unlockedAt).toBeTruthy()
    const unlockedTime = badge!.unlockedAt!.getTime()
    expect(unlockedTime).toBeGreaterThanOrEqual(before)
    expect(unlockedTime).toBeLessThanOrEqual(after)
  })

  it('hasUnlocked returns true for unlocked badge', () => {
    const { result } = renderHook(() => useBadges())
    act(() => {
      result.current.unlock('first_mission')
    })
    expect(result.current.hasUnlocked('first_mission')).toBe(true)
  })

  it('hasUnlocked returns false for locked badge', () => {
    const { result } = renderHook(() => useBadges())
    expect(result.current.hasUnlocked('first_mission')).toBe(false)
  })

  it('unlock with multiple badges keeps all of them', () => {
    const { result } = renderHook(() => useBadges())
    act(() => {
      result.current.unlock('first_mission')
      result.current.unlock('five_missions')
      result.current.unlock('saver_bronze')
    })
    expect(result.current.badges).toHaveLength(3)
    expect(result.current.hasUnlocked('first_mission')).toBe(true)
    expect(result.current.hasUnlocked('five_missions')).toBe(true)
    expect(result.current.hasUnlocked('saver_bronze')).toBe(true)
  })

  it('handles localStorage errors gracefully', () => {
    const spy = vi.spyOn(Storage.prototype, 'getItem').mockImplementationOnce(() => {
      throw new Error('QuotaExceededError')
    })
    const { result } = renderHook(() => useBadges())
    expect(result.current.badges).toEqual([])
    spy.mockRestore()
  })
})
