import { z } from 'zod'
import { apiFetch } from './client'

export const DropPredictionSchema = z.object({
  itemId: z.string(),
  probability: z.number(),
  predictedDropPercent: z.number(),
  horizon: z.number().int(),
  confidence: z.enum(['faible', 'moyen', 'élevé']),
  signals: z.array(z.string()),
  generatedAt: z.string(),
})

export type DropPrediction = z.infer<typeof DropPredictionSchema>

export async function getDropPrediction(missionId: string, itemId: string): Promise<DropPrediction> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/items/${itemId}/drop-prediction`)
  return DropPredictionSchema.parse(data)
}
