import { useSpendingVelocity } from '../../hooks/useSpendingVelocity'
import styles from './SpendingVelocityCard.module.css'

interface Props {
  missionId: string
  currency?: string
}

function formatCurrency(amount: number, currency = 'EUR'): string {
  return new Intl.NumberFormat('fr-FR', {
    style: 'currency',
    currency,
    maximumFractionDigits: 0,
  }).format(amount)
}

function paceStatusClass(status: string): string {
  switch (status) {
    case 'fast':
      return styles.fast
    case 'moderate':
      return styles.moderate
    case 'on_track':
      return styles.onTrack
    case 'slow':
      return styles.slow
    case 'unconfigured':
      return styles.unconfigured
    default:
      return styles.onTrack
  }
}

export function SpendingVelocityCard({ missionId, currency = 'EUR' }: Props) {
  const { data, isLoading, isError } = useSpendingVelocity(missionId)

  if (isLoading) {
    return (
      <section className={styles.card} aria-label="Vélocité de dépenses">
        <div className={styles.header}>
          <span className={styles.title}>⚡ Vélocité de dépenses</span>
        </div>
        <div className={`${styles.skeleton} ${styles.skeletonLine}`} />
        <div className={styles.skeletonGrid}>
          {[1, 2, 3].map((i) => (
            <div key={i} className={`${styles.skeleton} ${styles.skeletonCell}`} />
          ))}
        </div>
      </section>
    )
  }

  if (isError || !data) {
    return (
      <section className={styles.card} aria-label="Vélocité de dépenses">
        <div className={styles.header}>
          <span className={styles.title}>⚡ Vélocité de dépenses</span>
        </div>
        <p className={styles.errorState}>Impossible de charger la vélocité de dépenses.</p>
      </section>
    )
  }

  const progressPercent = Math.min(data.budgetUsedPercent, 100)
  const daysLeftDisplay = data.projectedDaysLeft === -1 ? '∞' : `${data.projectedDaysLeft}j`

  return (
    <section className={styles.card} aria-label="Vélocité de dépenses">
      <div className={styles.header}>
        <span className={styles.title}>⚡ Vélocité de dépenses</span>
        <span className={`${styles.paceBadge} ${paceStatusClass(data.paceStatus)}`}>
          {data.paceLabel}
        </span>
      </div>

      {/* Progress bar */}
      <div className={styles.progressSection}>
        <div className={styles.progressLabel}>
          <span className={styles.progressLabelText}>Budget utilisé</span>
          <span className={styles.progressPercent}>{data.budgetUsedPercent.toFixed(1)}%</span>
        </div>
        <div className={styles.progressBar}>
          <div
            className={styles.progressFill}
            style={{
              '--pct': progressPercent,
              '--pace-color': getPaceColor(data.paceStatus),
            } as React.CSSProperties}
          />
        </div>
      </div>

      {/* Key metrics */}
      <div className={styles.metrics}>
        <div className={styles.metric}>
          <span className={styles.metricLabel}>Par jour</span>
          <span className={styles.metricValue}>{formatCurrency(data.dailyVelocity, currency)}</span>
        </div>
        <div className={styles.metric}>
          <span className={styles.metricLabel}>Jours restants</span>
          <span className={styles.metricValue}>{daysLeftDisplay}</span>
        </div>
        <div className={styles.metric}>
          <span className={styles.metricLabel}>Jours actifs</span>
          <span className={styles.metricValue}>{data.daysActive}j</span>
        </div>
      </div>

      {/* Advice */}
      <div className={styles.advice}>
        <p className={styles.adviceText}>{data.advice}</p>
      </div>
    </section>
  )
}

function getPaceColor(status: string): string {
  switch (status) {
    case 'fast':
      return 'var(--coral)'
    case 'moderate':
      return 'var(--orange)'
    case 'on_track':
      return 'var(--green)'
    case 'slow':
      return 'var(--gray)'
    case 'unconfigured':
      return 'var(--text-dim)'
    default:
      return 'var(--accent)'
  }
}
