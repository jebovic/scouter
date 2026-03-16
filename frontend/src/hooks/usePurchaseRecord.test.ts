import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import React from 'react'
import { usePurchaseRecord } from './usePurchase'

vi.mock('../api/purchase', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../api/purchase')>()
  return { ...actual, getPurchaseRecord: vi.fn() }
})

import { getPurchaseRecord } from '../api/purchase'
const mockGet = vi.mocked(getPurchaseRecord)

// Use retry: 2 as QueryClient default to prove the hook's own retry: false overrides it
function wrapper({ children }: { children: React.ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: 2 } } })
  return React.createElement(QueryClientProvider, { client: qc }, children)
}

beforeEach(() => {
  mockGet.mockReset()
})

describe('usePurchaseRecord', () => {
  it('returns null data when getPurchaseRecord resolves null (no record yet)', async () => {
    mockGet.mockResolvedValue(null)
    const { result } = renderHook(() => usePurchaseRecord('mission-123'), { wrapper })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toBeNull()
    expect(mockGet).toHaveBeenCalledTimes(1)
  })

  it('calls getPurchaseRecord exactly once on failure (retry: false)', async () => {
    const error = new Error('server error')
    mockGet.mockRejectedValue(error)
    const { result } = renderHook(() => usePurchaseRecord('mission-456'), { wrapper })
    await waitFor(() => expect(result.current.isError).toBe(true))
    // With retry: false on the hook, mock must be called exactly once (not retried 2 more times from QueryClient)
    expect(mockGet).toHaveBeenCalledTimes(1)
  })
})
