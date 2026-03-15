import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ──────────────────────────────────────────────────────────────

export const FlightOfferSchema = z.object({
  airline: z.string(),
  flightNum: z.string(),
  departure: z.string(),
  arrival: z.string(),
  departAt: z.string(),
  arriveAt: z.string(),
  priceEur: z.number(),
  currency: z.string(),
  source: z.string(),
})

export const TrainOfferSchema = z.object({
  operator: z.string(),
  trainNum: z.string(),
  departure: z.string(),
  arrival: z.string(),
  departAt: z.string(),
  arriveAt: z.string(),
  durationMin: z.number(),
  priceEur: z.number(),
  source: z.string(),
})

export type FlightOffer = z.infer<typeof FlightOfferSchema>
export type TrainOffer = z.infer<typeof TrainOfferSchema>

// ── API functions ─────────────────────────────────────────────────────────────

export async function searchFlights(from: string, to: string, date: string): Promise<FlightOffer[]> {
  const qs = new URLSearchParams({ from, to, date })
  const data = await apiFetch<unknown>(`/api/travel/flights?${qs}`)
  return z.array(FlightOfferSchema).parse(data)
}

export async function searchTrains(from: string, to: string, date: string): Promise<TrainOffer[]> {
  const qs = new URLSearchParams({ from, to, date })
  const data = await apiFetch<unknown>(`/api/travel/trains?${qs}`)
  return z.array(TrainOfferSchema).parse(data)
}
