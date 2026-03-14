import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

const DiffEntrySchema = z.object({
  id: z.string().optional(),
  name: z.string(),
  status: z.string(),
})

const DiffSchema = z.object({
  added: z.array(DiffEntrySchema),
  removed: z.array(DiffEntrySchema),
  changed: z.array(DiffEntrySchema),
})

export const AgentRunSchema = z.object({
  id: z.string().uuid(),
  missionId: z.string().uuid(),
  agentType: z.enum(['research', 'pricing']),
  inputSnapshot: z.unknown(),
  resultSnapshot: z.unknown(),
  diff: DiffSchema.nullable(),
  feedback: z.string().optional(),
  createdAt: z.string(),
})

export type AgentRunDTO = z.infer<typeof AgentRunSchema>

export type DiffEntry = z.infer<typeof DiffEntrySchema>
export type AgentDiff = z.infer<typeof DiffSchema>

// ── API functions ────────────────────────────────────────────────────────────

export async function listAgentRuns(
  missionId: string,
  agentType?: 'research' | 'pricing',
): Promise<AgentRunDTO[]> {
  const params = agentType ? `?type=${agentType}` : ''
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/agent-runs${params}`)
  const { runs } = z.object({ runs: z.array(AgentRunSchema) }).parse(data)
  return runs
}
