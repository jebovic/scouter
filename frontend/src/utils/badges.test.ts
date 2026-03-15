import { describe, it, expect } from 'vitest'
import { ALL_BADGES, unlockIfEarned, type BadgeId } from './badges'

describe('badges.ts', () => {
  describe('ALL_BADGES', () => {
    it('defines all badge IDs as keys', () => {
      const expectedIds: BadgeId[] = [
        'first_mission',
        'mission_completed',
        'five_missions',
        'saver_bronze',
        'saver_silver',
        'saver_gold',
        'researcher',
        'negotiator',
        'deal_hunter',
        'planner',
      ]
      expectedIds.forEach((id) => {
        expect(ALL_BADGES).toHaveProperty(id)
      })
    })

    it('has exactly 10 badges', () => {
      expect(Object.keys(ALL_BADGES).length).toBe(10)
    })

    it('each badge has id, name, description, icon', () => {
      Object.entries(ALL_BADGES).forEach(([key, badge]) => {
        expect(badge.id).toBe(key)
        expect(typeof badge.name).toBe('string')
        expect(badge.name.length).toBeGreaterThan(0)
        expect(typeof badge.description).toBe('string')
        expect(badge.description.length).toBeGreaterThan(0)
        expect(typeof badge.icon).toBe('string')
        expect(badge.icon.length).toBeGreaterThan(0)
      })
    })

    it('all badge names are in French', () => {
      // Check that at least some badges have French names (accents, etc.)
      const frenchBadges = Object.values(ALL_BADGES).filter((b) => /[éàâêôûüœæù]/.test(b.name))
      expect(frenchBadges.length).toBeGreaterThan(0)
    })
  })

  describe('unlockIfEarned', () => {
    it('unlocks first_mission when missionCount >= 1', () => {
      const badges = unlockIfEarned(1, 0, 0)
      expect(badges).toContain('first_mission')
    })

    it('does not unlock first_mission when missionCount = 0', () => {
      const badges = unlockIfEarned(0, 0, 0)
      expect(badges).not.toContain('first_mission')
    })

    it('unlocks mission_completed when missionCount >= 1', () => {
      const badges = unlockIfEarned(1, 0, 0)
      expect(badges).toContain('mission_completed')
    })

    it('unlocks five_missions when missionCount >= 5', () => {
      const badges = unlockIfEarned(5, 0, 0)
      expect(badges).toContain('five_missions')
    })

    it('does not unlock five_missions when missionCount = 4', () => {
      const badges = unlockIfEarned(4, 0, 0)
      expect(badges).not.toContain('five_missions')
    })

    it('unlocks saver_bronze when savedTotal >= 100', () => {
      const badges = unlockIfEarned(0, 100, 0)
      expect(badges).toContain('saver_bronze')
    })

    it('does not unlock saver_bronze when savedTotal < 100', () => {
      const badges = unlockIfEarned(0, 99, 0)
      expect(badges).not.toContain('saver_bronze')
    })

    it('unlocks saver_silver when savedTotal >= 500', () => {
      const badges = unlockIfEarned(0, 500, 0)
      expect(badges).toContain('saver_silver')
    })

    it('does not unlock saver_silver when savedTotal < 500', () => {
      const badges = unlockIfEarned(0, 499, 0)
      expect(badges).not.toContain('saver_silver')
    })

    it('unlocks saver_gold when savedTotal >= 1000', () => {
      const badges = unlockIfEarned(0, 1000, 0)
      expect(badges).toContain('saver_gold')
    })

    it('does not unlock saver_gold when savedTotal < 1000', () => {
      const badges = unlockIfEarned(0, 999, 0)
      expect(badges).not.toContain('saver_gold')
    })

    it('unlocks researcher when researchCount >= 10', () => {
      const badges = unlockIfEarned(0, 0, 10)
      expect(badges).toContain('researcher')
    })

    it('does not unlock researcher when researchCount < 10', () => {
      const badges = unlockIfEarned(0, 0, 9)
      expect(badges).not.toContain('researcher')
    })

    it('returns array of earned BadgeIds', () => {
      const badges = unlockIfEarned(5, 1000, 10)
      expect(Array.isArray(badges)).toBe(true)
      expect(badges.every((b) => typeof b === 'string')).toBe(true)
    })

    it('handles zero inputs', () => {
      const badges = unlockIfEarned(0, 0, 0)
      expect(Array.isArray(badges)).toBe(true)
      expect(badges.length).toBe(0)
    })

    it('handles high values without error', () => {
      const badges = unlockIfEarned(1000, 100000, 1000)
      expect(Array.isArray(badges)).toBe(true)
      expect(badges.length).toBeGreaterThan(0)
    })
  })
})
