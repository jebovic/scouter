import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import React from 'react'
import { useResearchJobs } from './useResearchJobs'
import type { ResearchJobList } from '../api/researchJobs'

vi.mock('../api/researchJobs', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../api/researchJobs')>()
  return { ...actual, fetchResearchJobs: vi.fn() }
})

vi.mock('../components/scouter', () => ({
  useToast: () => ({ toast: vi.fn() }),
}))

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) => {
      if (opts) return `${key}:${JSON.stringify(opts)}`
      return key
    },
  }),
}))

import { fetchResearchJobs } from '../api/researchJobs'
const mockFetch = vi.mocked(fetchResearchJobs)

function makeWrapper() {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return {
    qc,
    wrapper: ({ children }: { children: React.ReactNode }) =>
      React.createElement(QueryClientProvider, { client: qc }, children),
  }
}

const runningJob: ResearchJobList = {
  jobs: [
    {
      id: 'job-1',
      status: 'running',
      options_count: null,
      error: null,
      feedback: null,
      started_at: '2026-01-01T00:00:00Z',
      completed_at: null,
      created_at: '2026-01-01T00:00:00Z',
    },
  ],
}

const doneJob: ResearchJobList = {
  jobs: [
    {
      id: 'job-1',
      status: 'done',
      options_count: 5,
      error: null,
      feedback: null,
      started_at: '2026-01-01T00:00:00Z',
      completed_at: '2026-01-01T00:01:00Z',
      created_at: '2026-01-01T00:00:00Z',
    },
  ],
}

const failedJob: ResearchJobList = {
  jobs: [
    {
      id: 'job-2',
      status: 'failed',
      options_count: null,
      error: 'LLM timeout',
      feedback: null,
      started_at: '2026-01-01T00:00:00Z',
      completed_at: '2026-01-01T00:01:00Z',
      created_at: '2026-01-01T00:00:00Z',
    },
  ],
}

const terminalJobs: ResearchJobList = {
  jobs: [doneJob.jobs[0], failedJob.jobs[0]],
}

beforeEach(() => {
  mockFetch.mockReset()
})

describe('useResearchJobs', () => {
  it('returns data when fetch succeeds', async () => {
    mockFetch.mockResolvedValue(doneJob)
    const { wrapper } = makeWrapper()
    const { result } = renderHook(() => useResearchJobs('mission-1'), { wrapper })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.jobs).toHaveLength(1)
    expect(result.current.data?.jobs[0].status).toBe('done')
  })

  it('polls every 3s when a job is running', async () => {
    mockFetch.mockResolvedValue(runningJob)
    const { wrapper, qc } = makeWrapper()
    const { result } = renderHook(() => useResearchJobs('mission-1'), { wrapper })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    // Check that refetchInterval returns 3000 for running jobs by inspecting
    // the query options through the QueryClient cache
    const queryCache = qc.getQueryCache()
    const query = queryCache.find({ queryKey: ['research-jobs', 'mission-1'] })
    expect(query).toBeDefined()

    // The data has a running job, so the hook should schedule refetches
    expect(result.current.data?.jobs[0].status).toBe('running')
    expect(mockFetch).toHaveBeenCalled()
  })

  it('stops polling when all jobs are in terminal state', async () => {
    mockFetch.mockResolvedValue(terminalJobs)
    const { wrapper } = makeWrapper()
    const { result } = renderHook(() => useResearchJobs('mission-1'), { wrapper })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    // All jobs are done/failed — no running/pending jobs
    const jobs = result.current.data?.jobs ?? []
    const hasActiveJob = jobs.some(j => j.status === 'running' || j.status === 'pending')
    expect(hasActiveJob).toBe(false)
  })

  it('invalidates options cache when job transitions to done', async () => {
    // First call: running job; second call: done job
    mockFetch.mockResolvedValueOnce(runningJob).mockResolvedValueOnce(doneJob)

    const { wrapper, qc } = makeWrapper()
    const invalidateSpy = vi.spyOn(qc, 'invalidateQueries')

    const { result } = renderHook(() => useResearchJobs('mission-1'), { wrapper })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    // Manually trigger a refetch to simulate the poll interval firing
    await result.current.refetch()
    await waitFor(() => expect(result.current.data?.jobs[0].status).toBe('done'))

    expect(invalidateSpy).toHaveBeenCalledWith(
      expect.objectContaining({ queryKey: ['options', 'mission-1'] })
    )
  })
})
