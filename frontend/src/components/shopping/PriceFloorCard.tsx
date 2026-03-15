import { usePriceFloor } from '../../hooks/usePriceFloor'
import { formatCurrency } from '../../utils/format'
import styles from './PriceFloorCard.module.css'

interface PriceFloorCardProps {
  missionId: string
  itemId: string
  currency?: string
}

export function PriceFloorCard({ missionId, itemId, currency = 'EUR' }: PriceFloorCardProps) {
  const { floor, isLoading } = usePriceFloor(missionId, itemId)
  const fmt = (n: number) => formatCurrency(n, currency)

  if (isLoading) {
    return (
      <div className={styles.card}>
        <div className={styles.title}>Plancher de prix estimé</div>
        <div className={styles.noData}>Chargement…</div>
      </div>
    )
  }

  if (!floor) {
    return (
      <div className={styles.card}>
        <div className={styles.title}>Plancher de prix estimé</div>
        <div className={styles.noData}>Données non disponibles.</div>
      </div>
    )
  }

  const confLabel =
    floor.floorConfidence === 'high'
      ? 'haute'
      : floor.floorConfidence === 'medium'
        ? 'moyenne'
        : 'faible'

  // Compute bar: position current price between floor and a "ceiling" for display.
  // Use historicalMin as floor anchor and currentPrice * 1.2 as a visual ceiling.
  const displayMin = Math.min(floor.predictedFloor, floor.historicalMin)
  const displayMax = Math.max(floor.currentPrice * 1.2, floor.predictedFloor * 1.3)
  const range = displayMax - displayMin

  const floorPct = range > 0 ? ((floor.predictedFloor - displayMin) / range) * 100 : 0
  const currentPct = range > 0 ? ((floor.currentPrice - displayMin) / range) * 100 : 50

  const verdictClass = floor.shouldWait ? styles.wait : styles.buy
  const verdictText = floor.shouldWait ? 'Attendez encore' : 'Proche du plancher !'

  const explanationText = floor.shouldWait ? floor.waitReason : floor.buyNowReason

  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <div className={styles.title}>Plancher de prix estimé</div>
        <span className={`${styles.confidenceBadge} ${styles[floor.floorConfidence]}`}>
          confiance {confLabel}
        </span>
      </div>

      {/* Main verdict */}
      <div className={`${styles.verdict} ${verdictClass}`}>{verdictText}</div>

      {/* Visual price bar */}
      <div className={styles.barSection}>
        <div className={styles.barLabel}>Position du prix actuel</div>
        <div className={styles.barTrack}>
          <div className={styles.barFill} style={{ width: `${Math.min(floorPct + 10, 100)}%` }} />
          <div className={styles.barMarker} style={{ left: `${Math.min(Math.max(currentPct, 2), 98)}%` }} />
        </div>
        <div className={styles.barLabels}>
          <span>Plancher {fmt(floor.predictedFloor)}</span>
          <span>Actuel {fmt(floor.currentPrice)}</span>
        </div>
      </div>

      {/* Stats */}
      <div className={styles.stats}>
        <div className={styles.stat}>
          <span className={styles.statLabel}>Plancher estimé</span>
          <span className={`${styles.statValue} ${styles.floor}`}>{fmt(floor.predictedFloor)}</span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statLabel}>Min historique</span>
          <span className={styles.statValue}>{fmt(floor.historicalMin)}</span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statLabel}>Écart</span>
          <span className={`${styles.statValue} ${styles.distance}`}>
            {fmt(floor.distanceToFloor)} ({floor.distancePct.toFixed(1)} %)
          </span>
        </div>
      </div>

      {/* Explanation */}
      {explanationText && (
        <div className={`${styles.explanation} ${verdictClass}`}>{explanationText}</div>
      )}
    </div>
  )
}
