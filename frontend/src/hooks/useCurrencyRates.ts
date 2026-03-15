import { useQuery } from '@tanstack/react-query'
import { getCurrencyRates, type CurrencyRates } from '../api/currency'

interface UseCurrencyRatesResult {
  rates: CurrencyRates | null
  isLoading: boolean
  error: string | null
  convert: (amount: number, from: string, to: string) => number
}

export function useCurrencyRates(): UseCurrencyRatesResult {
  const query = useQuery<CurrencyRates, Error>({
    queryKey: ['currencyRates'],
    queryFn: getCurrencyRates,
    staleTime: 24 * 60 * 60 * 1000, // 24h — rates don't change often
    gcTime: 24 * 60 * 60 * 1000,
    retry: false,
  })

  const convert = (amount: number, from: string, to: string): number => {
    const rates = query.data
    if (!rates) return amount
    if (from === to) return amount
    const fromRate = rates.rates[from] ?? 1
    const toRate = rates.rates[to] ?? 1
    // Convert through EUR base: from -> EUR -> to
    const inEUR = amount / fromRate
    return inEUR * toRate
  }

  return {
    rates: query.data ?? null,
    isLoading: query.isLoading,
    error: query.error?.message ?? null,
    convert,
  }
}
