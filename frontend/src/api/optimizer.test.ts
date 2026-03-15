import { describe, it, expect } from 'vitest'
import { OptimizedPlanSchema, ShoppingActionSchema } from './optimizer'

describe('optimizer API schemas', () => {
  describe('ShoppingActionSchema', () => {
    it('parses a valid shopping action', () => {
      const action = {
        priority: 1,
        itemIds: ['abc-123'],
        itemNames: ['Laptop'],
        merchant: 'Amazon',
        action: 'Acheter maintenant',
        reason: 'Meilleur prix actuel',
        estSavings: 15.5,
        currency: 'EUR',
      }
      const result = ShoppingActionSchema.safeParse(action)
      expect(result.success).toBe(true)
    })

    it('rejects action with missing required fields', () => {
      const action = {
        priority: 1,
        itemIds: ['abc'],
        // missing itemNames, merchant, action, reason, estSavings, currency
      }
      const result = ShoppingActionSchema.safeParse(action)
      expect(result.success).toBe(false)
    })

    it('accepts estSavings of zero', () => {
      const action = {
        priority: 2,
        itemIds: ['xyz'],
        itemNames: ['Souris'],
        merchant: 'Fnac',
        action: 'Comparer en magasin',
        reason: 'Prix incertain',
        estSavings: 0,
        currency: 'EUR',
      }
      expect(ShoppingActionSchema.safeParse(action).success).toBe(true)
    })
  })

  describe('OptimizedPlanSchema', () => {
    const validPlan = {
      missionId: 'mission-uuid-123',
      generatedAt: new Date().toISOString(),
      totalItems: 2,
      actions: [
        {
          priority: 1,
          itemIds: ['id-1'],
          itemNames: ['Laptop'],
          merchant: 'Amazon',
          action: 'Acheter maintenant',
          reason: 'Prix bas cette semaine',
          estSavings: 20,
          currency: 'EUR',
        },
      ],
      summary: 'Commencez par acheter le laptop.',
    }

    it('parses a valid optimized plan', () => {
      const result = OptimizedPlanSchema.safeParse(validPlan)
      expect(result.success).toBe(true)
    })

    it('rejects plan missing summary field', () => {
      const { summary: _omitted, ...withoutSummary } = validPlan
      const result = OptimizedPlanSchema.safeParse(withoutSummary)
      expect(result.success).toBe(false)
    })

    it('rejects plan with missing missionId', () => {
      const { missionId: _omitted, ...withoutId } = validPlan
      const result = OptimizedPlanSchema.safeParse(withoutId)
      expect(result.success).toBe(false)
    })

    it('accepts empty actions array', () => {
      const plan = { ...validPlan, actions: [] }
      expect(OptimizedPlanSchema.safeParse(plan).success).toBe(true)
    })

    it('rejects plan with invalid action inside actions array', () => {
      const plan = {
        ...validPlan,
        actions: [{ priority: 'not-a-number' }], // wrong type
      }
      expect(OptimizedPlanSchema.safeParse(plan).success).toBe(false)
    })
  })
})
