import { describe, it, expect } from 'vitest'
import { PriceDropItemSchema, WeeklyDigestSchema } from './digest'

const validItem = {
  missionId: 'm1',
  missionName: 'Laptop Mission',
  missionSlug: 'laptop-mission',
  itemId: 'i1',
  itemName: 'Dell XPS 15',
  merchant: 'Amazon',
  priceBefore: 1200,
  priceNow: 960,
  dropPct: 20,
  currency: 'EUR',
}

const validDigest = {
  generatedAt: '2026-03-15T10:00:00Z',
  periodDays: 7,
  totalDrops: 1,
  totalSavings: 240,
  currency: 'EUR',
  items: [validItem],
}

describe('PriceDropItemSchema', () => {
  it('accepts a valid price drop item', () => {
    expect(() => PriceDropItemSchema.parse(validItem)).not.toThrow()
    const result = PriceDropItemSchema.parse(validItem)
    expect(result.missionId).toBe('m1')
    expect(result.dropPct).toBe(20)
  })

  it('rejects missing missionId', () => {
    const { missionId: _, ...rest } = validItem
    expect(() => PriceDropItemSchema.parse(rest)).toThrow()
  })

  it('rejects missing itemName', () => {
    const { itemName: _, ...rest } = validItem
    expect(() => PriceDropItemSchema.parse(rest)).toThrow()
  })

  it('rejects non-numeric priceBefore', () => {
    expect(() => PriceDropItemSchema.parse({ ...validItem, priceBefore: 'expensive' })).toThrow()
  })

  it('rejects non-numeric priceNow', () => {
    expect(() => PriceDropItemSchema.parse({ ...validItem, priceNow: 'cheap' })).toThrow()
  })

  it('rejects non-numeric dropPct', () => {
    expect(() => PriceDropItemSchema.parse({ ...validItem, dropPct: '20%' })).toThrow()
  })

  it('rejects missing merchant', () => {
    const { merchant: _, ...rest } = validItem
    expect(() => PriceDropItemSchema.parse(rest)).toThrow()
  })

  it('rejects missing currency', () => {
    const { currency: _, ...rest } = validItem
    expect(() => PriceDropItemSchema.parse(rest)).toThrow()
  })
})

describe('WeeklyDigestSchema', () => {
  it('accepts a valid weekly digest', () => {
    expect(() => WeeklyDigestSchema.parse(validDigest)).not.toThrow()
    const result = WeeklyDigestSchema.parse(validDigest)
    expect(result.totalDrops).toBe(1)
    expect(result.periodDays).toBe(7)
    expect(result.currency).toBe('EUR')
    expect(result.items).toHaveLength(1)
  })

  it('accepts empty items array', () => {
    const noDrops = { ...validDigest, items: [], totalDrops: 0, totalSavings: 0 }
    expect(() => WeeklyDigestSchema.parse(noDrops)).not.toThrow()
    expect(WeeklyDigestSchema.parse(noDrops).items).toHaveLength(0)
  })

  it('rejects missing generatedAt', () => {
    const { generatedAt: _, ...rest } = validDigest
    expect(() => WeeklyDigestSchema.parse(rest)).toThrow()
  })

  it('rejects missing periodDays', () => {
    const { periodDays: _, ...rest } = validDigest
    expect(() => WeeklyDigestSchema.parse(rest)).toThrow()
  })

  it('rejects missing totalDrops', () => {
    const { totalDrops: _, ...rest } = validDigest
    expect(() => WeeklyDigestSchema.parse(rest)).toThrow()
  })

  it('rejects missing totalSavings', () => {
    const { totalSavings: _, ...rest } = validDigest
    expect(() => WeeklyDigestSchema.parse(rest)).toThrow()
  })

  it('rejects missing currency', () => {
    const { currency: _, ...rest } = validDigest
    expect(() => WeeklyDigestSchema.parse(rest)).toThrow()
  })

  it('rejects missing items', () => {
    const { items: _, ...rest } = validDigest
    expect(() => WeeklyDigestSchema.parse(rest)).toThrow()
  })

  it('rejects non-numeric totalSavings', () => {
    expect(() => WeeklyDigestSchema.parse({ ...validDigest, totalSavings: 'a lot' })).toThrow()
  })

  it('rejects items containing an invalid entry', () => {
    const badItem = { ...validItem, dropPct: 'much' }
    expect(() => WeeklyDigestSchema.parse({ ...validDigest, items: [badItem] })).toThrow()
  })
})
