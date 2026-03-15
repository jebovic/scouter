import { useState } from 'react'
import { useCurrencyRates } from '../../hooks/useCurrencyRates'
import styles from './CurrencyConverter.module.css'

const TARGET_CURRENCIES = ['EUR', 'USD', 'GBP', 'CHF', 'JPY', 'CAD'] as const

interface CurrencyConverterProps {
  basePrice: number
  baseCurrency: string
}

export function CurrencyConverter({ basePrice, baseCurrency }: CurrencyConverterProps) {
  const [target, setTarget] = useState<string>('EUR')
  const { convert, isLoading, error } = useCurrencyRates()

  const showConverted = target !== baseCurrency

  const formatted = (): string => {
    if (isLoading || error) return '—'
    const result = convert(basePrice, baseCurrency, target)
    try {
      return new Intl.NumberFormat('fr-FR', {
        style: 'currency',
        currency: target,
        maximumFractionDigits: 2,
      }).format(result)
    } catch {
      return `${result.toFixed(2)} ${target}`
    }
  }

  return (
    <div className={styles.wrapper} aria-label="Currency converter">
      <select
        className={styles.select}
        value={target}
        onChange={(e) => setTarget(e.target.value)}
        aria-label="Target currency"
      >
        {TARGET_CURRENCIES.map((c) => (
          <option key={c} value={c}>
            {c}
          </option>
        ))}
      </select>
      {showConverted && (
        <span className={styles.result} aria-live="polite">
          {formatted()}
        </span>
      )}
    </div>
  )
}
