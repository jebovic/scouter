import { z } from 'zod'

const OutcomeSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  merchant: z.string(),
  askingPrice: z.number(),
  negotiatedPrice: z.number().optional(),
  saved: z.number(),
  successRate: z.number(),
  tipForMerchant: z.string(),
  verdict: z.enum(['success', 'partial', 'failed', 'pending']),
  attempts: z.number(),
})

export const NegotiationSummarySchema = z.object({
  outcomes: z.array(OutcomeSchema),
  totalSaved: z.number(),
  avgSuccessRate: z.number(),
  bestMerchant: z.string(),
  totalAttempts: z.number(),
})

export type NegotiationOutcome = z.infer<typeof OutcomeSchema>
export type NegotiationSummary = z.infer<typeof NegotiationSummarySchema>

export async function getNegotiationOutcomes(missionId: string): Promise<NegotiationSummary> {
  const res = await fetch(`/api/missions/${missionId}/negotiation-outcomes`)
  if (!res.ok) throw new Error('Failed to fetch negotiation outcomes')
  return NegotiationSummarySchema.parse(await res.json())
}
