import { describe, it, expect } from 'vitest'
import { ShoppingPersonaSchema } from './persona'

const validPersona = {
  archetype: 'Le Chasseur de Deals',
  icon: '🎯',
  description: 'Vous êtes obsédé par les meilleures offres.',
  strengths: ['Économise beaucoup', 'Très patient', 'Analyse les prix'],
  blindspots: ['Peut manquer de qualité', 'Perd du temps'],
  tips: ['Utilisez des alertes de prix', 'Comparez toujours', 'Attendez les soldes'],
  scoreProfile: { frugalité: 90, impulsivité: 15, recherche: 85, budget: 80, qualité: 60 },
  cachedAt: 1700000000,
}

describe('ShoppingPersonaSchema', () => {
  it('parses a valid shopping persona', () => {
    const result = ShoppingPersonaSchema.parse(validPersona)
    expect(result.archetype).toBe('Le Chasseur de Deals')
    expect(result.icon).toBe('🎯')
  })

  it('parses scoreProfile as record of numbers', () => {
    const result = ShoppingPersonaSchema.parse(validPersona)
    expect(result.scoreProfile['frugalité']).toBe(90)
    expect(result.scoreProfile['impulsivité']).toBe(15)
  })

  it('parses strengths, blindspots, and tips as string arrays', () => {
    const result = ShoppingPersonaSchema.parse(validPersona)
    expect(result.strengths).toHaveLength(3)
    expect(result.blindspots).toHaveLength(2)
    expect(result.tips).toHaveLength(3)
  })

  it('rejects missing required archetype field', () => {
    const { archetype: _a, ...without } = validPersona
    expect(() => ShoppingPersonaSchema.parse(without)).toThrow()
  })

  it('rejects non-number values in scoreProfile', () => {
    const invalid = {
      ...validPersona,
      scoreProfile: { frugalité: 'high', impulsivité: 15 },
    }
    expect(() => ShoppingPersonaSchema.parse(invalid)).toThrow()
  })
})
