import { useState, useCallback } from 'react'
import { useCurrencyRates } from '../../hooks/useCurrencyRates'
import { formatCurrencyLocale, formatDate } from '../../utils/format'
import { SUPPORTED_CURRENCIES } from '../../utils/currencyRates'
import styles from './CurrencyConverter.module.css'

const CURRENCY_FLAGS: Record<string, string> = {
  EUR: '🇪🇺',
  USD: '🇺🇸',
  GBP: '🇬🇧',
  CHF: '🇨🇭',
  CAD: '🇨🇦',
  JPY: '🇯🇵',
  SEK: '🇸🇪',
  NOK: '🇳🇴',
  DKK: '🇩🇰',
  PLN: '🇵🇱',
  CZK: '🇨🇿',
}

export function CurrencyConverter() {
  const { rates, isLoading, convert } = useCurrencyRates()
  const [amount, setAmount] = useState<string>('100')
  const [fromCurrency, setFromCurrency] = useState('EUR')
  const [toCurrency, setToCurrency] = useState('USD')

  const handleSwap = useCallback(() => {
    setFromCurrency(toCurrency)
    setToCurrency(fromCurrency)
  }, [fromCurrency, toCurrency])

  const numAmount = parseFloat(amount) || 0
  const convertedAmount = convert(numAmount, fromCurrency, toCurrency)

  const getRate = useCallback(() => {
    if (!rates) return null
    const fromRate = rates.rates[fromCurrency] || 1
    const toRate = rates.rates[toCurrency] || 1
    return (toRate / fromRate).toFixed(6)
  }, [rates, fromCurrency, toCurrency])

  const rate = getRate()

  return (
    <div className={styles.converter}>
      <h3 className={styles.title}>💱 Convertisseur de Devises</h3>

      <div className={styles.inputGroup}>
        <input
          type="number"
          className={styles.amountInput}
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          placeholder="100"
          disabled={isLoading}
        />
        <select
          className={styles.currencySelect}
          value={fromCurrency}
          onChange={(e) => setFromCurrency(e.target.value)}
          disabled={isLoading}
        >
          {SUPPORTED_CURRENCIES.map((curr) => (
            <option key={curr} value={curr}>
              {CURRENCY_FLAGS[curr] || ''} {curr}
            </option>
          ))}
        </select>
        <button
          className={styles.swapButton}
          onClick={handleSwap}
          disabled={isLoading}
          title="Inverser les devises"
        >
          ⇄
        </button>
        <select
          className={styles.currencySelect}
          value={toCurrency}
          onChange={(e) => setToCurrency(e.target.value)}
          disabled={isLoading}
        >
          {SUPPORTED_CURRENCIES.map((curr) => (
            <option key={curr} value={curr}>
              {CURRENCY_FLAGS[curr] || ''} {curr}
            </option>
          ))}
        </select>
      </div>

      <div className={styles.divider} />

      <div className={styles.resultSection}>
        <div className={styles.rateLabel}>Taux du jour</div>
        <div className={styles.resultValue}>
          {isLoading ? '...' : formatCurrencyLocale(convertedAmount, 'fr-FR', toCurrency)}
        </div>

        {rate && !isLoading && (
          <div className={styles.rateInfo}>
            <span>1 {fromCurrency}</span>
            <span className={styles.rateValue}>= {rate} {toCurrency}</span>
          </div>
        )}

        <div className={styles.sourceCredit}>Source: BCE (Banque Centrale Européenne)</div>

        {rates && (
          <div className={styles.lastUpdated}>
            Mis à jour: {formatDate(new Date(rates.fetchedAt), 'fr-FR')}
          </div>
        )}
      </div>
    </div>
  )
}
