import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

export const PurchaseRecordSchema = z.object({
  id: z.string().uuid(),
  missionId: z.string().uuid(),
  optionId: z.string().uuid().nullish(),
  purchasedAt: z.string(),
  finalPrice: z.number(),
  merchant: z.string(),
  satisfaction: z.number().min(1).max(5).nullish(),
  review: z.string().nullish(),
  createdAt: z.string(),
})

export type PurchaseRecord = z.infer<typeof PurchaseRecordSchema>

export const CategoryStatSchema = z.object({
  category: z.string(),
  count: z.number(),
  totalSpent: z.number(),
  avgSatisfaction: z.number().nullish(),
})

export const StatsSchema = z.object({
  totalSpent: z.number(),
  totalBudget: z.number(),
  savings: z.number(),
  purchaseCount: z.number(),
  categoryBreakdown: z.array(CategoryStatSchema),
})

export type Stats = z.infer<typeof StatsSchema>
export type CategoryStat = z.infer<typeof CategoryStatSchema>

export type CreatePurchaseRequest = {
  optionId?: string
  purchasedAt: string
  finalPrice: number
  merchant: string
  satisfaction?: number
  review?: string
}

export type UpdatePurchaseRequest = {
  satisfaction?: number
  review?: string
  merchant?: string
  finalPrice?: number
}

// ── API functions ────────────────────────────────────────────────────────────

export async function getPurchaseRecord(missionId: string): Promise<PurchaseRecord | null> {
  try {
    const data = await apiFetch<unknown>(`/api/missions/${missionId}/purchase`)
    return PurchaseRecordSchema.parse(data)
  } catch (e: unknown) {
    if (e instanceof Error && 'status' in e && (e as { status: number }).status === 404) return null
    throw e
  }
}

export async function createPurchaseRecord(missionId: string, req: CreatePurchaseRequest): Promise<PurchaseRecord> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/purchase`, {
    method: 'POST',
    body: JSON.stringify(req),
  })
  return PurchaseRecordSchema.parse(data)
}

export async function updatePurchaseRecord(missionId: string, req: UpdatePurchaseRequest): Promise<PurchaseRecord> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/purchase`, {
    method: 'PATCH',
    body: JSON.stringify(req),
  })
  return PurchaseRecordSchema.parse(data)
}

export async function getStats(): Promise<Stats> {
  const data = await apiFetch<unknown>('/api/stats')
  return StatsSchema.parse(data)
}

// ── Monthly stats ─────────────────────────────────────────────────────────────

export const MonthlyStatPointSchema = z.object({
  month: z.string(),
  totalSpent: z.number(),
  totalBudget: z.number(),
  purchaseCount: z.number(),
})

export type MonthlyStatPoint = z.infer<typeof MonthlyStatPointSchema>

export const MonthlyStatsSchema = z.array(MonthlyStatPointSchema)

export async function fetchMonthlyStats(): Promise<MonthlyStatPoint[]> {
  const res = await fetch('/api/stats/monthly')
  if (!res.ok) throw new Error('Failed to fetch monthly stats')
  return MonthlyStatsSchema.parse(await res.json())
}
