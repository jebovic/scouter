import { useSettings } from './useSettings'
import { formatCurrency } from '../utils/currency'

export function useFormatCurrency() {
  const { data: settings } = useSettings()
  const currency = settings?.currency ?? 'EUR'
  const locale = settings?.locale ?? 'fr-FR'
  return {
    fmt: (amount: number) => formatCurrency(amount, currency, locale),
    currency,
    locale,
  }
}
