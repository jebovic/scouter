import { useState, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { useCurrencyRates } from '../../hooks/useCurrencyRates'
import { formatCurrencyLocale, formatDate } from '../../utils/format'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
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
  const { t } = useTranslation()
  const { rates, isLoading, convert } = useCurrencyRates()
  const { locale } = useFormatCurrency()
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
      <h3 className={styles.title}>{t('currencyConverter.title')}</h3>

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
          title={t('currencyConverter.swapTitle')}
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
        <div className={styles.rateLabel}>{t('currencyConverter.rateLabel')}</div>
        <div className={styles.resultValue}>
          {isLoading ? '...' : formatCurrencyLocale(convertedAmount, locale, toCurrency)}
        </div>

        {rate && !isLoading && (
          <div className={styles.rateInfo}>
            <span>1 {fromCurrency}</span>
            <span className={styles.rateValue}>= {rate} {toCurrency}</span>
          </div>
        )}

        <div className={styles.sourceCredit}>{t('currencyConverter.source')}</div>

        {rates && (
          <div className={styles.lastUpdated}>
            {t('currencyConverter.lastUpdated')} {formatDate(new Date(rates.date), locale)}
          </div>
        )}
      </div>
    </div>
  )
}
