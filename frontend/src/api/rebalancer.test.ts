import { describe, it, expect } from 'vitest'
import { EnvelopeAdjustmentSchema, RebalancePlanSchema } from './rebalancer'

const validAdjustment = {
  envelopeId: 'a1b2c3d4-0000-0000-0000-000000000001',
  envelopeName: 'Alimentation',
  currentMonthly: 250,
  suggestedMonthly: 320,
  deltaPercent: 28,
  reason: "Vous dépensez régulièrement 320€ en alimentation mais n'avez alloué que 250€.",
}

const validPlan = {
  adjustments: [validAdjustment],
  summary: "Votre budget actuel sous-estime vos dépenses alimentaires de 28%.",
  totalCurrent: 250,
  totalSuggested: 320,
  cachedAt: 1710000000,
}

describe('EnvelopeAdjustmentSchema', () => {
  it('parses a valid adjustment', () => {
    const result = EnvelopeAdjustmentSchema.safeParse(validAdjustment)
    expect(result.success).toBe(true)
  })

  it('rejects adjustment with missing envelopeId', () => {
    const { envelopeId: _omitted, ...without } = validAdjustment
    expect(EnvelopeAdjustmentSchema.safeParse(without).success).toBe(false)
  })

  it('rejects adjustment where deltaPercent is not a number', () => {
    const bad = { ...validAdjustment, deltaPercent: 'beaucoup' }
    expect(EnvelopeAdjustmentSchema.safeParse(bad).success).toBe(false)
  })

  it('accepts deltaPercent of zero (optimal allocation)', () => {
    const optimal = { ...validAdjustment, deltaPercent: 0, suggestedMonthly: 250 }
    expect(EnvelopeAdjustmentSchema.safeParse(optimal).success).toBe(true)
  })

  it('accepts negative deltaPercent (decrease suggestion)', () => {
    const decrease = { ...validAdjustment, deltaPercent: -15, suggestedMonthly: 212.5 }
    expect(EnvelopeAdjustmentSchema.safeParse(decrease).success).toBe(true)
  })
})

describe('RebalancePlanSchema', () => {
  it('parses a valid plan', () => {
    const result = RebalancePlanSchema.safeParse(validPlan)
    expect(result.success).toBe(true)
  })

  it('parses a plan with empty adjustments array', () => {
    const emptyPlan = { ...validPlan, adjustments: [] }
    expect(RebalancePlanSchema.safeParse(emptyPlan).success).toBe(true)
  })

  it('rejects plan missing summary', () => {
    const { summary: _omitted, ...without } = validPlan
    expect(RebalancePlanSchema.safeParse(without).success).toBe(false)
  })

  it('rejects plan with invalid adjustment inside adjustments array', () => {
    const bad = { ...validPlan, adjustments: [{ envelopeId: 123 }] }
    expect(RebalancePlanSchema.safeParse(bad).success).toBe(false)
  })

  it('rejects plan with non-numeric totalCurrent', () => {
    const bad = { ...validPlan, totalCurrent: 'beaucoup' }
    expect(RebalancePlanSchema.safeParse(bad).success).toBe(false)
  })
})
