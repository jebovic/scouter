import { z } from 'zod'
import { apiFetch } from './client'

const ItemScoreSchema = z.object({
  itemId: z.string(),
  name: z.string(),
  price: z.number(),
  priceScore: z.number(),
  urgencyScore: z.number(),
  gapScore: z.number(),
  compositeScore: z.number(),
  rank: z.number(),
  verdict: z.string(),
})

export type ItemScore = z.infer<typeof ItemScoreSchema>

const ComparisonReportSchema = z.object({
  missionId: z.string(),
  itemScores: z.array(ItemScoreSchema),
  topPick: z.string(),
  summary: z.string(),
})

export type ComparisonReport = z.infer<typeof ComparisonReportSchema>

export async function fetchComparisonScore(missionId: string): Promise<ComparisonReport> {
  return apiFetch<ComparisonReport>(`/api/missions/${missionId}/comparison-score`)
}
