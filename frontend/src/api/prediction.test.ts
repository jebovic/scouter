import { describe, it, expect } from 'vitest'
import { PricePredictionSchema } from './prediction'

const validPayload = {
  itemId: 'abc-123',
  currentPrice: 99.99,
  predictedPrice: 89.5,
  confidenceLevel: 'high' as const,
  trend: 'falling' as const,
  trendPct: -10.5,
  dataPoints: 7,
  forecastDays: 30,
}

describe('PricePredictionSchema', () => {
  it('accepts a valid prediction payload', () => {
    expect(() => PricePredictionSchema.parse(validPayload)).not.toThrow()
    const result = PricePredictionSchema.parse(validPayload)
    expect(result.itemId).toBe('abc-123')
    expect(result.forecastDays).toBe(30)
  })

  it('accepts all confidence levels', () => {
    for (const level of ['high', 'medium', 'low'] as const) {
      expect(() =>
        PricePredictionSchema.parse({ ...validPayload, confidenceLevel: level }),
      ).not.toThrow()
    }
  })

  it('accepts all trend values', () => {
    for (const trend of ['rising', 'falling', 'stable'] as const) {
      expect(() =>
        PricePredictionSchema.parse({ ...validPayload, trend }),
      ).not.toThrow()
    }
  })

  it('rejects unknown confidenceLevel', () => {
    expect(() =>
      PricePredictionSchema.parse({ ...validPayload, confidenceLevel: 'very-high' }),
    ).toThrow()
  })

  it('rejects unknown trend value', () => {
    expect(() =>
      PricePredictionSchema.parse({ ...validPayload, trend: 'unknown' }),
    ).toThrow()
  })

  it('rejects missing itemId', () => {
    const { itemId: _omit, ...rest } = validPayload
    expect(() => PricePredictionSchema.parse(rest)).toThrow()
  })

  it('rejects non-numeric predictedPrice', () => {
    expect(() =>
      PricePredictionSchema.parse({ ...validPayload, predictedPrice: 'expensive' }),
    ).toThrow()
  })

  it('accepts negative trendPct (falling scenario)', () => {
    const result = PricePredictionSchema.parse({ ...validPayload, trendPct: -99 })
    expect(result.trendPct).toBe(-99)
  })
})
