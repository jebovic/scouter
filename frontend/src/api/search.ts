import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

const PriceRangeSchema = z.object({
  min: z.number(),
  max: z.number(),
  best: z.number(),
})

export const SearchResultSchema = z.object({
  id: z.string().uuid(),
  missionId: z.string().uuid(),
  missionName: z.string(),
  missionSlug: z.string(),
  name: z.string(),
  category: z.string(),
  badge: z.enum(['recommended', 'alternative', 'rejected', 'watch']),
  priceRange: PriceRangeSchema.nullish(),
  notes: z.string().nullish(),
  similarity: z.number(),
})

export type SearchResult = z.infer<typeof SearchResultSchema>

const SearchResponseSchema = z.object({
  results: z.array(SearchResultSchema),
  query: z.string(),
})

// ── API functions ────────────────────────────────────────────────────────────

export async function searchOptions(query: string, limit = 10): Promise<SearchResult[]> {
  const params = new URLSearchParams({ q: query, limit: String(limit) })
  const data = await apiFetch<unknown>(`/api/search?${params}`)
  const { results } = SearchResponseSchema.parse(data)
  return results
}

export async function getSimilarOptions(optionId: string, limit = 5): Promise<SearchResult[]> {
  const params = new URLSearchParams({ limit: String(limit) })
  const data = await apiFetch<unknown>(`/api/options/${optionId}/similar?${params}`)
  const { results } = SearchResponseSchema.parse(data)
  return results
}

export async function reindexEmbeddings(): Promise<{ queued: number }> {
  const data = await apiFetch<unknown>('/api/search/reindex', { method: 'POST' })
  return z.object({ queued: z.number() }).parse(data)
}
