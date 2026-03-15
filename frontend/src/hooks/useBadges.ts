import { useState, useCallback } from 'react'
import { ALL_BADGES, type Badge, type BadgeId } from '../utils/badges'

const STORAGE_KEY = 'scouter_badges'

interface StoredBadge {
  id: BadgeId
  unlockedAt: string
}

function readBadgesFromStorage(): Badge[] {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (!stored) return []
    const parsed: StoredBadge[] = JSON.parse(stored)
    return parsed.map((b) => ({
      ...ALL_BADGES[b.id],
      unlockedAt: new Date(b.unlockedAt),
    }))
  } catch {
    return []
  }
}

function writeBadgesToStorage(badges: Badge[]): void {
  try {
    const toStore: StoredBadge[] = badges
      .filter((b) => b.unlockedAt)
      .map((b) => ({
        id: b.id,
        unlockedAt: b.unlockedAt!.toISOString(),
      }))
    localStorage.setItem(STORAGE_KEY, JSON.stringify(toStore))
  } catch {
    // ignore private browsing errors
  }
}

export interface UseBadgesReturn {
  badges: Badge[]
  unlock: (id: BadgeId) => void
  hasUnlocked: (id: BadgeId) => boolean
}

export function useBadges(): UseBadgesReturn {
  const [badges, setBadges] = useState<Badge[]>(() => readBadgesFromStorage())

  const unlock = useCallback((id: BadgeId) => {
    setBadges((prev) => {
      // Check if already unlocked (immutable pattern)
      if (prev.some((b) => b.id === id)) {
        return prev
      }
      // Create new badge with unlock timestamp
      const newBadge: Badge = {
        ...ALL_BADGES[id],
        unlockedAt: new Date(),
      }
      // Create new array with the new badge
      const updated = [...prev, newBadge]
      // Persist to localStorage
      writeBadgesToStorage(updated)
      return updated
    })
  }, [])

  const hasUnlocked = useCallback(
    (id: BadgeId): boolean => {
      return badges.some((b) => b.id === id)
    },
    [badges]
  )

  return { badges, unlock, hasUnlocked }
}
