import { z } from 'zod'

export const AggregatedVotesSchema = z.object({
  optionId: z.string(),
  upvotes: z.number(),
  downvotes: z.number(),
  score: z.number(),
  byCollaborator: z.record(z.string(), z.number()).default({}),
})

export type AggregatedVotes = z.infer<typeof AggregatedVotesSchema>

const API_BASE = import.meta.env.VITE_API_URL ?? '/api'

export async function recordVote(
  optionId: string,
  collaboratorId: string,
  vote: number,
): Promise<void> {
  const res = await fetch(`${API_BASE}/options/${optionId}/votes`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ collaboratorId, vote }),
  })
  if (!res.ok) throw new Error('Failed to record vote')
}

export async function getVotes(optionId: string): Promise<AggregatedVotes> {
  const res = await fetch(`${API_BASE}/options/${optionId}/votes`)
  if (!res.ok) throw new Error('Failed to load votes')
  return AggregatedVotesSchema.parse(await res.json())
}
