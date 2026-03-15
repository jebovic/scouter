import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useCoachTips, useRefreshCoach } from './useCoach'
import * as coachApi from '../api/coach'
import React, { ReactNode } from 'react'

vi.mock('../api/coach')
vi.mock('../components/scouter', () => ({
  useToast: () => ({ toast: vi.fn() }),
}))

describe('useCoach', () => {
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

  describe('useCoachTips', () => {
    it('fetches coach tips when slug is provided', async () => {
      const mockResponse = {
        missionId: 'mission-123',
        tips: [
          {
            id: 'tip1',
            category: 'research' as const,
            priority: 'high' as const,
            title: 'Expand options',
            body: 'Consider more brands',
          },
        ],
        generatedAt: new Date().toISOString(),
      }

      vi.spyOn(coachApi, 'fetchCoachTips').mockResolvedValue(mockResponse)

      const { result } = renderHook(() => useCoachTips('test-mission'), { wrapper })

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true)
      })

      expect(result.current.data).toEqual(mockResponse)
    })

    it('does not fetch when slug is undefined', () => {
      vi.spyOn(coachApi, 'fetchCoachTips')

      const { result } = renderHook(() => useCoachTips(undefined), { wrapper })

      expect(result.current.isLoading).toBe(false)
      expect(coachApi.fetchCoachTips).not.toHaveBeenCalled()
    })

    it('handles fetch errors gracefully', async () => {
      const error = new Error('Network error')
      vi.spyOn(coachApi, 'fetchCoachTips').mockRejectedValue(error)

      const { result } = renderHook(() => useCoachTips('test-mission'), { wrapper })

      await waitFor(() => {
        expect(result.current.isError).toBe(true)
      })

      expect(result.current.error).toBeDefined()
    })

    it('caches results for 15 minutes', async () => {
      const mockResponse = {
        missionId: 'mission-123',
        tips: [],
        generatedAt: new Date().toISOString(),
      }

      const fetchSpy = vi.spyOn(coachApi, 'fetchCoachTips').mockResolvedValue(mockResponse)

      const { result: result1 } = renderHook(() => useCoachTips('test-mission'), { wrapper })

      await waitFor(() => {
        expect(result1.current.isSuccess).toBe(true)
      })

      expect(fetchSpy).toHaveBeenCalledTimes(1)

      // Second call should use cache
      const { result: result2 } = renderHook(() => useCoachTips('test-mission'), { wrapper })

      expect(fetchSpy).toHaveBeenCalledTimes(1) // Still 1, not 2
    })
  })

  describe('useRefreshCoach', () => {
    it('refreshes coach tips and invalidates query', async () => {
      const mockResponse = {
        missionId: 'mission-123',
        tips: [],
        generatedAt: new Date().toISOString(),
      }

      vi.spyOn(coachApi, 'fetchCoachTips').mockResolvedValue(mockResponse)

      const { result } = renderHook(() => useRefreshCoach('test-mission'), { wrapper })

      await waitFor(() => {
        expect(result.current.isIdle).toBe(true)
      })

      result.current.mutate()

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true)
      })

      expect(coachApi.fetchCoachTips).toHaveBeenCalledWith('test-mission')
    })

    it('handles refresh errors', async () => {
      const error = new Error('Refresh failed')
      vi.spyOn(coachApi, 'fetchCoachTips').mockRejectedValue(error)

      const { result } = renderHook(() => useRefreshCoach('test-mission'), { wrapper })

      result.current.mutate()

      await waitFor(() => {
        expect(result.current.isError).toBe(true)
      })
    })

    it('does not allow refresh without slug', async () => {
      const { result } = renderHook(() => useRefreshCoach(undefined), { wrapper })

      result.current.mutate()

      await waitFor(() => {
        expect(result.current.isError).toBe(true)
      })
    })
  })
})
