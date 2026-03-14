import { z } from 'zod'
import { apiFetch } from './client'

const ProviderStatSchema = z.object({
  provider: z.string(),
  model: z.string(),
  calls: z.number(),
  input_tokens: z.number(),
  output_tokens: z.number(),
})

const UsageSummarySchema = z.object({
  period: z.string(),
  by_provider: z.array(ProviderStatSchema),
  total_input_tokens: z.number(),
  total_output_tokens: z.number(),
  fallback_calls: z.number(),
})

export type UsageSummary = z.infer<typeof UsageSummarySchema>
export type ProviderStat = z.infer<typeof ProviderStatSchema>

const VALID_PERIODS = ['24h', '7d', '30d'] as const
export type UsagePeriod = (typeof VALID_PERIODS)[number]

export async function getUsageSummary(period: UsagePeriod = '7d'): Promise<UsageSummary> {
  if (!VALID_PERIODS.includes(period)) throw new Error(`Invalid period: ${period}`)
  const data = await apiFetch<unknown>(`/api/usage?period=${period}`)
  return UsageSummarySchema.parse(data)
}
