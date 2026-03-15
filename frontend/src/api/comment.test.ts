import { describe, it, expect } from 'vitest'
import { CommentSchema } from './comment'

describe('CommentSchema', () => {
  it('validates a correct comment', () => {
    const valid = {
      id: 'abc-123',
      missionId: 'mission-xyz',
      author: 'Alice',
      body: 'Looks great!',
      createdAt: '2024-01-01T10:00:00Z',
      updatedAt: '2024-01-01T10:00:00Z',
    }
    expect(() => CommentSchema.parse(valid)).not.toThrow()
    const result = CommentSchema.parse(valid)
    expect(result.id).toBe('abc-123')
    expect(result.author).toBe('Alice')
    expect(result.body).toBe('Looks great!')
  })

  it('validates a comment with optional optionId', () => {
    const valid = {
      id: 'abc-123',
      missionId: 'mission-xyz',
      optionId: 'option-456',
      author: 'Bob',
      body: 'This option is better.',
      createdAt: '2024-01-01T10:00:00Z',
      updatedAt: '2024-01-01T10:00:00Z',
    }
    expect(() => CommentSchema.parse(valid)).not.toThrow()
    const result = CommentSchema.parse(valid)
    expect(result.optionId).toBe('option-456')
  })

  it('rejects a comment missing body', () => {
    const invalid = {
      id: 'abc-123',
      missionId: 'mission-xyz',
      author: 'Alice',
      createdAt: '2024-01-01T10:00:00Z',
      updatedAt: '2024-01-01T10:00:00Z',
    }
    expect(() => CommentSchema.parse(invalid)).toThrow()
  })

  it('rejects a comment missing author', () => {
    const invalid = {
      id: 'abc-123',
      missionId: 'mission-xyz',
      body: 'Some text',
      createdAt: '2024-01-01T10:00:00Z',
      updatedAt: '2024-01-01T10:00:00Z',
    }
    expect(() => CommentSchema.parse(invalid)).toThrow()
  })

  it('allows optionId to be absent (no field)', () => {
    const valid = {
      id: 'c1',
      missionId: 'm1',
      author: 'You',
      body: 'Hello',
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-01T00:00:00Z',
    }
    const result = CommentSchema.parse(valid)
    expect(result.optionId).toBeUndefined()
  })
})
