import { describe, it, expect } from 'vitest'
import { ReceiptItemSchema, ReceiptAnalysisSchema } from './receipt'

describe('ReceiptItemSchema', () => {
  it('parses a valid receipt item', () => {
    const item = { name: 'Pain', quantity: 2, unitPrice: 1.5, total: 3.0 }
    expect(ReceiptItemSchema.safeParse(item).success).toBe(true)
  })

  it('rejects missing name', () => {
    const item = { quantity: 1, unitPrice: 1.0, total: 1.0 }
    expect(ReceiptItemSchema.safeParse(item).success).toBe(false)
  })

  it('rejects non-numeric quantity', () => {
    const item = { name: 'Lait', quantity: 'deux', unitPrice: 1.2, total: 2.4 }
    expect(ReceiptItemSchema.safeParse(item).success).toBe(false)
  })
})

describe('ReceiptAnalysisSchema', () => {
  it('parses a full valid analysis', () => {
    const analysis = {
      id: 'abc-123',
      missionSlug: 'courses-semaine',
      rawText: 'Pain 2x1.50',
      items: [{ name: 'Pain', quantity: 2, unitPrice: 1.5, total: 3.0 }],
      total: 3.0,
      currency: 'EUR',
      analyzedAt: new Date().toISOString(),
    }
    expect(ReceiptAnalysisSchema.safeParse(analysis).success).toBe(true)
  })

  it('accepts empty items array', () => {
    const analysis = {
      id: 'abc-123',
      missionSlug: 'courses-semaine',
      rawText: 'texte non reconnu',
      items: [],
      total: 0,
      currency: 'EUR',
      analyzedAt: new Date().toISOString(),
    }
    expect(ReceiptAnalysisSchema.safeParse(analysis).success).toBe(true)
  })

  it('rejects missing missionSlug', () => {
    const analysis = {
      id: 'abc-123',
      rawText: 'Pain 2x1.50',
      items: [],
      total: 0,
      currency: 'EUR',
      analyzedAt: new Date().toISOString(),
    }
    expect(ReceiptAnalysisSchema.safeParse(analysis).success).toBe(false)
  })
})
