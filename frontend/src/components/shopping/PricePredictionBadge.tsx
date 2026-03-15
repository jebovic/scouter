import { usePricePrediction } from '../../hooks/usePricePrediction'
import styles from './PricePredictionBadge.module.css'

interface PricePredictionBadgeProps {
  itemId: string
}

const TREND_ICON: Record<string, string> = {
  rising: '📈',
  falling: '📉',
  stable: '➡️',
}

const CONFIDENCE_LABEL: Record<string, string> = {
  high: 'confiance élevée',
  medium: 'confiance moyenne',
  low: 'confiance faible',
}

const CONFIDENCE_DOT_CLASS: Record<string, string> = {
  high: styles.dotHigh,
  medium: styles.dotMedium,
  low: styles.dotLow,
}

export function PricePredictionBadge({ itemId }: PricePredictionBadgeProps) {
  const { data } = usePricePrediction(itemId)

  if (!data || data.dataPoints < 1) return null

  const { trend, trendPct, confidenceLevel, dataPoints } = data

  const label = (() => {
    const abs = Math.abs(trendPct).toFixed(1)
    if (trend === 'rising') return `En hausse +${abs}%`
    if (trend === 'falling') return `En baisse -${abs}%`
    return 'Stable'
  })()

  const tooltipText = `Prédiction sur 30j basée sur ${dataPoints} relevé${dataPoints > 1 ? 's' : ''} de prix`
  const confidenceTitle = CONFIDENCE_LABEL[confidenceLevel] ?? confidenceLevel

  return (
    <span
      className={`${styles.wrapper} ${styles[trend] ?? styles.stable}`}
      title={tooltipText}
    >
      <span
        className={`${styles.dot} ${CONFIDENCE_DOT_CLASS[confidenceLevel] ?? styles.dotLow}`}
        title={confidenceTitle}
        aria-label={confidenceTitle}
      />
      <span aria-hidden="true">{TREND_ICON[trend]}</span>
      {label}
    </span>
  )
}
