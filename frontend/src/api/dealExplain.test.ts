import { describe, it, expect } from 'vitest'
import { DealExplanationSchema } from './dealExplain'

describe('DealExplanationSchema', () => {
  it('validates correct data', () => {
    const valid = {
      itemId: 'abc-123',
      explanation: "C'est une excellente affaire car le prix est en baisse.",
      tags: ['prix bas', 'tendance baisse'],
      generatedAt: '2026-03-15T10:00:00Z',
    }
    const result = DealExplanationSchema.safeParse(valid)
    expect(result.success).toBe(true)
    if (result.success) {
      expect(result.data.itemId).toBe('abc-123')
      expect(result.data.tags).toHaveLength(2)
    }
  })

  it('rejects missing explanation field', () => {
    const invalid = {
      itemId: 'abc-123',
      tags: ['prix bas'],
      generatedAt: '2026-03-15T10:00:00Z',
    }
    const result = DealExplanationSchema.safeParse(invalid)
    expect(result.success).toBe(false)
  })

  it('rejects missing itemId', () => {
    const invalid = {
      explanation: 'Bonne affaire.',
      tags: ['prix bas'],
      generatedAt: '2026-03-15T10:00:00Z',
    }
    const result = DealExplanationSchema.safeParse(invalid)
    expect(result.success).toBe(false)
  })

  it('rejects non-array tags', () => {
    const invalid = {
      itemId: 'abc-123',
      explanation: 'Bonne affaire.',
      tags: 'prix bas',
      generatedAt: '2026-03-15T10:00:00Z',
    }
    const result = DealExplanationSchema.safeParse(invalid)
    expect(result.success).toBe(false)
  })

  it('accepts empty tags array', () => {
    const valid = {
      itemId: 'abc-123',
      explanation: 'Bonne affaire.',
      tags: [],
      generatedAt: '2026-03-15T10:00:00Z',
    }
    const result = DealExplanationSchema.safeParse(valid)
    expect(result.success).toBe(true)
  })
})
