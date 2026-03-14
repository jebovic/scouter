import { z } from 'zod'
import { apiFetch } from './client'
import type {
  ShoppingItem,
  PriceSnapshot,
  ShoppingItemCreateRequest,
  ShoppingItemUpdateRequest,
  PriceSnapshotRequest,
  DealScore,
} from '../types'

// ── Zod schemas ─────────────────────────────────────────────────────────────

export const ShoppingItemSchema = z.object({
  id: z.string().uuid(),
  missionId: z.string().uuid(),
  name: z.string(),
  merchant: z.string(),
  costCategory: z.string(),
  price: z.number(),
  originalEstimate: z.number().nullish(),
  targetPrice: z.number().nullish(),
  status: z.enum(['buy', 'flash-sale', 'preorder', 'defer', 'watch', 'crisis', 'recommended', 'rejected']),
  note: z.string().nullish(),
  url: z.string().nullish(),
  pinned: z.boolean().default(false),
  createdAt: z.string(),
})

export const DealScoreSchema = z.object({
  pctBelowAvg: z.number(),
  pctBelowTarget: z.number(),
  historicalAvg: z.number(),
  trend: z.enum(['dropping', 'stable', 'rising']),
  snapshotCount: z.number(),
})

export const PriceSnapshotSchema = z.object({
  id: z.string().uuid(),
  itemId: z.string().uuid(),
  price: z.number(),
  recordedAt: z.string(),
  note: z.string().nullish(),
})

// ── API functions ────────────────────────────────────────────────────────────

export async function listShoppingItems(missionId: string): Promise<ShoppingItem[]> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/shopping`)
  const { items } = z.object({ items: z.array(ShoppingItemSchema) }).parse(data)
  return items as ShoppingItem[]
}

export async function getShoppingItem(missionId: string, itemId: string): Promise<ShoppingItem> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/shopping/${itemId}`)
  return ShoppingItemSchema.parse(data) as ShoppingItem
}

export async function createShoppingItem(
  missionId: string,
  req: ShoppingItemCreateRequest,
): Promise<ShoppingItem> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/shopping`, {
    method: 'POST',
    body: JSON.stringify(req),
  })
  return ShoppingItemSchema.parse(data) as ShoppingItem
}

export async function updateShoppingItem(
  missionId: string,
  itemId: string,
  req: ShoppingItemUpdateRequest,
): Promise<ShoppingItem> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/shopping/${itemId}`, {
    method: 'PUT',
    body: JSON.stringify(req),
  })
  return ShoppingItemSchema.parse(data) as ShoppingItem
}

export async function deleteShoppingItem(missionId: string, itemId: string): Promise<void> {
  await apiFetch<void>(`/api/missions/${missionId}/shopping/${itemId}`, { method: 'DELETE' })
}

export async function addPriceSnapshot(
  missionId: string,
  itemId: string,
  req: PriceSnapshotRequest,
): Promise<PriceSnapshot> {
  const data = await apiFetch<unknown>(
    `/api/missions/${missionId}/shopping/${itemId}/snapshots`,
    { method: 'POST', body: JSON.stringify(req) },
  )
  return PriceSnapshotSchema.parse(data) as PriceSnapshot
}

export async function pinShoppingItem(missionId: string, itemId: string): Promise<ShoppingItem> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/shopping/${itemId}/pin`, {
    method: 'PATCH',
  })
  return ShoppingItemSchema.parse(data) as ShoppingItem
}

export async function deletePinnedShoppingItems(missionId: string): Promise<void> {
  await apiFetch<void>(`/api/missions/${missionId}/shopping/pinned`, { method: 'DELETE' })
}

// Snapshots are not paginated — the backend returns a bare array (not the paged envelope).
export async function listPriceSnapshots(
  missionId: string,
  itemId: string,
): Promise<PriceSnapshot[]> {
  const data = await apiFetch<unknown>(
    `/api/missions/${missionId}/shopping/${itemId}/snapshots`,
  )
  return z.array(PriceSnapshotSchema).parse(data) as PriceSnapshot[]
}

// getDealScore returns null when there are insufficient price snapshots.
export async function getDealScore(
  missionId: string,
  itemId: string,
): Promise<DealScore | null> {
  const data = await apiFetch<unknown>(
    `/api/missions/${missionId}/shopping/${itemId}/deal-score`,
  )
  if (data === null || data === undefined) return null
  return DealScoreSchema.parse(data) as DealScore
}
