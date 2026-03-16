import { renderHook, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, it, expect, vi } from 'vitest'
import React from 'react'

const mockToast = vi.fn()
vi.mock('../components/scouter', () => ({
  useToast: () => ({ toast: mockToast }),
}))

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string, opts?: Record<string, unknown>) => opts ? `${key}:${JSON.stringify(opts)}` : key }),
}))

vi.mock('../api/options', () => ({
  listOptions: vi.fn().mockResolvedValue([]),
  pinOption: vi.fn().mockResolvedValue({ id: 'opt-1', pinned: false }),
  deleteOption: vi.fn(),
  updateOption: vi.fn(),
  rejectOption: vi.fn(),
  unrejectOption: vi.fn(),
  deletePinnedOptions: vi.fn(),
  retranslateOption: vi.fn(),
  getOption: vi.fn(),
  createOption: vi.fn(),
  getOptionsExportURL: vi.fn(),
}))

function wrapper({ children }: { children: React.ReactNode }) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return React.createElement(QueryClientProvider, { client: queryClient }, children)
}

describe('useUnpinAllOptions', () => {
  it('exports useUnpinAllOptions', async () => {
    const { useUnpinAllOptions } = await import('./useOptions')
    expect(typeof useUnpinAllOptions).toBe('function')
  })

  it('calls pinOption for each provided pinned option id', async () => {
    const { pinOption } = await import('../api/options')
    const { useUnpinAllOptions } = await import('./useOptions')
    const { result } = renderHook(() => useUnpinAllOptions('mission-1'), { wrapper })
    await act(async () => {
      await result.current.unpinAllOptions(['opt-1', 'opt-2'])
    })
    expect(pinOption).toHaveBeenCalledWith('mission-1', 'opt-1')
    expect(pinOption).toHaveBeenCalledWith('mission-1', 'opt-2')
  })

  it('shows an error toast when some pinOption calls fail', async () => {
    const { pinOption } = await import('../api/options')
    const { useUnpinAllOptions } = await import('./useOptions')

    const mockPinOption = pinOption as ReturnType<typeof vi.fn>
    mockPinOption
      .mockResolvedValueOnce({ id: 'opt-1', pinned: false })
      .mockRejectedValueOnce(new Error('Network error'))

    mockToast.mockClear()

    const { result } = renderHook(() => useUnpinAllOptions('mission-1'), { wrapper })
    await act(async () => {
      await result.current.unpinAllOptions(['opt-1', 'opt-2'])
    })

    expect(mockToast).toHaveBeenCalledWith(
      expect.stringContaining('option.actions.unpinAllError'),
      'error'
    )
  })
})
