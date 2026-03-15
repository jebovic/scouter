import { useDropPrediction } from '../../hooks/useDropPrediction'
import styles from './DropPredictionCard.module.css'

interface DropPredictionCardProps {
  missionId: string
  itemId: string
}

function barColorClass(probability: number): string {
  if (probability < 0.3) return styles.green
  if (probability < 0.6) return styles.yellow
  return styles.red
}

export function DropPredictionCard({ missionId, itemId }: DropPredictionCardProps) {
  const { prediction, isLoading } = useDropPrediction(missionId, itemId)

  if (isLoading) {
    return (
      <div className={styles.card}>
        <div className={styles.title}>Prédiction de baisse</div>
        <div style={{ fontSize: '0.8rem', color: 'var(--text-dim)', fontStyle: 'italic' }}>
          Analyse en cours…
        </div>
      </div>
    )
  }

  if (!prediction || prediction.probability <= 0.1) {
    return null
  }

  const pct = Math.round(prediction.probability * 100)
  const colorClass = barColorClass(prediction.probability)
  const confidenceLevel = prediction.confidence === 'élevé' ? 'eleve' : prediction.confidence

  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <div className={styles.title}>Prédiction de baisse de prix</div>
        <span
          className={`${styles.confidenceBadge} ${prediction.confidence === 'faible' ? styles.faible : prediction.confidence === 'moyen' ? styles.moyen : ''}`}
          data-level={confidenceLevel}
        >
          confiance {prediction.confidence}
        </span>
      </div>

      {/* Probability bar */}
      <div className={styles.barSection}>
        <div className={styles.barLabel}>
          <span className={styles.barLabelText}>Probabilité de baisse</span>
          <span className={styles.barPct}>{pct}%</span>
        </div>
        <div className={styles.barTrack}>
          <div
            className={`${styles.barFill} ${colorClass}`}
            style={{ width: `${pct}%` } as React.CSSProperties}
          />
        </div>
      </div>

      {/* Stats */}
      <div className={styles.stats}>
        {prediction.predictedDropPercent > 0 && (
          <div className={styles.stat}>
            <span className={styles.statLabel}>Baisse estimée</span>
            <span className={styles.statValue}>−{prediction.predictedDropPercent}%</span>
          </div>
        )}
        <div className={styles.stat}>
          <span className={styles.statLabel}>Horizon</span>
          <span className={styles.statValue}>{prediction.horizon} jours</span>
        </div>
      </div>

      {/* Signals */}
      {prediction.signals.length > 0 && (
        <div className={styles.signals}>
          <div className={styles.signalsLabel}>Signaux</div>
          <ul className={styles.signalList}>
            {prediction.signals.map((signal, i) => (
              <li key={i} className={styles.signalItem}>
                {signal}
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}
