import { z } from 'zod'
import { apiFetch } from './client'
import type { Option, OptionCreateRequest, OptionUpdateRequest } from '../types'

// ── Zod schemas ─────────────────────────────────────────────────────────────

const AttributeSchema = z.object({
  key: z.string(),
  label: z.string(),
  value: z.unknown(),
  type: z.enum(['text', 'price', 'score', 'boolean']),
  max: z.number().optional(),
  pass: z.boolean().optional(),
})

const PriceRangeSchema = z.object({
  min: z.number(),
  max: z.number(),
  best: z.number(),
})

export const OptionSchema = z.object({
  id: z.string().uuid(),
  missionId: z.string().uuid(),
  name: z.string(),
  category: z.string(),
  badge: z.enum(['recommended', 'alternative', 'rejected', 'watch']),
  attributes: z.array(AttributeSchema),
  priceRange: PriceRangeSchema.nullish(),
  notes: z.string().nullish(),
  warnings: z.array(z.string()),
  url: z.string().nullish(),
  createdAt: z.string(),
})

export type OptionDTO = z.infer<typeof OptionSchema>

// ── API functions ────────────────────────────────────────────────────────────

export async function listOptions(missionId: string): Promise<Option[]> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/options`)
  const { items } = z.object({ items: z.array(OptionSchema) }).parse(data)
  return items as Option[]
}

export async function getOption(missionId: string, optionId: string): Promise<Option> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/options/${optionId}`)
  return OptionSchema.parse(data) as Option
}

export async function createOption(missionId: string, req: OptionCreateRequest): Promise<Option> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/options`, {
    method: 'POST',
    body: JSON.stringify(req),
  })
  return OptionSchema.parse(data) as Option
}

export async function updateOption(
  missionId: string,
  optionId: string,
  req: OptionUpdateRequest,
): Promise<Option> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/options/${optionId}`, {
    method: 'PUT',
    body: JSON.stringify(req),
  })
  return OptionSchema.parse(data) as Option
}

export async function deleteOption(missionId: string, optionId: string): Promise<void> {
  await apiFetch<void>(`/api/missions/${missionId}/options/${optionId}`, { method: 'DELETE' })
}
