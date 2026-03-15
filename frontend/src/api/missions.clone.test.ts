import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { cloneMission, MissionSchema } from './missions'

// Minimal valid Mission shape that satisfies MissionSchema.
const validMission = {
  id: 'f47ac10b-58cc-4372-a567-0e02b2c3d479',
  slug: 'gaming-pc-copie',
  name: 'Gaming PC (Copie)',
  icon: 'target',
  category: 'electronics' as const,
  budget: 1500,
  currency: 'EUR',
  locale: 'fr-FR',
  phase: 'researching' as const,
  constraints: [],
  costCategories: [],
  timeline: [],
  weightProfile: { price: 0, quality: 0, feature: 0 },
  createdAt: '2026-03-15T00:00:00Z',
  updatedAt: '2026-03-15T00:00:00Z',
}

describe('cloneMission', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('returns a parsed Mission on 201', async () => {
    const mockFetch = vi.mocked(fetch)
    mockFetch.mockResolvedValueOnce(
      new Response(JSON.stringify(validMission), {
        status: 201,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    const result = await cloneMission('gaming-pc')

    expect(result.slug).toBe('gaming-pc-copie')
    expect(result.name).toBe('Gaming PC (Copie)')
    expect(result.phase).toBe('researching')
  })

  it('calls POST /api/missions/{slug}/clone', async () => {
    const mockFetch = vi.mocked(fetch)
    mockFetch.mockResolvedValueOnce(
      new Response(JSON.stringify(validMission), {
        status: 201,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    await cloneMission('gaming-pc')

    expect(mockFetch).toHaveBeenCalledOnce()
    const [url, opts] = mockFetch.mock.calls[0] as [string, RequestInit]
    expect(url).toContain('/api/missions/gaming-pc/clone')
    expect(opts.method).toBe('POST')
  })

  it('throws ApiError on non-200 response', async () => {
    const mockFetch = vi.mocked(fetch)
    mockFetch.mockResolvedValueOnce(
      new Response(JSON.stringify({ error: 'mission not found' }), {
        status: 404,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    await expect(cloneMission('nonexistent')).rejects.toThrow('mission not found')
  })

  it('throws on malformed response body', async () => {
    const mockFetch = vi.mocked(fetch)
    mockFetch.mockResolvedValueOnce(
      new Response(JSON.stringify({ bad: 'shape' }), {
        status: 201,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    await expect(cloneMission('gaming-pc')).rejects.toThrow()
  })
})

describe('MissionSchema — clone fields', () => {
  it('accepts a cloned mission with nullish sensitive fields', () => {
    const payload = { ...validMission, lessons: null, shareToken: null, archivedAt: null }
    expect(() => MissionSchema.parse(payload)).not.toThrow()
    const parsed = MissionSchema.parse(payload)
    expect(parsed.lessons).toBeUndefined()
  })

  it('slug containing "copie" is valid', () => {
    const payload = { ...validMission, slug: 'gaming-pc-copie-ab12' }
    expect(() => MissionSchema.parse(payload)).not.toThrow()
  })

  it('phase "researching" is valid', () => {
    const payload = { ...validMission, phase: 'researching' as const }
    expect(() => MissionSchema.parse(payload)).not.toThrow()
    expect(MissionSchema.parse(payload).phase).toBe('researching')
  })

  it('rejects invalid phase', () => {
    const payload = { ...validMission, phase: 'cloning' }
    expect(() => MissionSchema.parse(payload)).toThrow()
  })
})
