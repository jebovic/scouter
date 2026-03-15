import { describe, it, expect } from 'vitest'
import { WeightSchema, OptionScoreSchema, MatrixResponseSchema } from './comparison'

describe('WeightSchema', () => {
  const validWeight = {
    id: '11111111-1111-1111-1111-111111111111',
    missionId: '22222222-2222-2222-2222-222222222222',
    criteria: 'price',
    weight: 75,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
  }

  it('validates a correct weight', () => {
    expect(() => WeightSchema.parse(validWeight)).not.toThrow()
    const result = WeightSchema.parse(validWeight)
    expect(result.criteria).toBe('price')
    expect(result.weight).toBe(75)
  })

  it('accepts weight of 0', () => {
    expect(() => WeightSchema.parse({ ...validWeight, weight: 0 })).not.toThrow()
  })

  it('accepts weight of 100', () => {
    expect(() => WeightSchema.parse({ ...validWeight, weight: 100 })).not.toThrow()
  })

  it('rejects missing id', () => {
    const { id: _id, ...noId } = validWeight
    expect(() => WeightSchema.parse(noId)).toThrow()
  })

  it('rejects missing criteria', () => {
    const { criteria: _c, ...noCriteria } = validWeight
    expect(() => WeightSchema.parse(noCriteria)).toThrow()
  })

  it('rejects non-numeric weight', () => {
    expect(() => WeightSchema.parse({ ...validWeight, weight: 'high' })).toThrow()
  })

  it('rejects missing missionId', () => {
    const { missionId: _m, ...noMission } = validWeight
    expect(() => WeightSchema.parse(noMission)).toThrow()
  })
})

describe('OptionScoreSchema', () => {
  const validScore = {
    optionId: '33333333-3333-3333-3333-333333333333',
    optionName: 'MacBook Pro 16"',
    score: 87.5,
    rank: 1,
  }

  it('validates a correct option score', () => {
    expect(() => OptionScoreSchema.parse(validScore)).not.toThrow()
    const result = OptionScoreSchema.parse(validScore)
    expect(result.optionName).toBe('MacBook Pro 16"')
    expect(result.score).toBe(87.5)
    expect(result.rank).toBe(1)
  })

  it('accepts score of 0', () => {
    expect(() => OptionScoreSchema.parse({ ...validScore, score: 0 })).not.toThrow()
  })

  it('accepts score of 100', () => {
    expect(() => OptionScoreSchema.parse({ ...validScore, score: 100 })).not.toThrow()
  })

  it('rejects missing optionId', () => {
    const { optionId: _o, ...noId } = validScore
    expect(() => OptionScoreSchema.parse(noId)).toThrow()
  })

  it('rejects non-numeric score', () => {
    expect(() => OptionScoreSchema.parse({ ...validScore, score: 'best' })).toThrow()
  })

  it('rejects non-numeric rank', () => {
    expect(() => OptionScoreSchema.parse({ ...validScore, rank: 'first' })).toThrow()
  })
})

describe('MatrixResponseSchema', () => {
  const validMatrix = {
    weights: [
      {
        id: '11111111-1111-1111-1111-111111111111',
        missionId: '22222222-2222-2222-2222-222222222222',
        criteria: 'price',
        weight: 80,
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-01T00:00:00Z',
      },
    ],
    scores: [
      {
        optionId: '33333333-3333-3333-3333-333333333333',
        optionName: 'Option A',
        score: 90,
        rank: 1,
      },
    ],
  }

  it('validates a correct matrix response', () => {
    expect(() => MatrixResponseSchema.parse(validMatrix)).not.toThrow()
    const result = MatrixResponseSchema.parse(validMatrix)
    expect(result.weights).toHaveLength(1)
    expect(result.scores).toHaveLength(1)
  })

  it('validates empty weights and scores', () => {
    const empty = { weights: [], scores: [] }
    expect(() => MatrixResponseSchema.parse(empty)).not.toThrow()
    const result = MatrixResponseSchema.parse(empty)
    expect(result.weights).toHaveLength(0)
    expect(result.scores).toHaveLength(0)
  })

  it('validates multiple weights and scores', () => {
    const multi = {
      weights: [
        { ...validMatrix.weights[0], criteria: 'price' },
        { ...validMatrix.weights[0], id: '44444444-4444-4444-4444-444444444444', criteria: 'quality' },
      ],
      scores: [
        validMatrix.scores[0],
        { optionId: '55555555-5555-5555-5555-555555555555', optionName: 'Option B', score: 60, rank: 2 },
      ],
    }
    expect(() => MatrixResponseSchema.parse(multi)).not.toThrow()
    const result = MatrixResponseSchema.parse(multi)
    expect(result.weights).toHaveLength(2)
    expect(result.scores).toHaveLength(2)
  })

  it('rejects missing weights field', () => {
    const { weights: _w, ...noWeights } = validMatrix
    expect(() => MatrixResponseSchema.parse(noWeights)).toThrow()
  })

  it('rejects missing scores field', () => {
    const { scores: _s, ...noScores } = validMatrix
    expect(() => MatrixResponseSchema.parse(noScores)).toThrow()
  })

  it('rejects invalid weight inside weights array', () => {
    const invalid = {
      weights: [{ criteria: 'price' }], // missing required fields
      scores: [],
    }
    expect(() => MatrixResponseSchema.parse(invalid)).toThrow()
  })
})
