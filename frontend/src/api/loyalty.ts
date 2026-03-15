import { z } from 'zod'
import { apiFetch } from './client'

export const LoyaltyProgramSchema = z.object({
  name: z.string(),
  retailer: z.string(),
  pointsRate: z.number(),
  cashbackRate: z.number(),
  minRedemption: z.number(),
  notes: z.string(),
})
export type LoyaltyProgram = z.infer<typeof LoyaltyProgramSchema>

export const LoyaltyCalculateResponseSchema = z.object({
  pointsEarned: z.number(),
  estimatedCashback: z.number(),
  notes: z.string(),
})
export type LoyaltyCalculateResponse = z.infer<typeof LoyaltyCalculateResponseSchema>

export async function getPrograms(): Promise<LoyaltyProgram[]> {
  const data = await apiFetch<unknown>('/api/loyalty/programs')
  return z.array(LoyaltyProgramSchema).parse(data)
}

export async function calculatePoints(
  retailer: string,
  price: number,
  programName: string,
): Promise<LoyaltyCalculateResponse> {
  const data = await apiFetch<unknown>('/api/loyalty/calculate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ retailer, price, programName }),
  })
  return LoyaltyCalculateResponseSchema.parse(data)
}
