import type { WishListItem } from '../api/wishlist'

// Compact representation for URL encoding
export interface CompactWishlistItem {
  n: string         // name
  p: number | null  // targetPrice
  u: string | null  // url
  c: string         // currency
}

/**
 * Serialize wishlist items to a compact base64 URL-safe string.
 * Uses single-char keys to keep the encoded string short.
 */
export function encodeWishlist(items: WishListItem[]): string {
  try {
    const compact: CompactWishlistItem[] = items.map((item) => ({
      n: item.name,
      p: item.targetPrice ?? null,
      u: item.url ?? null,
      c: item.currency,
    }))
    return btoa(JSON.stringify(compact))
  } catch {
    return btoa('[]')
  }
}

/**
 * Decode a base64 wishlist string back to compact items.
 * Returns empty array on any error.
 */
export function decodeWishlist(encoded: string): CompactWishlistItem[] {
  if (!encoded) return []
  try {
    const json = atob(encoded)
    const parsed: unknown = JSON.parse(json)
    if (!Array.isArray(parsed)) return []
    return parsed as CompactWishlistItem[]
  } catch {
    return []
  }
}

/**
 * Build a shareable URL encoding the wishlist items in the query string.
 */
export function buildShareUrl(items: WishListItem[]): string {
  const encoded = encodeWishlist(items)
  return `${window.location.origin}/wishlist/shared?data=${encoded}`
}

/**
 * Copy the share URL to the clipboard.
 */
export async function copyShareLink(items: WishListItem[]): Promise<void> {
  const url = buildShareUrl(items)
  await navigator.clipboard.writeText(url)
}
