import { formatCurrencyLocale } from '../../utils/format'
import styles from './PriceAlertBadge.module.css'

interface PriceAlertBadgeProps {
  currentPrice: number
  targetPrice: number | undefined | null
  currency?: string
}

export function PriceAlertBadge({
  currentPrice,
  targetPrice,
  currency = 'USD',
}: PriceAlertBadgeProps) {
  // Return null if no target price is set
  if (targetPrice === undefined || targetPrice === null) {
    return null
  }

  const formattedTarget = formatCurrencyLocale(targetPrice, 'fr-FR', currency)
  const percentageAboveTarget = targetPrice === 0
    ? (currentPrice === 0 ? 0 : Infinity)
    : ((currentPrice - targetPrice) / targetPrice) * 100

  // Target achieved: currentPrice <= targetPrice
  if (currentPrice <= targetPrice) {
    return (
      <span
        className={`${styles.badge} ${styles.achieved}`}
        aria-label={`Target price achieved: ${formattedTarget}`}
      >
        🎯 Objectif atteint!
      </span>
    )
  }

  // Close to target: currentPrice is 1-10% above targetPrice
  if (percentageAboveTarget <= 10) {
    return (
      <span
        className={`${styles.badge} ${styles.close}`}
        aria-label={`Close to target: ${formattedTarget}`}
      >
        📉 Proche de l'objectif
      </span>
    )
  }

  // Target set but not close: currentPrice is more than 10% above targetPrice
  return (
    <span
      className={`${styles.badge} ${styles.set}`}
      aria-label={`Target price set: ${formattedTarget}`}
    >
      🎯 Objectif: {formattedTarget}
    </span>
  )
}
