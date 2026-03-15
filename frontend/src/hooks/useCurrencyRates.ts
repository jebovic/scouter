import { useQuery } from '@tanstack/react-query'
import { fetchExchangeRates, convertCurrency, type ExchangeRates } from '../utils/currencyRates'

interface UseCurrencyRatesResult {
  rates: ExchangeRates | null
  isLoading: boolean
  error: string | null
  convert: (amount: number, from: string, to: string) => number
}

export function useCurrencyRates(): UseCurrencyRatesResult {
  const query = useQuery<ExchangeRates, Error>({
    queryKey: ['exchangeRates'],
    queryFn: () => fetchExchangeRates(),
    staleTime: 4 * 60 * 60 * 1000, // 4 hours
    gcTime: 24 * 60 * 60 * 1000, // 24 hours
    retry: 1,
  })

  const convert = (amount: number, from: string, to: string): number => {
    if (!query.data) return amount
    return convertCurrency(amount, from, to, query.data)
  }

  return {
    rates: query.data ?? null,
    isLoading: query.isLoading,
    error: query.error?.message ?? null,
    convert,
  }
}
