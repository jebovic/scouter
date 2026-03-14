import { z } from 'zod'
import { apiFetch } from './client'
import type {
  Mission,
  MissionCreateRequest,
  MissionUpdateRequest,
} from '../types'

// ── Zod schemas ─────────────────────────────────────────────────────────────

const ConstraintSchema = z.object({
  key: z.string(),
  label: z.string(),
  value: z.unknown(),
  type: z.enum(['hard', 'soft']),
})

const TimelineEventSchema = z.object({
  date: z.string(),
  label: z.string(),
  urgent: z.boolean(),
})

const WeightProfileSchema = z.object({
  price: z.number(),
  quality: z.number(),
  feature: z.number(),
})

export const MissionSchema = z.object({
  id: z.string().uuid(),
  slug: z.string(),
  name: z.string(),
  icon: z.string(),
  category: z.enum(['travel', 'electronics', 'computing', 'renovation', 'custom']),
  budget: z.number(),
  currency: z.string(),
  locale: z.string(),
  phase: z.enum(['researching', 'comparing', 'buying', 'done']),
  constraints: z.array(ConstraintSchema),
  costCategories: z.array(z.string()),
  timeline: z.array(TimelineEventSchema),
  weightProfile: WeightProfileSchema.default({ price: 0, quality: 0, feature: 0 }),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export type MissionDTO = z.infer<typeof MissionSchema>

// ── API functions ────────────────────────────────────────────────────────────

export async function listMissions(): Promise<Mission[]> {
  const data = await apiFetch<unknown>('/api/missions')
  const { items } = z.object({ items: z.array(MissionSchema) }).parse(data)
  return items as Mission[]
}

export async function getMission(slug: string): Promise<Mission> {
  const data = await apiFetch<unknown>(`/api/missions/${slug}`)
  return MissionSchema.parse(data) as Mission
}

export async function createMission(req: MissionCreateRequest): Promise<Mission> {
  const data = await apiFetch<unknown>('/api/missions', {
    method: 'POST',
    body: JSON.stringify(req),
  })
  return MissionSchema.parse(data) as Mission
}

export async function updateMission(slug: string, req: MissionUpdateRequest): Promise<Mission> {
  const data = await apiFetch<unknown>(`/api/missions/${slug}`, {
    method: 'PATCH',
    body: JSON.stringify(req),
  })
  return MissionSchema.parse(data) as Mission
}

export async function deleteMission(slug: string): Promise<void> {
  await apiFetch<void>(`/api/missions/${slug}`, { method: 'DELETE' })
}

// items are consumed via dedicated /api/options and /api/shopping endpoints
const AgentResultSchema = z.object({
  count: z.number(),
})

export type AgentResult = z.infer<typeof AgentResultSchema>

export async function triggerResearch(missionId: string, feedback?: string): Promise<AgentResult> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/research`, {
    method: 'POST',
    body: feedback ? JSON.stringify({ feedback }) : undefined,
  })
  return AgentResultSchema.parse(data)
}

export async function triggerPricing(missionId: string, feedback?: string): Promise<AgentResult> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/pricing`, {
    method: 'POST',
    body: feedback ? JSON.stringify({ feedback }) : undefined,
  })
  return AgentResultSchema.parse(data)
}
