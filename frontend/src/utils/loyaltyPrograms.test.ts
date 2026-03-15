import { describe, it, expect } from 'vitest'
import {
  LOYALTY_PROGRAMS,
  getNextMilestone,
  getEquivalentValue,
} from './loyaltyPrograms'

describe('loyaltyPrograms', () => {
  describe('LOYALTY_PROGRAMS', () => {
    it('should have 5 programs', () => {
      expect(LOYALTY_PROGRAMS).toHaveLength(5)
    })

    it('should have required fields for each program', () => {
      LOYALTY_PROGRAMS.forEach((program) => {
        expect(program).toHaveProperty('id')
        expect(program).toHaveProperty('name')
        expect(program).toHaveProperty('color')
        expect(program).toHaveProperty('icon')
        expect(program).toHaveProperty('pointsPerEuro')
        expect(program).toHaveProperty('eurPerPoint')
        expect(program).toHaveProperty('milestones')
      })
    })

    it('should have milestones array for each program', () => {
      LOYALTY_PROGRAMS.forEach((program) => {
        expect(Array.isArray(program.milestones)).toBe(true)
        expect(program.milestones.length).toBeGreaterThan(0)
        program.milestones.forEach((milestone) => {
          expect(milestone).toHaveProperty('points')
          expect(milestone).toHaveProperty('reward')
          expect(typeof milestone.points).toBe('number')
          expect(typeof milestone.reward).toBe('string')
        })
      })
    })
  })

  describe('getNextMilestone', () => {
    it('should return the first milestone above current points', () => {
      const fnac = LOYALTY_PROGRAMS.find((p) => p.id === 'fnac')!
      const nextMilestone = getNextMilestone(fnac, 250)
      expect(nextMilestone).not.toBeNull()
      expect(nextMilestone?.points).toBe(500)
      expect(nextMilestone?.reward).toBe('5€ de réduction')
    })

    it('should return next milestone when at intermediate level', () => {
      const fnac = LOYALTY_PROGRAMS.find((p) => p.id === 'fnac')!
      const nextMilestone = getNextMilestone(fnac, 750)
      expect(nextMilestone).not.toBeNull()
      expect(nextMilestone?.points).toBe(1000)
    })

    it('should return null when at maximum milestone', () => {
      const fnac = LOYALTY_PROGRAMS.find((p) => p.id === 'fnac')!
      const nextMilestone = getNextMilestone(fnac, 2000)
      expect(nextMilestone).toBeNull()
    })

    it('should return null when above all milestones', () => {
      const fnac = LOYALTY_PROGRAMS.find((p) => p.id === 'fnac')!
      const nextMilestone = getNextMilestone(fnac, 5000)
      expect(nextMilestone).toBeNull()
    })

    it('should return first milestone when at 0 points', () => {
      const cdiscount = LOYALTY_PROGRAMS.find((p) => p.id === 'cdiscount')!
      const nextMilestone = getNextMilestone(cdiscount, 0)
      expect(nextMilestone).not.toBeNull()
      expect(nextMilestone?.points).toBe(1000)
    })
  })

  describe('getEquivalentValue', () => {
    it('should calculate correct euro value for Fnac', () => {
      const fnac = LOYALTY_PROGRAMS.find((p) => p.id === 'fnac')!
      const value = getEquivalentValue(fnac, 1000)
      expect(value).toBe(10)
    })

    it('should calculate correct euro value for Cdiscount', () => {
      const cdiscount = LOYALTY_PROGRAMS.find((p) => p.id === 'cdiscount')!
      const value = getEquivalentValue(cdiscount, 1000)
      expect(value).toBe(5)
    })

    it('should return 0 for 0 points', () => {
      const fnac = LOYALTY_PROGRAMS.find((p) => p.id === 'fnac')!
      const value = getEquivalentValue(fnac, 0)
      expect(value).toBe(0)
    })

    it('should handle decimal values correctly', () => {
      const fnac = LOYALTY_PROGRAMS.find((p) => p.id === 'fnac')!
      const value = getEquivalentValue(fnac, 1240)
      expect(value).toBe(12.4)
    })
  })
})
