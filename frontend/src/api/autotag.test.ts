import { describe, it, expect } from 'vitest'
import { CategorySuggestionSchema } from './autotag'

describe('CategorySuggestionSchema', () => {
  it('validates a correct suggestion', () => {
    const valid = {
      category: 'Informatique',
      confidence: 0.92,
      alternatives: ['Électronique', 'Audiovisuel'],
    }
    const result = CategorySuggestionSchema.safeParse(valid)
    expect(result.success).toBe(true)
    if (result.success) {
      expect(result.data.category).toBe('Informatique')
      expect(result.data.confidence).toBe(0.92)
      expect(result.data.alternatives).toHaveLength(2)
    }
  })

  it('defaults alternatives to [] when omitted', () => {
    const payload = { category: 'Voyage', confidence: 0.75 }
    const result = CategorySuggestionSchema.safeParse(payload)
    expect(result.success).toBe(true)
    if (result.success) {
      expect(result.data.alternatives).toEqual([])
    }
  })

  it('rejects missing category', () => {
    const invalid = { confidence: 0.8, alternatives: [] }
    const result = CategorySuggestionSchema.safeParse(invalid)
    expect(result.success).toBe(false)
  })

  it('rejects missing confidence', () => {
    const invalid = { category: 'Maison', alternatives: [] }
    const result = CategorySuggestionSchema.safeParse(invalid)
    expect(result.success).toBe(false)
  })

  it('rejects non-number confidence', () => {
    const invalid = { category: 'Sport', confidence: 'high', alternatives: [] }
    const result = CategorySuggestionSchema.safeParse(invalid)
    expect(result.success).toBe(false)
  })
})
