import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { usePricePrediction } from './usePricePrediction'
import * as predApi from '../api/prediction'
import React, { ReactNode } from 'react'

vi.mock('../api/prediction')

describe('usePricePrediction', () => {
  let queryClient: QueryClient

  const mockPrediction: predApi.PricePrediction = {
    itemId: 'item-abc',
    currentPrice: 100,
    predictedPrice: 120,
    confidenceLevel: 'high',
    trend: 'rising',
    trendPct: 20,
    dataPoints: 8,
    forecastDays: 30,
  }

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    })
    vi.clearAllMocks()
  })

  const wrapper = ({ children }: { children: ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children)

  it('fetches prediction when itemId is provided', async () => {
    vi.spyOn(predApi, 'fetchPricePrediction').mockResolvedValue(mockPrediction)

    const { result } = renderHook(() => usePricePrediction('item-abc'), { wrapper })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data).toEqual(mockPrediction)
    expect(predApi.fetchPricePrediction).toHaveBeenCalledWith('item-abc')
  })

  it('does not fetch when itemId is empty string', () => {
    vi.spyOn(predApi, 'fetchPricePrediction')

    const { result } = renderHook(() => usePricePrediction(''), { wrapper })

    expect(result.current.isLoading).toBe(false)
    expect(predApi.fetchPricePrediction).not.toHaveBeenCalled()
  })

  it('does not fetch when itemId is undefined', () => {
    vi.spyOn(predApi, 'fetchPricePrediction')

    const { result } = renderHook(() => usePricePrediction(undefined), { wrapper })

    expect(result.current.isLoading).toBe(false)
    expect(predApi.fetchPricePrediction).not.toHaveBeenCalled()
  })

  it('handles fetch errors gracefully', async () => {
    vi.spyOn(predApi, 'fetchPricePrediction').mockRejectedValue(new Error('Network error'))

    const { result } = renderHook(() => usePricePrediction('item-abc'), { wrapper })

    await waitFor(() => expect(result.current.isError).toBe(true))

    expect(result.current.error).toBeDefined()
  })

  it('returns undefined data before fetch completes', () => {
    vi.spyOn(predApi, 'fetchPricePrediction').mockImplementation(
      () => new Promise(() => {}), // never resolves
    )

    const { result } = renderHook(() => usePricePrediction('item-abc'), { wrapper })

    expect(result.current.data).toBeUndefined()
  })

  it('uses staleTime of 1 hour — does not refetch for same itemId within stale window', async () => {
    const fetchSpy = vi
      .spyOn(predApi, 'fetchPricePrediction')
      .mockResolvedValue(mockPrediction)

    // First hook render
    const { result: r1 } = renderHook(() => usePricePrediction('item-abc'), { wrapper })
    await waitFor(() => expect(r1.current.isSuccess).toBe(true))

    // Second hook render with same key — should use cache
    const { result: r2 } = renderHook(() => usePricePrediction('item-abc'), { wrapper })
    await waitFor(() => expect(r2.current.isSuccess).toBe(true))

    expect(fetchSpy).toHaveBeenCalledTimes(1)
  })
})
