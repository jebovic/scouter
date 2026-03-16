import { renderHook, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, it, expect, vi } from 'vitest'
import React from 'react'

const mockToast = vi.fn()
vi.mock('../components/scouter', () => ({
  useToast: () => ({ toast: mockToast }),
}))

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}))

vi.mock('../api', () => ({
  listMissions: vi.fn().mockResolvedValue([]),
  getMission: vi.fn(),
  createMission: vi.fn(),
  updateMission: vi.fn(),
  deleteMission: vi.fn(),
  archiveMission: vi.fn().mockResolvedValue(undefined),
  unarchiveMission: vi.fn().mockResolvedValue(undefined),
  duplicateMission: vi.fn(),
  cloneMission: vi.fn(),
  triggerResearch: vi.fn(),
  triggerPricing: vi.fn(),
}))

function wrapper({ children }: { children: React.ReactNode }) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return React.createElement(QueryClientProvider, { client: queryClient }, children)
}

describe('useArchiveMission', () => {
  it('exports useArchiveMission hook', async () => {
    const { useArchiveMission } = await import('./useMission')
    expect(typeof useArchiveMission).toBe('function')
  })

  it('calls archiveMission API on mutate', async () => {
    const { archiveMission } = await import('../api')
    const { useArchiveMission } = await import('./useMission')
    const { result } = renderHook(() => useArchiveMission(), { wrapper })
    await act(async () => {
      await result.current.archiveMission('mission-uuid-123')
    })
    expect(archiveMission).toHaveBeenCalledWith('mission-uuid-123')
  })
})

describe('useUnarchiveMission', () => {
  it('exports useUnarchiveMission hook', async () => {
    const { useUnarchiveMission } = await import('./useMission')
    expect(typeof useUnarchiveMission).toBe('function')
  })
})

describe('useMissions with includeArchived', () => {
  it('exports useMissions hook', async () => {
    const { useMissions } = await import('./useMission')
    expect(typeof useMissions).toBe('function')
  })
})
