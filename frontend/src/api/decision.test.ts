import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ApiError } from './client'
import { getDecision } from './decision'

vi.mock('./client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('./client')>()
  return { ...actual, apiFetch: vi.fn() }
})

import { apiFetch } from './client'
const mockApiFetch = vi.mocked(apiFetch)

beforeEach(() => {
  mockApiFetch.mockReset()
})

describe('getDecision', () => {
  it('returns null when the API responds 404 (no decision yet)', async () => {
    mockApiFetch.mockRejectedValueOnce(new ApiError(404, 'not found'))
    const result = await getDecision('some-mission-id')
    expect(result).toBeNull()
  })

  it('throws when the API responds 500 (server error should propagate)', async () => {
    mockApiFetch.mockRejectedValueOnce(new ApiError(500, 'internal error'))
    await expect(getDecision('some-mission-id')).rejects.toBeInstanceOf(ApiError)
  })

  it('returns parsed DTO on success', async () => {
    const validDecision = {
      id: '550e8400-e29b-41d4-a716-446655440001',
      missionId: '550e8400-e29b-41d4-a716-446655440002',
      scores: [],
      summary: 'Aucune option',
      createdAt: '2026-03-16T00:00:00Z',
    }
    mockApiFetch.mockResolvedValueOnce(validDecision)
    const result = await getDecision('some-mission-id')
    expect(result).not.toBeNull()
    expect(result!.summary).toBe('Aucune option')
  })
})
