import { describe, it, expect } from 'vitest'
import { NegotiationScriptSchema } from './negotiation'

const validScript = {
  openingOffer: 850,
  walkAwayPrice: 950,
  suggestedDiscount: 10,
  talkingPoints: ['Competitor X offers the same at $900', 'Bundle deal possible'],
  counterOfferScript: ['Start with $850', 'Accept up to $920', 'Walk away above $950'],
  confidenceScore: 72,
}

describe('NegotiationScriptSchema', () => {
  it('validates a complete valid negotiation script', () => {
    expect(() => NegotiationScriptSchema.parse(validScript)).not.toThrow()
    const result = NegotiationScriptSchema.parse(validScript)
    expect(result.openingOffer).toBe(850)
    expect(result.walkAwayPrice).toBe(950)
    expect(result.suggestedDiscount).toBe(10)
    expect(result.talkingPoints).toHaveLength(2)
    expect(result.counterOfferScript).toHaveLength(3)
    expect(result.confidenceScore).toBe(72)
  })

  it('validates with empty arrays for talkingPoints and counterOfferScript', () => {
    const minimal = { ...validScript, talkingPoints: [], counterOfferScript: [] }
    expect(() => NegotiationScriptSchema.parse(minimal)).not.toThrow()
    const result = NegotiationScriptSchema.parse(minimal)
    expect(result.talkingPoints).toHaveLength(0)
    expect(result.counterOfferScript).toHaveLength(0)
  })

  it('validates with zero values for numeric fields', () => {
    const zero = {
      ...validScript,
      openingOffer: 0,
      walkAwayPrice: 0,
      suggestedDiscount: 0,
      confidenceScore: 0,
    }
    expect(() => NegotiationScriptSchema.parse(zero)).not.toThrow()
  })

  it('validates with decimal numeric values', () => {
    const decimals = {
      ...validScript,
      openingOffer: 849.99,
      suggestedDiscount: 12.5,
      confidenceScore: 85.5,
    }
    expect(() => NegotiationScriptSchema.parse(decimals)).not.toThrow()
    const result = NegotiationScriptSchema.parse(decimals)
    expect(result.openingOffer).toBe(849.99)
    expect(result.suggestedDiscount).toBe(12.5)
  })

  it('rejects missing openingOffer', () => {
    const { openingOffer: _, ...rest } = validScript
    expect(() => NegotiationScriptSchema.parse(rest)).toThrow()
  })

  it('rejects missing walkAwayPrice', () => {
    const { walkAwayPrice: _, ...rest } = validScript
    expect(() => NegotiationScriptSchema.parse(rest)).toThrow()
  })

  it('rejects missing suggestedDiscount', () => {
    const { suggestedDiscount: _, ...rest } = validScript
    expect(() => NegotiationScriptSchema.parse(rest)).toThrow()
  })

  it('rejects missing talkingPoints', () => {
    const { talkingPoints: _, ...rest } = validScript
    expect(() => NegotiationScriptSchema.parse(rest)).toThrow()
  })

  it('rejects missing counterOfferScript', () => {
    const { counterOfferScript: _, ...rest } = validScript
    expect(() => NegotiationScriptSchema.parse(rest)).toThrow()
  })

  it('rejects missing confidenceScore', () => {
    const { confidenceScore: _, ...rest } = validScript
    expect(() => NegotiationScriptSchema.parse(rest)).toThrow()
  })

  it('rejects non-numeric openingOffer', () => {
    expect(() => NegotiationScriptSchema.parse({ ...validScript, openingOffer: 'eight-fifty' })).toThrow()
  })

  it('rejects non-numeric confidenceScore', () => {
    expect(() => NegotiationScriptSchema.parse({ ...validScript, confidenceScore: 'high' })).toThrow()
  })

  it('rejects non-array talkingPoints', () => {
    expect(() => NegotiationScriptSchema.parse({ ...validScript, talkingPoints: 'many points' })).toThrow()
  })

  it('rejects non-array counterOfferScript', () => {
    expect(() => NegotiationScriptSchema.parse({ ...validScript, counterOfferScript: 'script' })).toThrow()
  })

  it('rejects talkingPoints array with non-string entries', () => {
    expect(() =>
      NegotiationScriptSchema.parse({ ...validScript, talkingPoints: [1, 2, 3] }),
    ).toThrow()
  })

  it('rejects completely empty payload', () => {
    expect(() => NegotiationScriptSchema.parse({})).toThrow()
  })

  it('rejects null payload', () => {
    expect(() => NegotiationScriptSchema.parse(null)).toThrow()
  })
})
