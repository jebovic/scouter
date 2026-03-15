import { z } from 'zod'
import { apiFetch } from './client'

const ItemEcoScoreSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  category: z.string(),
  co2kg: z.number(),
  grade: z.enum(['A', 'B', 'C', 'D', 'E']),
})

export const EcoScoreSchema = z.object({
  missionId: z.string(),
  totalCO2kg: z.number(),
  ecoGrade: z.enum(['A', 'B', 'C', 'D', 'E']),
  gradeColor: z.string(),
  itemScores: z.array(ItemEcoScoreSchema),
  sustainableShopping: z.array(z.string()),
  ecoTips: z.array(z.string()),
  compensationTreesNeeded: z.number(),
})

export type EcoScore = z.infer<typeof EcoScoreSchema>
export type ItemEcoScore = z.infer<typeof ItemEcoScoreSchema>

export async function getEcoScore(missionId: string): Promise<EcoScore> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/eco-score`)
  return EcoScoreSchema.parse(data)
}
