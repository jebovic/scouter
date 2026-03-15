import { z } from 'zod'

const API_BASE = import.meta.env.VITE_API_URL ?? '/api'

export const VotedItemSchema = z.object({
  itemId: z.string(),
  name: z.string(),
  price: z.number(),
  ups: z.number().int(),
  downs: z.number().int(),
  netScore: z.number().int(),
  totalVotes: z.number().int(),
  approvalRate: z.number(),
  sentiment: z.string(),
})

export type VotedItem = z.infer<typeof VotedItemSchema>

export const VoteSummaryResponseSchema = z.object({
  missionId: z.string(),
  items: z.array(VotedItemSchema),
  mostApproved: VotedItemSchema.nullable(),
  mostControversial: VotedItemSchema.nullable(),
})

export type VoteSummaryResponse = z.infer<typeof VoteSummaryResponseSchema>

export async function getVoteSummary(missionId: string): Promise<VoteSummaryResponse> {
  const res = await fetch(`${API_BASE}/missions/${missionId}/vote-summary`)
  if (!res.ok) throw new Error('Failed to load vote summary')
  return VoteSummaryResponseSchema.parse(await res.json())
}
