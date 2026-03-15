import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

export const DuplicateItemSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  price: z.number(),
  status: z.string(),
  merchant: z.string(),
})

export const DuplicateGroupSchema = z.object({
  items: z.array(DuplicateItemSchema),
  similarity: z.number(),
  reason: z.string(),
  recommended: z.string(),
})

export const DuplicateReportSchema = z.object({
  missionId: z.string(),
  groups: z.array(DuplicateGroupSchema),
  totalDupes: z.number(),
  generatedAt: z.string(),
})

export type DuplicateItem = z.infer<typeof DuplicateItemSchema>
export type DuplicateGroup = z.infer<typeof DuplicateGroupSchema>
export type DuplicateReport = z.infer<typeof DuplicateReportSchema>

// ── API function ─────────────────────────────────────────────────────────────

export async function getDuplicates(missionId: string): Promise<DuplicateReport> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/duplicates`)
  return DuplicateReportSchema.parse(data)
}
