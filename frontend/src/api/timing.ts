import { z } from 'zod'
import { apiFetch } from './client'

export const TimingAdviceSchema = z.object({
  missionId: z.string(),
  recommendation: z.enum(['BUY_NOW', 'WAIT', 'UNSURE']),
  rationale: z.string(),
  waitUntil: z.string().optional(),
  nextSaleEvent: z.string().optional(),
  confidence: z.number(),
  generatedAt: z.string(),
})

export type TimingAdvice = z.infer<typeof TimingAdviceSchema>

export async function getTimingAdvice(missionId: string): Promise<TimingAdvice> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/timing-advice`, {
    method: 'POST',
  })
  return TimingAdviceSchema.parse(data)
}
