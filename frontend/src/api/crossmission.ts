import { z } from 'zod'
import { apiFetch } from './client'

export const MissionComparisonSchema = z.object({
  missionId: z.string(),
  name: z.string(),
  slug: z.string(),
  budget: z.number(),
  itemCount: z.number().int(),
  totalPrice: z.number(),
  totalEstimate: z.number(),
  boughtCount: z.number().int(),
  savingsAmount: z.number(),
  savingsPct: z.number(),
  budgetUsedPct: z.number(),
  efficiency: z.enum(['excellent', 'good', 'overspent', 'no data']),
})

export const CrossMissionResponseSchema = z.object({
  missions: z.array(MissionComparisonSchema),
  bestMission: MissionComparisonSchema.nullable(),
  totalSavingsAllMissions: z.number(),
  generatedAt: z.string(),
})

export type MissionComparison = z.infer<typeof MissionComparisonSchema>
export type CrossMissionResponse = z.infer<typeof CrossMissionResponseSchema>

export async function getCrossMission(): Promise<CrossMissionResponse> {
  const data = await apiFetch<unknown>('/api/analytics/cross-mission')
  return CrossMissionResponseSchema.parse(data)
}
