import { describe, it, expect, vi, beforeEach } from 'vitest'
import { WishListItemSchema, WishListSchema } from './wishlist'

// ── Schema validation tests ──────────────────────────────────────────────────

const validItem = {
  id: 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
  name: 'Sony WH-1000XM5',
  url: 'https://example.com/sony-wh1000xm5',
  targetPrice: 299.99,
  currency: 'EUR',
  notes: 'Wait for Black Friday deal',
  alertEnabled: false,
  lastCheckedAt: null,
  lastCheckedPrice: null,
  createdAt: '2026-01-15T10:00:00Z',
  updatedAt: '2026-01-15T10:00:00Z',
}

describe('WishListItemSchema', () => {
  it('validates a complete valid wish list item', () => {
    expect(() => WishListItemSchema.parse(validItem)).not.toThrow()
    const result = WishListItemSchema.parse(validItem)
    expect(result.id).toBe(validItem.id)
    expect(result.name).toBe('Sony WH-1000XM5')
    expect(result.targetPrice).toBe(299.99)
    expect(result.currency).toBe('EUR')
  })

  it('allows null url', () => {
    const item = { ...validItem, url: null }
    expect(() => WishListItemSchema.parse(item)).not.toThrow()
    const result = WishListItemSchema.parse(item)
    expect(result.url).toBeNull()
  })

  it('allows null targetPrice', () => {
    const item = { ...validItem, targetPrice: null }
    expect(() => WishListItemSchema.parse(item)).not.toThrow()
    const result = WishListItemSchema.parse(item)
    expect(result.targetPrice).toBeNull()
  })

  it('allows null notes', () => {
    const item = { ...validItem, notes: null }
    expect(() => WishListItemSchema.parse(item)).not.toThrow()
    const result = WishListItemSchema.parse(item)
    expect(result.notes).toBeNull()
  })

  it('validates item with all nullable fields null', () => {
    const minimal = { ...validItem, url: null, targetPrice: null, notes: null }
    expect(() => WishListItemSchema.parse(minimal)).not.toThrow()
  })

  it('rejects missing name', () => {
    const { name: _, ...rest } = validItem
    expect(() => WishListItemSchema.parse(rest)).toThrow()
  })

  it('rejects missing id', () => {
    const { id: _, ...rest } = validItem
    expect(() => WishListItemSchema.parse(rest)).toThrow()
  })

  it('rejects missing currency', () => {
    const { currency: _, ...rest } = validItem
    expect(() => WishListItemSchema.parse(rest)).toThrow()
  })

  it('rejects non-numeric targetPrice', () => {
    expect(() => WishListItemSchema.parse({ ...validItem, targetPrice: 'cheap' })).toThrow()
  })

  it('rejects non-string name', () => {
    expect(() => WishListItemSchema.parse({ ...validItem, name: 123 })).toThrow()
  })

  it('rejects completely empty payload', () => {
    expect(() => WishListItemSchema.parse({})).toThrow()
  })

  it('rejects null payload', () => {
    expect(() => WishListItemSchema.parse(null)).toThrow()
  })

  it('validates zero targetPrice', () => {
    const item = { ...validItem, targetPrice: 0 }
    expect(() => WishListItemSchema.parse(item)).not.toThrow()
    const result = WishListItemSchema.parse(item)
    expect(result.targetPrice).toBe(0)
  })
})

describe('WishListSchema', () => {
  it('validates an array of wish list items', () => {
    expect(() => WishListSchema.parse([validItem])).not.toThrow()
    const result = WishListSchema.parse([validItem])
    expect(result).toHaveLength(1)
  })

  it('validates an empty array', () => {
    expect(() => WishListSchema.parse([])).not.toThrow()
    const result = WishListSchema.parse([])
    expect(result).toHaveLength(0)
  })

  it('validates multiple items', () => {
    const items = [validItem, { ...validItem, id: 'bbbb-cccc', name: 'iPad Pro' }]
    expect(() => WishListSchema.parse(items)).not.toThrow()
    const result = WishListSchema.parse(items)
    expect(result).toHaveLength(2)
  })

  it('rejects null', () => {
    expect(() => WishListSchema.parse(null)).toThrow()
  })

  it('rejects non-array', () => {
    expect(() => WishListSchema.parse(validItem)).toThrow()
  })
})

// ── API fetch function tests ─────────────────────────────────────────────────

describe('fetchWishList', () => {
  beforeEach(() => {
    vi.resetAllMocks()
  })

  it('fetches and parses wish list items', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve([validItem]),
    }))

    const { fetchWishList } = await import('./wishlist')
    const result = await fetchWishList()
    expect(result).toHaveLength(1)
    expect(result[0].name).toBe('Sony WH-1000XM5')
  })

  it('throws ApiError on non-ok response', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: false,
      status: 404,
      statusText: 'Not Found',
      json: () => Promise.resolve({ error: 'not found' }),
    }))

    const { fetchWishList } = await import('./wishlist')
    await expect(fetchWishList()).rejects.toThrow()
  })
})

describe('createWishListItem', () => {
  beforeEach(() => {
    vi.resetAllMocks()
  })

  it('posts and returns a new wish list item', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve(validItem),
    }))

    const { createWishListItem } = await import('./wishlist')
    const result = await createWishListItem({ name: 'Sony WH-1000XM5', currency: 'EUR' })
    expect(result.name).toBe('Sony WH-1000XM5')
    expect(result.currency).toBe('EUR')
  })

  it('sends POST with JSON body', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve(validItem),
    })
    vi.stubGlobal('fetch', mockFetch)

    const { createWishListItem } = await import('./wishlist')
    await createWishListItem({ name: 'Test Item', currency: 'USD', targetPrice: 100 })

    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/wishlist'),
      expect.objectContaining({ method: 'POST' }),
    )
  })
})

describe('deleteWishListItem', () => {
  beforeEach(() => {
    vi.resetAllMocks()
  })

  it('sends DELETE to the correct endpoint', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 204,
      json: () => Promise.resolve(undefined),
    })
    vi.stubGlobal('fetch', mockFetch)

    const { deleteWishListItem } = await import('./wishlist')
    await deleteWishListItem('some-uuid')

    expect(mockFetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/wishlist/some-uuid'),
      expect.objectContaining({ method: 'DELETE' }),
    )
  })

  it('resolves without error on success', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      status: 204,
      json: () => Promise.resolve(undefined),
    }))

    const { deleteWishListItem } = await import('./wishlist')
    await expect(deleteWishListItem('some-uuid')).resolves.toBeUndefined()
  })
})
