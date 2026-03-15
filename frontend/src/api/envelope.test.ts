import { describe, it, expect } from 'vitest'
import { EnvelopeSchema, EnvelopeSummarySchema } from './envelope'

// Valid RFC 4122 v4 UUIDs for tests
const UUID_1 = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'
const UUID_2 = 'b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a22'

describe('EnvelopeSchema', () => {
  const valid = {
    id: UUID_1,
    name: 'Alimentation',
    monthlyAmount: 300,
    currency: 'EUR',
    createdAt: '2026-01-01T00:00:00Z',
  }

  it('parses a valid envelope', () => {
    const result = EnvelopeSchema.parse(valid)
    expect(result.id).toBe(valid.id)
    expect(result.name).toBe('Alimentation')
    expect(result.monthlyAmount).toBe(300)
  })

  it('rejects missing id', () => {
    const { id: _id, ...rest } = valid
    expect(() => EnvelopeSchema.parse(rest)).toThrow()
  })

  it('rejects non-uuid id', () => {
    expect(() => EnvelopeSchema.parse({ ...valid, id: 'not-a-uuid' })).toThrow()
  })

  it('rejects non-number monthlyAmount', () => {
    expect(() => EnvelopeSchema.parse({ ...valid, monthlyAmount: 'three hundred' })).toThrow()
  })
})

describe('EnvelopeSummarySchema', () => {
  const validSummary = {
    envelope: {
      id: UUID_2,
      name: 'Loisirs',
      monthlyAmount: 200,
      currency: 'EUR',
      createdAt: '2026-01-01T00:00:00Z',
    },
    missionCount: 3,
    totalBudget: 1500,
    totalSpent: 750,
  }

  it('parses a valid envelope summary', () => {
    const result = EnvelopeSummarySchema.parse(validSummary)
    expect(result.missionCount).toBe(3)
    expect(result.totalBudget).toBe(1500)
    expect(result.totalSpent).toBe(750)
    expect(result.envelope.name).toBe('Loisirs')
  })

  it('parses summary with zero values', () => {
    const result = EnvelopeSummarySchema.parse({ ...validSummary, missionCount: 0, totalBudget: 0, totalSpent: 0 })
    expect(result.missionCount).toBe(0)
    expect(result.totalSpent).toBe(0)
  })

  it('rejects missing missionCount', () => {
    const { missionCount: _mc, ...rest } = validSummary
    expect(() => EnvelopeSummarySchema.parse(rest)).toThrow()
  })

  it('rejects non-number totalSpent', () => {
    expect(() => EnvelopeSummarySchema.parse({ ...validSummary, totalSpent: 'a lot' })).toThrow()
  })

  it('rejects invalid nested envelope', () => {
    expect(() =>
      EnvelopeSummarySchema.parse({ ...validSummary, envelope: { ...validSummary.envelope, id: 'bad' } }),
    ).toThrow()
  })
})
