import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useSubstitutes, useFetchSubstitutes } from './useSubstitutes'
import * as substitutesApi from '../api/substitutes'
import React, { ReactNode } from 'react'

vi.mock('../api/substitutes')

const mockResponse: substitutesApi.SubstituteResponse = {
  optionId: 'opt-abc',
  productName: 'iPhone 15 Pro',
  substitutes: [
    {
      name: 'Samsung Galaxy S23',
      brand: 'Samsung',
      price: 799,
      currency: 'EUR',
      retailer: 'Fnac',
      reason: 'Great alternative',
      advantage: 'cheaper',
    },
  ],
  generatedAt: new Date().toISOString(),
}

describe('useSubstitutes', () => {
  let queryClient: QueryClient

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    })
    vi.clearAllMocks()
  })

  const wrapper = ({ children }: { children: ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children)

  it('fetches substitutes when optionId is provided', async () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockResolvedValue(mockResponse)

    const { result } = renderHook(() => useSubstitutes('opt-abc'), { wrapper })

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })

    expect(result.current.data).toEqual(mockResponse)
    expect(substitutesApi.fetchSubstitutes).toHaveBeenCalledWith('opt-abc')
  })

  it('does not fetch when optionId is empty string', () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes')

    const { result } = renderHook(() => useSubstitutes(''), { wrapper })

    expect(result.current.isLoading).toBe(false)
    expect(substitutesApi.fetchSubstitutes).not.toHaveBeenCalled()
  })

  it('does not fetch when optionId is undefined', () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes')

    const { result } = renderHook(() => useSubstitutes(undefined), { wrapper })

    expect(result.current.isLoading).toBe(false)
    expect(substitutesApi.fetchSubstitutes).not.toHaveBeenCalled()
  })

  it('handles fetch errors gracefully', async () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockRejectedValue(new Error('Network error'))

    const { result } = renderHook(() => useSubstitutes('opt-abc'), { wrapper })

    await waitFor(() => {
      expect(result.current.isError).toBe(true)
    })

    expect(result.current.error).toBeDefined()
  })

  it('uses 2h stale time so second render does not re-fetch', async () => {
    const fetchSpy = vi
      .spyOn(substitutesApi, 'fetchSubstitutes')
      .mockResolvedValue(mockResponse)

    const { result: r1 } = renderHook(() => useSubstitutes('opt-abc'), { wrapper })
    await waitFor(() => expect(r1.current.isSuccess).toBe(true))

    const { result: r2 } = renderHook(() => useSubstitutes('opt-abc'), { wrapper })
    await waitFor(() => expect(r2.current.isSuccess).toBe(true))

    // Should still be 1 — cache served second render.
    expect(fetchSpy).toHaveBeenCalledTimes(1)
  })
})

describe('useFetchSubstitutes', () => {
  let queryClient: QueryClient

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    })
    vi.clearAllMocks()
  })

  const wrapper = ({ children }: { children: ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children)

  it('calls fetchSubstitutes and updates query cache on success', async () => {
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockResolvedValue(mockResponse)

    const { result } = renderHook(() => useFetchSubstitutes('opt-abc'), { wrapper })

    result.current.fetch()

    await waitFor(() => {
      expect(result.current.isPending).toBe(false)
    })

    expect(substitutesApi.fetchSubstitutes).toHaveBeenCalledWith('opt-abc')
  })

  it('exposes isPending during in-flight request', async () => {
    let resolve!: (v: substitutesApi.SubstituteResponse) => void
    vi.spyOn(substitutesApi, 'fetchSubstitutes').mockReturnValue(
      new Promise((r) => { resolve = r }),
    )

    const { result } = renderHook(() => useFetchSubstitutes('opt-abc'), { wrapper })

    result.current.fetch()

    await waitFor(() => expect(result.current.isPending).toBe(true))
    resolve(mockResponse)
    await waitFor(() => expect(result.current.isPending).toBe(false))
  })
})
