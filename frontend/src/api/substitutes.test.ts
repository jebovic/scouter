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
