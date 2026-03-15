import { describe, it, expect } from 'vitest'
import { SubstituteSchema, SubstituteResponseSchema } from './substitutes'

const validSubstitute = {
  name: 'Samsung Galaxy S23',
  brand: 'Samsung',
  price: 799.0,
  currency: 'EUR',
  retailer: 'Fnac',
  reason: 'Comparable specs at a lower price',
  advantage: 'cheaper' as const,
  url: 'https://fnac.com/samsung-s23',
}

const validResponse = {
  optionId: 'opt-111',
  productName: 'iPhone 15 Pro',
  substitutes: [validSubstitute],
  generatedAt: new Date().toISOString(),
}

describe('SubstituteSchema', () => {
  it('accepts a valid substitute with all fields', () => {
    expect(() => SubstituteSchema.parse(validSubstitute)).not.toThrow()
    const result = SubstituteSchema.parse(validSubstitute)
    expect(result.name).toBe('Samsung Galaxy S23')
    expect(result.advantage).toBe('cheaper')
  })

  it('accepts substitute without optional url', () => {
    const { url: _omit, ...withoutUrl } = validSubstitute
    expect(() => SubstituteSchema.parse(withoutUrl)).not.toThrow()
    const result = SubstituteSchema.parse(withoutUrl)
    expect(result.url).toBeUndefined()
  })

  it('accepts all valid advantage values', () => {
    const advantages = ['cheaper', 'better_rated', 'eco_friendly', 'local'] as const
    for (const advantage of advantages) {
      expect(() => SubstituteSchema.parse({ ...validSubstitute, advantage })).not.toThrow()
    }
  })

  it('rejects unknown advantage value', () => {
    expect(() =>
      SubstituteSchema.parse({ ...validSubstitute, advantage: 'unknown' }),
    ).toThrow()
  })

  it('rejects missing required fields', () => {
    const { name: _omit, ...withoutName } = validSubstitute
    expect(() => SubstituteSchema.parse(withoutName)).toThrow()
  })

  // Phase 79 enrichment fields
  it('accepts Phase 79 enrichment fields: savingsPercent, pros, cons, whyConsider', () => {
    const enriched = {
      ...validSubstitute,
      savingsPercent: 46,
      pros: ['Autonomie 60h', 'ANC efficace'],
      cons: ['Son moins précis'],
      whyConsider: 'La meilleure alternative budget.',
    }
    expect(() => SubstituteSchema.parse(enriched)).not.toThrow()
    const result = SubstituteSchema.parse(enriched)
    expect(result.savingsPercent).toBe(46)
    expect(result.pros).toHaveLength(2)
    expect(result.cons).toHaveLength(1)
    expect(result.whyConsider).toBe('La meilleure alternative budget.')
  })

  it('accepts substitute without Phase 79 enrichment fields (backward compatible)', () => {
    expect(() => SubstituteSchema.parse(validSubstitute)).not.toThrow()
    const result = SubstituteSchema.parse(validSubstitute)
    expect(result.savingsPercent).toBeUndefined()
    expect(result.pros).toBeUndefined()
    expect(result.cons).toBeUndefined()
    expect(result.whyConsider).toBeUndefined()
  })

  it('savingsPercent zero is accepted', () => {
    const s = { ...validSubstitute, savingsPercent: 0 }
    expect(() => SubstituteSchema.parse(s)).not.toThrow()
    expect(SubstituteSchema.parse(s).savingsPercent).toBe(0)
  })
})

describe('SubstituteResponseSchema', () => {
  it('accepts a valid response payload', () => {
    expect(() => SubstituteResponseSchema.parse(validResponse)).not.toThrow()
    const result = SubstituteResponseSchema.parse(validResponse)
    expect(result.optionId).toBe('opt-111')
    expect(result.productName).toBe('iPhone 15 Pro')
    expect(result.substitutes).toHaveLength(1)
  })

  it('accepts response with empty substitutes array', () => {
    const empty = { ...validResponse, substitutes: [] }
    expect(() => SubstituteResponseSchema.parse(empty)).not.toThrow()
    const result = SubstituteResponseSchema.parse(empty)
    expect(result.substitutes).toHaveLength(0)
  })

  it('accepts response with Phase 79 fields: originalPrice and cachedAt', () => {
    const enriched = { ...validResponse, originalPrice: 1000.0, cachedAt: 1700000000 }
    expect(() => SubstituteResponseSchema.parse(enriched)).not.toThrow()
    const result = SubstituteResponseSchema.parse(enriched)
    expect(result.originalPrice).toBe(1000.0)
    expect(result.cachedAt).toBe(1700000000)
  })

  it('rejects response missing optionId', () => {
    const { optionId: _omit, ...rest } = validResponse
    expect(() => SubstituteResponseSchema.parse(rest)).toThrow()
  })

  it('rejects response missing productName', () => {
    const { productName: _omit, ...rest } = validResponse
    expect(() => SubstituteResponseSchema.parse(rest)).toThrow()
  })

  it('rejects response with invalid substitute in array', () => {
    const bad = {
      ...validResponse,
      substitutes: [{ ...validSubstitute, advantage: 'not-valid' }],
    }
    expect(() => SubstituteResponseSchema.parse(bad)).toThrow()
  })
})
