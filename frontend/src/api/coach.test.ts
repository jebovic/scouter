import { describe, it, expect, vi, beforeEach } from 'vitest'
import { CoachResponseSchema, CoachTipSchema, fetchCoachTips } from './coach'

describe('Coach API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('CoachTipSchema', () => {
    it('parses a valid coach tip', () => {
      const tip = {
        id: 'abc123',
        category: 'research' as const,
        priority: 'high' as const,
        title: 'Expand your options',
        body: 'Consider more brands for comparison',
        action: 'Launch research',
      }
      const result = CoachTipSchema.safeParse(tip)
      expect(result.success).toBe(true)
    })

    it('validates category enum', () => {
      const tip = {
        id: 'abc123',
        category: 'invalid',
        priority: 'high',
        title: 'Test',
        body: 'Test body',
      }
      const result = CoachTipSchema.safeParse(tip)
      expect(result.success).toBe(false)
    })

    it('validates priority enum', () => {
      const tip = {
        id: 'abc123',
        category: 'research',
        priority: 'urgent',
        title: 'Test',
        body: 'Test body',
      }
      const result = CoachTipSchema.safeParse(tip)
      expect(result.success).toBe(false)
    })

    it('makes action optional', () => {
      const tip = {
        id: 'abc123',
        category: 'budget',
        priority: 'medium',
        title: 'Set budget alerts',
        body: 'Monitor spending',
      }
      const result = CoachTipSchema.safeParse(tip)
      expect(result.success).toBe(true)
    })
  })

  describe('CoachResponseSchema', () => {
    it('parses a valid coach response', () => {
      const response = {
        missionId: 'mission-123',
        tips: [
          {
            id: 'tip1',
            category: 'research' as const,
            priority: 'high' as const,
            title: 'Research more',
            body: 'Consider additional options',
          },
        ],
        generatedAt: new Date().toISOString(),
      }
      const result = CoachResponseSchema.safeParse(response)
      expect(result.success).toBe(true)
    })

    it('allows empty tips array', () => {
      const response = {
        missionId: 'mission-123',
        tips: [],
        generatedAt: new Date().toISOString(),
      }
      const result = CoachResponseSchema.safeParse(response)
      expect(result.success).toBe(true)
    })
  })
})
