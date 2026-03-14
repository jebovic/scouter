import { z } from 'zod'
import { apiFetch } from './client'

export const ModelStatusSchema = z.object({
  name: z.string(),
  healthy: z.boolean(),
  circuit_state: z.enum(['closed', 'open', 'half_open']),
  last_latency_ms: z.number(),
  capabilities: z.array(z.string()),
})

export const PoolHealthSchema = z.object({
  models: z.array(ModelStatusSchema),
})

export type ModelStatus = z.infer<typeof ModelStatusSchema>
export type PoolHealth = z.infer<typeof PoolHealthSchema>

export async function getLLMHealth(): Promise<PoolHealth> {
  const data = await apiFetch<unknown>('/api/health/llm')
  return PoolHealthSchema.parse(data)
}
