import { z } from 'zod'
import { apiFetch } from './client'

export const CurrencyRatesSchema = z.object({
  base: z.string(),
  date: z.string(),
  rates: z.record(z.number()),
})
export type CurrencyRates = z.infer<typeof CurrencyRatesSchema>

export const CurrencyConvertSchema = z.object({
  result: z.number(),
  from: z.string(),
  to: z.string(),
  amount: z.number(),
})
export type CurrencyConvert = z.infer<typeof CurrencyConvertSchema>

export async function getCurrencyRates(): Promise<CurrencyRates> {
  const data = await apiFetch<unknown>('/api/currency/rates')
  return CurrencyRatesSchema.parse(data)
}

export async function convertCurrency(
  amount: number,
  from: string,
  to: string,
): Promise<CurrencyConvert> {
  const params = new URLSearchParams({ from, to, amount: String(amount) })
  const data = await apiFetch<unknown>(`/api/currency/convert?${params}`)
  return CurrencyConvertSchema.parse(data)
}
