import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import {
  encodeWishlist,
  decodeWishlist,
  buildShareUrl,
  copyShareLink,
} from './wishlistShare'
import type { WishListItem } from '../api/wishlist'

// Minimal WishListItem fixture
function makeItem(overrides: Partial<WishListItem> = {}): WishListItem {
  return {
    id: 'abc-123',
    name: 'Test Item',
    url: 'https://example.com',
    targetPrice: 49.99,
    currency: 'EUR',
    notes: null,
    alertEnabled: false,
    lastCheckedAt: null,
    lastCheckedPrice: null,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  }
}

describe('wishlistShare', () => {
  describe('encodeWishlist / decodeWishlist — roundtrip', () => {
    it('roundtrips a list of items', () => {
      const items = [
        makeItem({ name: 'Laptop', targetPrice: 999, currency: 'USD' }),
        makeItem({ id: 'b', name: 'Mouse', url: null, targetPrice: null }),
      ]
      const encoded = encodeWishlist(items)
      const decoded = decodeWishlist(encoded)
      expect(decoded).toHaveLength(2)
      expect(decoded[0].n).toBe('Laptop')
      expect(decoded[0].p).toBe(999)
      expect(decoded[0].c).toBe('USD')
      expect(decoded[1].n).toBe('Mouse')
      expect(decoded[1].u).toBeNull()
      expect(decoded[1].p).toBeNull()
    })

    it('roundtrips an empty list', () => {
      const encoded = encodeWishlist([])
      const decoded = decodeWishlist(encoded)
      expect(decoded).toEqual([])
    })

    it('compacts field names (no full field names in encoded string)', () => {
      const items = [makeItem({ name: 'Bike' })]
      const encoded = encodeWishlist(items)
      // The encoded string should not contain the full field name "name"
      // because we compact to single-char keys
      const raw = atob(encoded)
      expect(raw).not.toContain('"name"')
      expect(raw).toContain('"n"')
    })
  })

  describe('decodeWishlist — error handling', () => {
    it('returns empty array for empty string', () => {
      expect(decodeWishlist('')).toEqual([])
    })

    it('returns empty array for malformed base64', () => {
      expect(decodeWishlist('not-valid-base64!!!')).toEqual([])
    })

    it('returns empty array for valid base64 but invalid JSON', () => {
      const invalid = btoa('{ not json }')
      expect(decodeWishlist(invalid)).toEqual([])
    })

    it('returns empty array when decoded JSON is not an array', () => {
      const obj = btoa(JSON.stringify({ foo: 'bar' }))
      expect(decodeWishlist(obj)).toEqual([])
    })
  })

  describe('buildShareUrl', () => {
    beforeEach(() => {
      // Mock window.location.origin
      Object.defineProperty(window, 'location', {
        value: { origin: 'https://scouter.app' },
        writable: true,
      })
    })

    it('builds a URL with /wishlist/shared path and data query param', () => {
      const items = [makeItem({ name: 'Camera' })]
      const url = buildShareUrl(items)
      expect(url).toMatch(/^https:\/\/scouter\.app\/wishlist\/shared\?data=/)
    })

    it('encoded data in URL is decodable', () => {
      const items = [makeItem({ name: 'Tripod', targetPrice: 79, currency: 'GBP' })]
      const url = buildShareUrl(items)
      const params = new URL(url).searchParams
      const data = params.get('data')
      expect(data).not.toBeNull()
      const decoded = decodeWishlist(data!)
      expect(decoded[0].n).toBe('Tripod')
    })
  })

  describe('copyShareLink', () => {
    afterEach(() => {
      vi.restoreAllMocks()
    })

    it('calls navigator.clipboard.writeText with the share URL', async () => {
      const writeMock = vi.fn().mockResolvedValue(undefined)
      Object.defineProperty(navigator, 'clipboard', {
        value: { writeText: writeMock },
        writable: true,
      })
      Object.defineProperty(window, 'location', {
        value: { origin: 'https://scouter.app' },
        writable: true,
      })
      const items = [makeItem({ name: 'Headphones' })]
      await copyShareLink(items)
      expect(writeMock).toHaveBeenCalledOnce()
      expect(writeMock.mock.calls[0][0]).toMatch(/\/wishlist\/shared\?data=/)
    })
  })
})
