import { usePriceForecast } from '../../hooks/usePriceForecast'
import type { ForecastPoint } from '../../api/priceforecast'
import styles from './PriceForecastCard.module.css'

// ── SVG area chart ────────────────────────────────────────────────────────────

function ForecastChart({ points, currency }: { points: ForecastPoint[]; currency: string }) {
  if (points.length === 0) return null

  const W = 400
  const H = 120
  const PAD = { top: 8, right: 8, bottom: 20, left: 36 }
  const chartW = W - PAD.left - PAD.right
  const chartH = H - PAD.top - PAD.bottom

  const allValues = points.flatMap((p) => [p.low, p.high])
  const minVal = Math.min(...allValues) * 0.98
  const maxVal = Math.max(...allValues) * 1.02
  const range = maxVal - minVal || 1

  const xStep = chartW / Math.max(points.length - 1, 1)
  const scaleY = (v: number) => chartH - ((v - minVal) / range) * chartH

  const px = (i: number) => PAD.left + i * xStep
  const py = (v: number) => PAD.top + scaleY(v)

  // confidence band path (low → high → reverse)
  const bandTop = points.map((p, i) => `${i === 0 ? 'M' : 'L'}${px(i).toFixed(1)},${py(p.high).toFixed(1)}`).join(' ')
  const bandBot = points
    .slice()
    .reverse()
    .map((p, i, arr) => `${i === 0 ? 'L' : 'L'}${px(arr.length - 1 - i).toFixed(1)},${py(p.low).toFixed(1)}`)
    .join(' ')
  const bandPath = `${bandTop} ${bandBot} Z`

  // mid-price line
  const midLine = points.map((p, i) => `${i === 0 ? 'M' : 'L'}${px(i).toFixed(1)},${py(p.mid).toFixed(1)}`).join(' ')

  // axis labels
  const fmt = new Intl.NumberFormat('fr-FR', {
    style: 'currency',
    currency,
    maximumFractionDigits: 0,
  })

  return (
    <svg
      viewBox={`0 0 ${W} ${H}`}
      className={styles.svgChart}
      aria-hidden="true"
      preserveAspectRatio="none"
    >
      {/* confidence band */}
      <path d={bandPath} fill="var(--cyan)" opacity="0.12" />
      {/* mid line */}
      <path d={midLine} stroke="var(--cyan)" strokeWidth="1.5" fill="none" />

      {/* Y axis labels */}
      <text x={PAD.left - 4} y={PAD.top + 4} className={styles.axisLabel} textAnchor="end">
        {fmt.format(maxVal)}
      </text>
      <text x={PAD.left - 4} y={PAD.top + chartH} className={styles.axisLabel} textAnchor="end">
        {fmt.format(minVal)}
      </text>

      {/* X axis first/last date */}
      {points.length > 0 && (
        <>
          <text x={PAD.left} y={H - 4} className={styles.axisLabel} textAnchor="start">
            J+1
          </text>
          <text x={W - PAD.right} y={H - 4} className={styles.axisLabel} textAnchor="end">
            J+{points.length}
          </text>
        </>
      )}
    </svg>
  )
}

// ── Main component ────────────────────────────────────────────────────────────

interface PriceForecastCardProps {
  missionId: string
  itemId: string
  currency?: string
}

export function PriceForecastCard({ missionId, itemId, currency = 'EUR' }: PriceForecastCardProps) {
  const { data, isLoading, error } = usePriceForecast(missionId, itemId)

  if (isLoading) {
    return (
      <div className={styles.skeleton}>
        <div className={styles.skeletonLine} style={{ width: '40%' }} />
        <div className={styles.skeletonChart} />
        <div className={styles.skeletonLine} style={{ width: '80%' }} />
        <div className={styles.skeletonLine} style={{ width: '60%' }} />
      </div>
    )
  }

  if (error || !data) return null

  const confPct = Math.round(data.confidence * 100)
  const confClass =
    data.confidence >= 0.7
      ? styles.confidenceBadge
      : data.confidence >= 0.4
        ? `${styles.confidenceBadge} ${styles.confidenceMid}`
        : `${styles.confidenceBadge} ${styles.confidenceLow}`

  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <span className={styles.title}>Prévision 30j</span>
        <span className={confClass} title={`Confiance : ${confPct}%`}>
          {confPct}% confiance
        </span>
      </div>

      <div className={styles.chartContainer}>
        <ForecastChart points={data.forecast} currency={data.currency || currency} />
      </div>

      <div className={styles.metaRow}>
        {data.best_buy_window && (
          <span className={styles.bestBuy}>
            <span className={styles.bestBuyLabel}>Acheter</span>
            {data.best_buy_window}
          </span>
        )}
        {data.summary && <p className={styles.summary}>{data.summary}</p>}
      </div>
    </div>
  )
}
