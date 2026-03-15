import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ──────────────────────────────────────────────────────────────

const ItemSavingSchema = z.object({
  itemName: z.string(),
  currentPrice: z.number(),
  targetPrice: z.number(),
  saving: z.number(),
})

export const MissionProgressSchema = z.object({
  missionId: z.string(),
  totalItems: z.number().int(),
  boughtItems: z.number().int(),
  watchingItems: z.number().int(),
  deferredItems: z.number().int(),
  flashSaleItems: z.number().int(),
  completionPct: z.number(),
  budgetUsed: z.number(),
  budgetTotal: z.number(),
  budgetPct: z.number(),
  topSaving: ItemSavingSchema.nullish(),
  generatedAt: z.string(),
})

export type MissionProgress = z.infer<typeof MissionProgressSchema>
export type ItemSaving = z.infer<typeof ItemSavingSchema>

// ── API function ─────────────────────────────────────────────────────────────

export async function fetchMissionProgress(missionId: string): Promise<MissionProgress> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/progress`)
  return MissionProgressSchema.parse(data)
}
