import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useMissionHealth } from './useMissionHealth'
import * as healthApi from '../api/health'
import React, { ReactNode } from 'react'

vi.mock('../api/health')

const mockReport: healthApi.HealthReport = {
  missionId: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
  overallScore: 72,
  grade: 'C',
  signals: [
    {
      name: 'Research Depth',
      score: 60,
      weight: 0.25,
      description: 'Options researched',
      icon: '🔬',
    },
  ],
  summary: 'Bonne progression.',
  advice: 'Continuez.',
  generatedAt: '2026-03-15T10:00:00Z',
}

describe('useMissionHealth', () => {
  let queryClient: QueryClient

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

  it('fetches health report when slug is provided', async () => {
    vi.spyOn(healthApi, 'fetchHealthReport').mockResolvedValue(mockReport)

    const { result } = renderHook(() => useMissionHealth('mon-laptop'), { wrapper })

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })

    expect(result.current.data).toEqual(mockReport)
    expect(healthApi.fetchHealthReport).toHaveBeenCalledWith('mon-laptop')
  })

  it('does not fetch when slug is undefined', () => {
    vi.spyOn(healthApi, 'fetchHealthReport')

    const { result } = renderHook(() => useMissionHealth(undefined), { wrapper })

    expect(result.current.isLoading).toBe(false)
    expect(healthApi.fetchHealthReport).not.toHaveBeenCalled()
  })

  it('does not fetch when slug is empty string', () => {
    vi.spyOn(healthApi, 'fetchHealthReport')

    const { result } = renderHook(() => useMissionHealth(''), { wrapper })

    expect(result.current.isLoading).toBe(false)
    expect(healthApi.fetchHealthReport).not.toHaveBeenCalled()
  })

  it('exposes error state on API failure', async () => {
    vi.spyOn(healthApi, 'fetchHealthReport').mockRejectedValue(new Error('Network error'))

    const { result } = renderHook(() => useMissionHealth('mon-laptop'), { wrapper })

    await waitFor(() => {
      expect(result.current.isError).toBe(true)
    })

    expect(result.current.error).toBeDefined()
  })

  it('uses 30-minute stale time so second render does not re-fetch', async () => {
    const fetchSpy = vi
      .spyOn(healthApi, 'fetchHealthReport')
      .mockResolvedValue(mockReport)

    const { result: r1 } = renderHook(() => useMissionHealth('mon-laptop'), { wrapper })
    await waitFor(() => expect(r1.current.isSuccess).toBe(true))

    const { result: r2 } = renderHook(() => useMissionHealth('mon-laptop'), { wrapper })
    await waitFor(() => expect(r2.current.isSuccess).toBe(true))

    expect(fetchSpy).toHaveBeenCalledTimes(1)
  })
})
