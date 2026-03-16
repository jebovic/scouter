import { describe, it, expect, vi, beforeEach } from 'vitest'
import * as client from './client'

vi.mock('./client', () => ({ apiFetch: vi.fn() }))
const mockApiFetch = vi.mocked(client.apiFetch)

beforeEach(() => {
  mockApiFetch.mockReset()
})

describe('archiveMission', () => {
  it('POSTs to /api/missions/:id/archive', async () => {
    mockApiFetch.mockResolvedValueOnce(undefined)
    const { archiveMission } = await import('./missions')
    await archiveMission('mission-uuid-123')
    expect(mockApiFetch).toHaveBeenCalledWith(
      '/api/missions/mission-uuid-123/archive',
      expect.objectContaining({ method: 'POST' })
    )
  })
})

describe('unarchiveMission', () => {
  it('POSTs to /api/missions/:id/unarchive', async () => {
    mockApiFetch.mockResolvedValueOnce(undefined)
    const { unarchiveMission } = await import('./missions')
    await unarchiveMission('mission-uuid-123')
    expect(mockApiFetch).toHaveBeenCalledWith(
      '/api/missions/mission-uuid-123/unarchive',
      expect.objectContaining({ method: 'POST' })
    )
  })
})

describe('listMissions', () => {
  it('passes include_archived param when true', async () => {
    mockApiFetch.mockResolvedValueOnce({ items: [] })
    const { listMissions } = await import('./missions')
    await listMissions({ includeArchived: true })
    expect(mockApiFetch).toHaveBeenCalledWith(
      expect.stringContaining('include_archived=true')
    )
  })

  it('omits include_archived param by default', async () => {
    mockApiFetch.mockResolvedValueOnce({ items: [] })
    const { listMissions } = await import('./missions')
    await listMissions()
    const url = mockApiFetch.mock.calls[0][0] as string
    expect(url).not.toContain('include_archived')
  })
})
