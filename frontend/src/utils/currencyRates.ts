export const SUPPORTED_CURRENCIES = ['EUR', 'USD', 'GBP', 'CHF', 'CAD', 'JPY', 'SEK', 'NOK', 'DKK', 'PLN', 'CZK'] as const

export type SupportedCurrency = (typeof SUPPORTED_CURRENCIES)[number]

export interface ExchangeRates {
  base: 'EUR'
  rates: Record<string, number>
  fetchedAt: number // unix ms
}

const ECB_API_URL =
  'https://data-api.ecb.europa.eu/service/data/EXR/D.USD+GBP+CHF+CAD+JPY+SEK+NOK+DKK+PLN+CZK.EUR.SP00.A?lastNObservations=1&format=jsondata'

const CACHE_KEY = 'scouter_fx_rates'
const CACHE_DURATION_MS = 4 * 60 * 60 * 1000 // 4 hours

/**
 * Fetch exchange rates from ECB API with 4-hour localStorage cache
 */
export async function fetchExchangeRates(): Promise<ExchangeRates> {
  // Check cache first
  const cached = localStorage.getItem(CACHE_KEY)
  if (cached) {
    try {
      const parsed = JSON.parse(cached) as ExchangeRates
      const age = Date.now() - parsed.fetchedAt
      if (age < CACHE_DURATION_MS) {
        return parsed
      }
    } catch {
      // Invalid cache, continue to fetch
    }
  }

  try {
    const response = await fetch(ECB_API_URL)
    if (!response.ok) throw new Error('ECB API error')

    const data = await response.json()

    // Parse ECB response: data.data.dataSets[0].series[*:0].observations
    const series = data.data?.dataSets?.[0]?.series || {}
    const rates: Record<string, number> = { EUR: 1.0 }

    for (const key in series) {
      // Key format: "0:0" (first entry), "1:0" (second), etc.
      const observations = series[key]?.observations?.[0]
      if (observations !== undefined && typeof observations === 'number') {
        // Extract currency from dimension
        const currencyIndex = parseInt(key.split(':')[0], 10)
        const currencies = ['USD', 'GBP', 'CHF', 'CAD', 'JPY', 'SEK', 'NOK', 'DKK', 'PLN', 'CZK']
        if (currencyIndex >= 0 && currencyIndex < currencies.length) {
          rates[currencies[currencyIndex]] = observations
        }
      }
    }

    const result: ExchangeRates = {
      base: 'EUR',
      rates,
      fetchedAt: Date.now(),
    }

    // Cache it
    localStorage.setItem(CACHE_KEY, JSON.stringify(result))
    return result
  } catch (error) {
    console.error('Failed to fetch exchange rates:', error)

    // Fallback: return cached data even if stale
    const cached = localStorage.getItem(CACHE_KEY)
    if (cached) {
      try {
        return JSON.parse(cached) as ExchangeRates
      } catch {
        // Ignore parse error
      }
    }

    // Final fallback: return hardcoded EUR rates if everything fails
    return {
      base: 'EUR',
      rates: {
        EUR: 1.0,
        USD: 1.08,
        GBP: 0.86,
        CHF: 0.95,
        CAD: 1.45,
        JPY: 155,
        SEK: 11.2,
        NOK: 11.5,
        DKK: 7.45,
        PLN: 4.35,
        CZK: 24.5,
      },
      fetchedAt: Date.now(),
    }
  }
}

/**
 * Convert an amount from one currency to another
 */
export function convertCurrency(amount: number, from: string, to: string, rates: ExchangeRates): number {
  if (from === to) return amount
  if (amount === 0) return 0

  const fromRate = rates.rates[from] || 1
  const toRate = rates.rates[to] || 1

  // Convert through EUR base: from -> EUR -> to
  const inEur = amount / fromRate
  const converted = inEur * toRate
  return converted
}

/**
 * Format a currency amount with Intl.NumberFormat
 */
export function formatCurrency(amount: number, currency: string, locale = 'fr-FR'): string {
  try {
    return new Intl.NumberFormat(locale, {
      style: 'currency',
      currency,
      maximumFractionDigits: 2,
    }).format(amount)
  } catch {
    // Fallback for unsupported locale/currency
    return `${amount} ${currency}`
  }
}
