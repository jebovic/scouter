import { useElasticity } from '../../hooks/useElasticity'
import { formatCurrency } from '../../utils/format'
import styles from './ElasticityCard.module.css'

interface ElasticityCardProps {
  missionId: string
  itemId: string
  currency?: string
}

/** Maps elasticity value (-3 to +3) to a 0–100% position on the gauge. */
function elasticityToPercent(e: number): number {
  return Math.round(((e + 3) / 6) * 100)
}

function classificationBadgeClass(classification: string): string {
  switch (classification) {
    case 'Très élastique':
      return styles.badgeVeryElastic
    case 'Élastique':
      return styles.badgeElastic
    case 'Fortement inélastique':
      return styles.badgeStronglyInelastic
    default:
      return styles.badgeInelastic
  }
}

export function ElasticityCard({ missionId, itemId, currency = 'EUR' }: ElasticityCardProps) {
  const { result, isLoading } = useElasticity(missionId, itemId)

  if (isLoading) {
    return (
      <div className={styles.card}>
        <div className={styles.title}>Élasticité prix</div>
        <div style={{ fontSize: '0.8rem', color: 'var(--text-dim)', fontStyle: 'italic' }}>
          Analyse en cours…
        </div>
      </div>
    )
  }

  if (!result || result.elasticity === null) {
    return null
  }

  const pct = elasticityToPercent(result.elasticity)

  return (
    <div className={styles.card}>
      <div className={styles.title}>Élasticité prix — sensibilité à la demande</div>

      {/* Horizontal gauge -3 to +3 */}
      <div className={styles.gaugeWrapper}>
        <div className={styles.gaugeTrack}>
          <div
            className={styles.gaugeMarker}
            style={{ left: `${pct}%` } as React.CSSProperties}
            role="img"
            aria-label={`Élasticité: ${result.elasticity}`}
          />
        </div>
        <div className={styles.gaugeLabels}>
          <span>−3 Inélastique</span>
          <span>0</span>
          <span>+3 Élastique</span>
        </div>
      </div>

      {/* Classification badge */}
      <span className={`${styles.badge} ${classificationBadgeClass(result.classification)}`}>
        {result.classification} ({result.elasticity > 0 ? '+' : ''}{result.elasticity})
      </span>

      {/* Insight */}
      <div className={styles.insight}>{result.insight}</div>

      {/* Optimal buy threshold */}
      {result.optimalBuyThreshold !== null && (
        <div className={styles.threshold}>
          <span className={styles.thresholdLabel}>Seuil d'achat optimal :</span>
          <span className={styles.thresholdValue}>
            {formatCurrency(result.optimalBuyThreshold, currency)}
          </span>
        </div>
      )}

      {/* Data points */}
      <div className={styles.meta}>{result.pricePoints} point{result.pricePoints > 1 ? 's' : ''} de prix analysés</div>
    </div>
  )
}
