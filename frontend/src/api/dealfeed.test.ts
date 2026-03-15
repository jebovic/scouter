import { describe, it, expect } from 'vitest'
import { PromoOfferSchema, PromoFeedResultSchema } from './dealfeed'

const validOffer = {
  title: 'MacBook Air M3 — Offre spéciale',
  retailer: 'Fnac',
  originalPrice: 1299.0,
  promoPrice: 999.0,
  discount: 23,
  url: 'fnac.com/macbook-air-m3',
  category: 'informatique',
  isFlash: false,
  source: 'ai_generated',
}

const validResult = {
  offers: [validOffer],
  query: 'laptop',
  category: 'informatique',
  fetchedAt: '2026-03-15T10:00:00Z',
}

describe('PromoOfferSchema', () => {
  it('accepts a valid promo offer', () => {
    expect(() => PromoOfferSchema.parse(validOffer)).not.toThrow()
    const result = PromoOfferSchema.parse(validOffer)
    expect(result.title).toBe('MacBook Air M3 — Offre spéciale')
    expect(result.retailer).toBe('Fnac')
    expect(result.discount).toBe(23)
  })

  it('accepts a flash offer with expiresAt', () => {
    const flashOffer = { ...validOffer, isFlash: true, expiresAt: '2026-03-20T23:59:00Z' }
    expect(() => PromoOfferSchema.parse(flashOffer)).not.toThrow()
    const result = PromoOfferSchema.parse(flashOffer)
    expect(result.isFlash).toBe(true)
    expect(result.expiresAt).toBe('2026-03-20T23:59:00Z')
  })

  it('accepts null expiresAt (no expiry)', () => {
    const noExpiry = { ...validOffer, expiresAt: null }
    expect(() => PromoOfferSchema.parse(noExpiry)).not.toThrow()
    const result = PromoOfferSchema.parse(noExpiry)
    expect(result.expiresAt).toBeNull()
  })

  it('accepts omitted expiresAt (nullish)', () => {
    const { expiresAt: _omit, ...withoutExpiry } = { ...validOffer, expiresAt: undefined }
    expect(() => PromoOfferSchema.parse(withoutExpiry)).not.toThrow()
  })

  it('rejects missing required retailer field', () => {
    const { retailer: _omit, ...without } = validOffer
    expect(() => PromoOfferSchema.parse(without)).toThrow()
  })

  it('rejects non-numeric originalPrice', () => {
    expect(() => PromoOfferSchema.parse({ ...validOffer, originalPrice: 'not-a-number' })).toThrow()
  })

  it('rejects non-boolean isFlash', () => {
    expect(() => PromoOfferSchema.parse({ ...validOffer, isFlash: 'yes' })).toThrow()
  })
})

describe('PromoFeedResultSchema', () => {
  it('accepts a valid promo feed result', () => {
    expect(() => PromoFeedResultSchema.parse(validResult)).not.toThrow()
    const result = PromoFeedResultSchema.parse(validResult)
    expect(result.query).toBe('laptop')
    expect(result.category).toBe('informatique')
    expect(result.offers).toHaveLength(1)
  })

  it('accepts empty offers array', () => {
    const empty = { ...validResult, offers: [] }
    expect(() => PromoFeedResultSchema.parse(empty)).not.toThrow()
    const result = PromoFeedResultSchema.parse(empty)
    expect(result.offers).toHaveLength(0)
  })

  it('accepts multiple offers', () => {
    const multi = {
      ...validResult,
      offers: [
        validOffer,
        { ...validOffer, title: 'Dell XPS 13', retailer: 'Darty', isFlash: true },
      ],
    }
    expect(() => PromoFeedResultSchema.parse(multi)).not.toThrow()
    const result = PromoFeedResultSchema.parse(multi)
    expect(result.offers).toHaveLength(2)
  })

  it('rejects missing query field', () => {
    const { query: _omit, ...without } = validResult
    expect(() => PromoFeedResultSchema.parse(without)).toThrow()
  })

  it('rejects missing fetchedAt field', () => {
    const { fetchedAt: _omit, ...without } = validResult
    expect(() => PromoFeedResultSchema.parse(without)).toThrow()
  })

  it('rejects invalid offer inside offers array', () => {
    const bad = { ...validResult, offers: [{ ...validOffer, discount: 'big' }] }
    expect(() => PromoFeedResultSchema.parse(bad)).toThrow()
  })
})
