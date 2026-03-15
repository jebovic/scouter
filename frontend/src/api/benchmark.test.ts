import { describe, it, expect } from 'vitest'
import { PriceRangeSchema, BenchmarkResultSchema } from './benchmark'

const validRange = {
  low: 299.0,
  average: 349.0,
  high: 399.0,
}

const validResult = {
  itemId: 'item-abc-123',
  itemName: 'Sony WH-1000XM5',
  currentPrice: 279.99,
  currency: 'EUR',
  marketRange: validRange,
  verdict: 'good_deal' as const,
  diffPercent: -19.77,
  explanation: "Ce prix est 20% en dessous de la moyenne du marché français.",
  cachedAt: 1700000000,
}

describe('PriceRangeSchema', () => {
  it('accepts a valid price range', () => {
    expect(() => PriceRangeSchema.parse(validRange)).not.toThrow()
    const result = PriceRangeSchema.parse(validRange)
    expect(result.low).toBe(299.0)
    expect(result.average).toBe(349.0)
    expect(result.high).toBe(399.0)
  })

  it('rejects missing low field', () => {
    const { low: _omit, ...without } = validRange
    expect(() => PriceRangeSchema.parse(without)).toThrow()
  })

  it('rejects non-numeric average', () => {
    expect(() => PriceRangeSchema.parse({ ...validRange, average: 'not-a-number' })).toThrow()
  })
})

describe('BenchmarkResultSchema', () => {
  it('accepts a valid good_deal result', () => {
    expect(() => BenchmarkResultSchema.parse(validResult)).not.toThrow()
    const result = BenchmarkResultSchema.parse(validResult)
    expect(result.itemId).toBe('item-abc-123')
    expect(result.verdict).toBe('good_deal')
    expect(result.diffPercent).toBeLessThan(0)
  })

  it('accepts overpriced verdict', () => {
    const overpriced = { ...validResult, verdict: 'overpriced' as const, diffPercent: 28.5 }
    expect(() => BenchmarkResultSchema.parse(overpriced)).not.toThrow()
    const result = BenchmarkResultSchema.parse(overpriced)
    expect(result.verdict).toBe('overpriced')
  })

  it('accepts average verdict', () => {
    const average = { ...validResult, verdict: 'average' as const, diffPercent: 4.2 }
    expect(() => BenchmarkResultSchema.parse(average)).not.toThrow()
    const result = BenchmarkResultSchema.parse(average)
    expect(result.verdict).toBe('average')
  })

  it('rejects unknown verdict value', () => {
    expect(() =>
      BenchmarkResultSchema.parse({ ...validResult, verdict: 'unknown_verdict' }),
    ).toThrow()
  })

  it('rejects missing required itemName field', () => {
    const { itemName: _omit, ...without } = validResult
    expect(() => BenchmarkResultSchema.parse(without)).toThrow()
  })
})
