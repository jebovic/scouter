import { z } from 'zod'
import { apiFetch } from './client'

export const ShoppingSummarySchema = z.object({
  id: z.string().uuid(),
  missionId: z.string().uuid(),
  content: z.string(),
  generatedAt: z.string(),
})

export type ShoppingSummaryDTO = z.infer<typeof ShoppingSummarySchema>

export async function generateSummary(missionId: string): Promise<ShoppingSummaryDTO> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/summary`, {
    method: 'POST',
  })
  return ShoppingSummarySchema.parse(data)
}

export async function fetchSummary(missionId: string): Promise<ShoppingSummaryDTO | null> {
  try {
    const data = await apiFetch<unknown>(`/api/missions/${missionId}/summary`)
    return ShoppingSummarySchema.parse(data)
  } catch {
    return null
  }
}
