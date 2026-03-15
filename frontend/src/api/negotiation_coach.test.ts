import { describe, it, expect } from 'vitest'
import { NegotiationTipSchema, NegotiationAdviceSchema } from './negotiation_coach'

const validTip = {
  title: 'Demander le prix affiché en magasin',
  description: 'Montrez les offres concurrentes sur votre téléphone et demandez poliment l\'alignement de prix.',
  difficulty: 'easy' as const,
}

const validAdvice = {
  itemId: 'item-abc-123',
  itemName: 'iPhone 15 Pro',
  currentPrice: 1299.0,
  targetPrice: 1299.0 * 0.87, // 1130.13
  tips: [
    validTip,
    { title: 'T2', description: 'D2', difficulty: 'medium' as const },
    { title: 'T3', description: 'D3', difficulty: 'hard' as const },
    { title: 'T4', description: 'D4', difficulty: 'easy' as const },
  ],
  script: 'Bonjour, j\'ai vu ce produit à 1129€ chez [concurrent]. Pouvez-vous vous aligner sur ce prix ?',
  cachedAt: 1700000000,
}

describe('NegotiationTipSchema', () => {
  it('accepts a valid tip with easy difficulty', () => {
    expect(() => NegotiationTipSchema.parse(validTip)).not.toThrow()
    const result = NegotiationTipSchema.parse(validTip)
    expect(result.difficulty).toBe('easy')
    expect(result.title).toBe('Demander le prix affiché en magasin')
  })

  it('accepts medium and hard difficulty values', () => {
    const medium = { ...validTip, difficulty: 'medium' as const }
    const hard = { ...validTip, difficulty: 'hard' as const }
    expect(() => NegotiationTipSchema.parse(medium)).not.toThrow()
    expect(() => NegotiationTipSchema.parse(hard)).not.toThrow()
  })

  it('rejects unknown difficulty value', () => {
    expect(() =>
      NegotiationTipSchema.parse({ ...validTip, difficulty: 'very-hard' }),
    ).toThrow()
  })

  it('rejects missing title field', () => {
    const { title: _omit, ...without } = validTip
    expect(() => NegotiationTipSchema.parse(without)).toThrow()
  })

  it('rejects non-string description', () => {
    expect(() =>
      NegotiationTipSchema.parse({ ...validTip, description: 42 }),
    ).toThrow()
  })
})

describe('NegotiationAdviceSchema', () => {
  it('accepts a valid negotiation advice', () => {
    expect(() => NegotiationAdviceSchema.parse(validAdvice)).not.toThrow()
    const result = NegotiationAdviceSchema.parse(validAdvice)
    expect(result.itemId).toBe('item-abc-123')
    expect(result.targetPrice).toBeLessThan(result.currentPrice)
    expect(result.tips).toHaveLength(4)
  })

  it('targetPrice is less than currentPrice (13% reduction)', () => {
    const result = NegotiationAdviceSchema.parse(validAdvice)
    const expectedTarget = result.currentPrice * 0.87
    expect(result.targetPrice).toBeCloseTo(expectedTarget, 0)
  })

  it('rejects missing script field', () => {
    const { script: _omit, ...without } = validAdvice
    expect(() => NegotiationAdviceSchema.parse(without)).toThrow()
  })

  it('rejects non-array tips', () => {
    expect(() =>
      NegotiationAdviceSchema.parse({ ...validAdvice, tips: 'not-an-array' }),
    ).toThrow()
  })

  it('rejects tips with invalid difficulty inside array', () => {
    const badTips = [{ title: 'T', description: 'D', difficulty: 'extreme' }]
    expect(() =>
      NegotiationAdviceSchema.parse({ ...validAdvice, tips: badTips }),
    ).toThrow()
  })
})
