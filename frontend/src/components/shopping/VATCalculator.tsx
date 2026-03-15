import { useState } from 'react'
import { htToTtc, ttcToHt, vatAmount, VAT_RATES, type VATRate } from '../../utils/vat'
import styles from './VATCalculator.module.css'

interface VATCalculatorProps {
  priceHT?: number
  priceTTC?: number
  onClose: () => void
}

/**
 * French VAT (TVA) calculator widget for converting between HT and TTC prices.
 * Helps users understand pricing with different French VAT rates.
 */
export function VATCalculator({ priceHT = 0, priceTTC = 0, onClose }: VATCalculatorProps) {
  const [selectedRate, setSelectedRate] = useState<VATRate>(0.20)
  const [ht, setHt] = useState<string>(priceHT > 0 ? priceHT.toString() : '')
  const [ttc, setTtc] = useState<string>(priceTTC > 0 ? priceTTC.toString() : '')

  // Get current rate option for description
  const currentRate = VAT_RATES.find((r) => r.rate === selectedRate) || VAT_RATES[0]

  // Handle HT input change — update TTC based on HT
  function handleHTChange(value: string) {
    setHt(value)
    if (value === '' || value === '0') {
      setTtc('')
    } else {
      const htNum = parseFloat(value)
      if (!isNaN(htNum) && htNum > 0) {
        const ttcNum = htToTtc(htNum, selectedRate)
        setTtc(ttcNum.toFixed(2))
      }
    }
  }

  // Handle TTC input change — update HT based on TTC
  function handleTTCChange(value: string) {
    setTtc(value)
    if (value === '' || value === '0') {
      setHt('')
    } else {
      const ttcNum = parseFloat(value)
      if (!isNaN(ttcNum) && ttcNum > 0) {
        const htNum = ttcToHt(ttcNum, selectedRate)
        setHt(htNum.toFixed(2))
      }
    }
  }

  // Handle rate change — recalculate both if either field has value
  function handleRateChange(e: React.ChangeEvent<HTMLSelectElement>) {
    const rate = parseFloat(e.target.value) as VATRate
    setSelectedRate(rate)

    // Recalculate based on whichever field has a value
    if (ht && ht !== '0') {
      const htNum = parseFloat(ht)
      if (!isNaN(htNum) && htNum > 0) {
        const ttcNum = htToTtc(htNum, rate)
        setTtc(ttcNum.toFixed(2))
      }
    } else if (ttc && ttc !== '0') {
      const ttcNum = parseFloat(ttc)
      if (!isNaN(ttcNum) && ttcNum > 0) {
        const htNum = ttcToHt(ttcNum, rate)
        setHt(htNum.toFixed(2))
      }
    }
  }

  // Calculate VAT amount from HT
  const htNum = ht ? parseFloat(ht) : 0
  const vat = htNum > 0 ? vatAmount(htNum, selectedRate) : 0

  return (
    <div className={styles.panel}>
      {/* VAT Rate Selector */}
      <div className={styles.section}>
        <label className={styles.sectionLabel}>Taux TVA</label>
        <select
          value={selectedRate}
          onChange={handleRateChange}
          className={styles.rateSelector}
          aria-label="Sélectionner le taux TVA"
        >
          {VAT_RATES.map((option) => (
            <option key={option.rate} value={option.rate}>
              {option.label} — {option.description}
            </option>
          ))}
        </select>
      </div>

      {/* HT Input */}
      <div className={styles.section}>
        <label className={styles.sectionLabel}>Prix HT (Hors Taxe)</label>
        <input
          type="number"
          value={ht}
          onChange={(e) => handleHTChange(e.target.value)}
          placeholder="0,00"
          className={styles.input}
          inputMode="decimal"
          step="0.01"
          aria-label="Prix HT"
        />
      </div>

      {/* TTC Input */}
      <div className={styles.section}>
        <label className={styles.sectionLabel}>Prix TTC (Toutes Taxes Comprises)</label>
        <input
          type="number"
          value={ttc}
          onChange={(e) => handleTTCChange(e.target.value)}
          placeholder="0,00"
          className={styles.input}
          inputMode="decimal"
          step="0.01"
          aria-label="Prix TTC"
        />
      </div>

      {/* VAT Amount Result */}
      {(ht || ttc) && (
        <div className={styles.result}>
          <div className={styles.resultRow}>
            <span className={styles.resultLabel}>Montant TVA</span>
            <span className={styles.resultValue}>{vat.toFixed(2)}</span>
          </div>
          {ht && ttc && (
            <>
              <div className={styles.resultRow}>
                <span className={styles.resultLabel}>HT</span>
                <span className={styles.resultValue}>{parseFloat(ht).toFixed(2)}</span>
              </div>
              <div className={styles.resultRow}>
                <span className={styles.resultLabel}>TTC</span>
                <span className={styles.resultValue}>{parseFloat(ttc).toFixed(2)}</span>
              </div>
            </>
          )}
        </div>
      )}

      {/* Rate Description */}
      <div className={styles.description}>{currentRate.description}</div>

      {/* Close Button */}
      <div className={styles.footer}>
        <button onClick={onClose} className={styles.closeBtn}>
          Fermer
        </button>
      </div>
    </div>
  )
}
