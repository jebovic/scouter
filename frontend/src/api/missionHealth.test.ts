import { describe, it, expect } from 'vitest'
import { SignalScoreSchema, HealthReportSchema } from './health'

const validSignal = {
  name: 'Research Depth',
  score: 60,
  weight: 0.25,
  description: 'Nombre d\'options recherchées',
  icon: '🔬',
}

const validReport = {
  missionId: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  overallScore: 62,
  grade: 'B' as const,
  signals: [validSignal],
  summary: 'Mission en bonne progression.',
  advice: 'Continuez la recherche.',
  generatedAt: '2026-03-15T10:00:00Z',
}

describe('SignalScoreSchema', () => {
  it('validates a correct signal', () => {
    expect(() => SignalScoreSchema.parse(validSignal)).not.toThrow()
    const result = SignalScoreSchema.parse(validSignal)
    expect(result.name).toBe('Research Depth')
    expect(result.score).toBe(60)
    expect(result.weight).toBe(0.25)
  })

  it('rejects missing name', () => {
    const invalid = { ...validSignal, name: undefined }
    expect(() => SignalScoreSchema.parse(invalid)).toThrow()
  })

  it('rejects non-numeric score', () => {
    const invalid = { ...validSignal, score: 'high' }
    expect(() => SignalScoreSchema.parse(invalid)).toThrow()
  })

  it('rejects non-numeric weight', () => {
    const invalid = { ...validSignal, weight: 'heavy' }
    expect(() => SignalScoreSchema.parse(invalid)).toThrow()
  })

  it('accepts score of 0 and weight of 0', () => {
    const edge = { ...validSignal, score: 0, weight: 0 }
    expect(() => SignalScoreSchema.parse(edge)).not.toThrow()
  })
})

describe('HealthReportSchema', () => {
  it('validates a correct health report', () => {
    expect(() => HealthReportSchema.parse(validReport)).not.toThrow()
    const result = HealthReportSchema.parse(validReport)
    expect(result.missionId).toBe('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa')
    expect(result.overallScore).toBe(62)
    expect(result.grade).toBe('B')
    expect(result.signals).toHaveLength(1)
    expect(result.summary).toBe('Mission en bonne progression.')
    expect(result.advice).toBe('Continuez la recherche.')
  })

  it('validates all grade values', () => {
    const grades = ['A', 'B', 'C', 'D', 'F'] as const
    for (const grade of grades) {
      expect(() => HealthReportSchema.parse({ ...validReport, grade })).not.toThrow()
    }
  })

  it('rejects invalid grade', () => {
    const invalid = { ...validReport, grade: 'Z' }
    expect(() => HealthReportSchema.parse(invalid)).toThrow()
  })

  it('rejects missing missionId', () => {
    const { missionId: _, ...invalid } = validReport
    expect(() => HealthReportSchema.parse(invalid)).toThrow()
  })

  it('validates empty signals array', () => {
    const withEmpty = { ...validReport, signals: [] }
    expect(() => HealthReportSchema.parse(withEmpty)).not.toThrow()
    expect(HealthReportSchema.parse(withEmpty).signals).toHaveLength(0)
  })
})
