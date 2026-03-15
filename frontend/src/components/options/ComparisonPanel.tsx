import type { Option } from '../../types'
import styles from './ComparisonPanel.module.css'

interface ComparisonPanelProps {
  options: Option[]
  currency?: string
  onClose: () => void
}

function formatPrice(value: number, currency: string): string {
  return new Intl.NumberFormat('fr-FR', { style: 'currency', currency }).format(value)
}

export function ComparisonPanel({ options, currency = 'EUR', onClose }: ComparisonPanelProps) {
  if (options.length === 0) return null

  // Collect all unique attribute keys, preserving first-seen order
  const attrKeys = Array.from(
    new Set(options.flatMap((o) => o.attributes.map((a) => a.key))),
  )
  const attrLabels = Object.fromEntries(
    options.flatMap((o) => o.attributes).map((a) => [a.key, a.label]),
  )

  // Find lowest best price index for highlighting
  const prices = options.map((o) => o.priceRange?.best ?? null)
  const validPrices = prices.filter((p): p is number => p !== null)
  const lowestPrice = validPrices.length > 0 ? Math.min(...validPrices) : null

  // Find highest score among numeric score-type attributes for each key
  function getBestScoreIndex(key: string): number | null {
    const vals = options.map((o) => {
      const attr = o.attributes.find((a) => a.key === key)
      if (!attr) return null
      const n = Number(attr.value)
      return Number.isFinite(n) ? n : null
    })
    const valid = vals.filter((v): v is number => v !== null)
    if (valid.length === 0) return null
    return Math.max(...valid)
  }

  return (
    <div className={styles.panel} role="region" aria-label="Comparison panel">
      <div className={styles.panelHeader}>
        <span className={styles.panelTitle}>Comparaison</span>
        <button
          className={styles.closeBtn}
          onClick={onClose}
          aria-label="Fermer"
        >
          ×
        </button>
      </div>

      <table className={styles.table}>
        <thead>
          <tr>
            <th className={styles.headerLabel}>Attribut</th>
            {options.map((o) => (
              <th key={o.id} className={styles.headerOption}>
                {o.name}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {/* Price row */}
          <tr className={styles.priceRow}>
            <td className={`${styles.rowLabel}`}>Prix</td>
            {options.map((o) => {
              const price = o.priceRange?.best ?? null
              const isBest = price !== null && price === lowestPrice
              return (
                <td
                  key={o.id}
                  className={`${styles.cell} ${isBest ? styles.cellBest : ''}`}
                  data-best={isBest ? 'true' : undefined}
                >
                  {price !== null ? formatPrice(price, currency) : <span className={styles.cellMissing}>—</span>}
                </td>
              )
            })}
          </tr>

          {/* Attribute rows */}
          {attrKeys.map((key) => {
            const bestScore = getBestScoreIndex(key)
            return (
              <tr key={key}>
                <td className={styles.rowLabel}>{attrLabels[key] ?? key}</td>
                {options.map((o) => {
                  const attr = o.attributes.find((a) => a.key === key)
                  if (!attr) {
                    return (
                      <td key={o.id} className={`${styles.cell} ${styles.cellMissing}`}>
                        —
                      </td>
                    )
                  }
                  const numVal = Number(attr.value)
                  const isBest =
                    bestScore !== null && Number.isFinite(numVal) && numVal === bestScore
                  const displayValue =
                    attr.type === 'boolean'
                      ? attr.value
                        ? '✓'
                        : '✗'
                      : String(attr.value)

                  return (
                    <td
                      key={o.id}
                      className={`${styles.cell} ${isBest ? styles.cellBest : ''}`}
                      data-best={isBest ? 'true' : undefined}
                    >
                      {displayValue}
                    </td>
                  )
                })}
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}
