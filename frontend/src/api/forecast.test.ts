import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ApiError } from './client'
import { fetchForecast } from './forecast'

vi.mock('./client', async (importOriginal) => {
  const actual = await importOriginal<typeof import('./client')>()
  return { ...actual, apiFetch: vi.fn() }
})

import { apiFetch } from './client'
const mockApiFetch = vi.mocked(apiFetch)

beforeEach(() => {
  mockApiFetch.mockReset()
})

describe('fetchForecast', () => {
  it('returns null when the API responds 404 (no forecast yet)', async () => {
    mockApiFetch.mockRejectedValueOnce(new ApiError(404, 'not found'))
    const result = await fetchForecast('some-mission-id')
    expect(result).toBeNull()
  })

  it('throws when the API responds 500 (server error should propagate)', async () => {
    mockApiFetch.mockRejectedValueOnce(new ApiError(500, 'internal error'))
    await expect(fetchForecast('some-mission-id')).rejects.toBeInstanceOf(ApiError)
  })

  it('returns parsed DTO on success', async () => {
    const validForecast = {
      id: 'f1',
      missionId: 'm1',
      estimatedTotal: 299.99,
      confidence: 'high',
      recommendations: ['Buy now'],
      riskFactors: [],
      monthsToSave: null,
      optimalBuyTime: null,
      createdAt: '2026-03-16T00:00:00Z',
    }
    mockApiFetch.mockResolvedValueOnce(validForecast)
    const result = await fetchForecast('m1')
    expect(result).not.toBeNull()
    expect(result!.confidence).toBe('high')
  })
})
