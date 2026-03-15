import { z } from 'zod'
import { apiFetch } from './client'

const RuleTypeSchema = z.enum(['drop_percent', 'fixed_below', 'weekly_low'])

export const AlertRuleSchema = z.object({
  id: z.string(),
  itemId: z.string(),
  type: RuleTypeSchema,
  value: z.number(),
  basePrice: z.number(),
  createdAt: z.string(),
  triggered: z.boolean(),
})

export const CheckResultSchema = z.object({
  itemId: z.string(),
  ruleId: z.string(),
  triggered: z.boolean(),
  message: z.string(),
})

export type AlertRule = z.infer<typeof AlertRuleSchema>
export type CheckResult = z.infer<typeof CheckResultSchema>

export async function getAlertRules(itemId: string): Promise<AlertRule[]> {
  const data = await apiFetch<unknown>(`/api/items/${itemId}/alert-rules`)
  return z.array(AlertRuleSchema).parse(data)
}

export async function addAlertRule(
  itemId: string,
  rule: { type: string; value: number; basePrice: number },
): Promise<AlertRule> {
  const data = await apiFetch<unknown>(`/api/items/${itemId}/alert-rules`, {
    method: 'POST',
    body: JSON.stringify(rule),
  })
  return AlertRuleSchema.parse(data)
}

export async function deleteAlertRule(itemId: string, ruleId: string): Promise<void> {
  await apiFetch<unknown>(`/api/items/${itemId}/alert-rules/${ruleId}`, { method: 'DELETE' })
}

export async function checkAlertRules(
  itemId: string,
  currentPrice: number,
  weeklyLow?: number,
): Promise<CheckResult[]> {
  const data = await apiFetch<unknown>(`/api/items/${itemId}/alert-rules/check`, {
    method: 'POST',
    body: JSON.stringify({ currentPrice, weeklyLow }),
  })
  return z.array(CheckResultSchema).parse(data)
}
