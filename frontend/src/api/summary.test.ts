import { describe, it, expect } from 'vitest'
import { MissionSummarySchema, VerdictSchema } from './summary'

describe('MissionSummarySchema', () => {
  const validPayload = {
    missionSlug: 'casque-audio',
    bullets: [
      '✅ Meilleure option: Sony WH-1000XM5 à 279€ (score 94/100)',
      '⚠️ Budget: 85% utilisé (255€/300€)',
      '🎯 Prochaine action: comparer avec Bose QC45 avant d\'acheter',
    ],
    verdict: 'go' as const,
    cachedAt: 1710000000,
  }

  it('parses a valid go verdict payload', () => {
    const result = MissionSummarySchema.parse(validPayload)
    expect(result.missionSlug).toBe('casque-audio')
    expect(result.verdict).toBe('go')
    expect(result.bullets).toHaveLength(3)
    expect(result.cachedAt).toBe(1710000000)
  })

  it('parses a wait verdict payload', () => {
    const result = MissionSummarySchema.parse({ ...validPayload, verdict: 'wait' })
    expect(result.verdict).toBe('wait')
  })

  it('parses a research_more verdict payload', () => {
    const result = MissionSummarySchema.parse({ ...validPayload, verdict: 'research_more' })
    expect(result.verdict).toBe('research_more')
  })

  it('rejects an invalid verdict', () => {
    expect(() =>
      MissionSummarySchema.parse({ ...validPayload, verdict: 'buy_now' })
    ).toThrow()
  })

  it('rejects missing bullets array', () => {
    const { bullets: _bullets, ...withoutBullets } = validPayload
    expect(() => MissionSummarySchema.parse(withoutBullets)).toThrow()
  })

  it('rejects empty bullets array', () => {
    expect(() =>
      MissionSummarySchema.parse({ ...validPayload, bullets: [] })
    ).toThrow()
  })

  it('rejects missing missionSlug', () => {
    const { missionSlug: _slug, ...withoutSlug } = validPayload
    expect(() => MissionSummarySchema.parse(withoutSlug)).toThrow()
  })

  it('rejects non-integer cachedAt', () => {
    expect(() =>
      MissionSummarySchema.parse({ ...validPayload, cachedAt: 1.5 })
    ).toThrow()
  })

  it('accepts multiple bullets (5 points)', () => {
    const fiveBullets = {
      ...validPayload,
      bullets: [
        '✅ Point 1',
        '⚠️ Point 2',
        '🎯 Point 3',
        '📉 Point 4',
        '🔍 Point 5',
      ],
    }
    const result = MissionSummarySchema.parse(fiveBullets)
    expect(result.bullets).toHaveLength(5)
  })
})

describe('VerdictSchema', () => {
  it('accepts go', () => expect(VerdictSchema.parse('go')).toBe('go'))
  it('accepts wait', () => expect(VerdictSchema.parse('wait')).toBe('wait'))
  it('accepts research_more', () => expect(VerdictSchema.parse('research_more')).toBe('research_more'))
  it('rejects unknown values', () => expect(() => VerdictSchema.parse('unknown')).toThrow())
})
