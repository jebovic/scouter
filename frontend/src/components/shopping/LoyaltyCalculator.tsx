import { useState, useEffect } from 'react'
import { usePrograms, useCalculatePoints } from '../../hooks/useLoyaltyPoints'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import styles from './LoyaltyCalculator.module.css'

interface LoyaltyCalculatorProps {
  basePrice: number
  merchant?: string
}

/** Inline widget that shows loyalty points / cashback earned for a purchase. */
export function LoyaltyCalculator({ basePrice, merchant }: LoyaltyCalculatorProps) {
  const { data: programs = [] } = usePrograms()
  const { mutate: calculate, data: result, isPending } = useCalculatePoints()
  const { fmt } = useFormatCurrency()

  // Derive the initial program from the merchant name if a match exists.
  const initialProgram = (): string => {
    if (!merchant) return ''
    const merchantLower = merchant.toLowerCase()
    const match = programs.find((p) =>
      merchantLower.includes(p.retailer.toLowerCase()),
    )
    return match ? match.name : (programs[0]?.name ?? '')
  }

  const [selectedProgram, setSelectedProgram] = useState<string>('')

  // Once programs load, pick a sensible default.
  useEffect(() => {
    if (programs.length > 0 && !selectedProgram) {
      setSelectedProgram(initialProgram())
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [programs.length])

  // Recalculate whenever the selection or price changes.
  useEffect(() => {
    if (!selectedProgram || basePrice <= 0) return
    const prog = programs.find((p) => p.name === selectedProgram)
    if (!prog) return
    calculate({
      retailer: prog.retailer,
      price: basePrice,
      programName: selectedProgram,
    })
  }, [selectedProgram, basePrice, programs, calculate])

  if (basePrice <= 0 || programs.length === 0) return null

  return (
    <div className={styles.wrapper} aria-label="Loyalty points calculator">
      <select
        className={styles.select}
        value={selectedProgram}
        onChange={(e) => setSelectedProgram(e.target.value)}
        aria-label="Programme de fidélité"
      >
        {programs.map((p) => (
          <option key={p.name} value={p.name}>
            {p.retailer}
          </option>
        ))}
      </select>

      {isPending && <span className={styles.loading}>…</span>}

      {result && !isPending && (
        <span className={styles.result} aria-live="polite">
          {result.pointsEarned > 0
            ? `+${result.pointsEarned.toFixed(0)} pts`
            : null}
          {result.pointsEarned > 0 && result.estimatedCashback > 0 ? ' · ' : null}
          {result.estimatedCashback > 0
            ? `≈ ${fmt(result.estimatedCashback)}`
            : null}
        </span>
      )}
    </div>
  )
}
