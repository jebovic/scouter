import { useBurnRate } from '../../hooks/useBurnRate'
import styles from './BurnRateCard.module.css'

interface Props {
  missionId: string
  currency?: string
}

const STATUS_LABELS: Record<string, string> = {
  on_track: 'En bonne voie',
  at_risk: 'À risque',
  over_budget: 'Dépassement',
}

function formatCurrency(amount: number, currency = 'EUR'): string {
  return new Intl.NumberFormat('fr-FR', {
    style: 'currency',
    currency,
    maximumFractionDigits: 0,
  }).format(amount)
}

function statusBadgeClass(status: string): string {
  switch (status) {
    case 'on_track':
      return styles.onTrack
    case 'at_risk':
      return styles.atRisk
    case 'over_budget':
      return styles.overBudget
    default:
      return styles.onTrack
  }
}

export function BurnRateCard({ missionId, currency = 'EUR' }: Props) {
  const { data, isLoading, isError } = useBurnRate(missionId)

  if (isLoading) {
    return (
      <section className={styles.card} aria-label="Rythme de dépenses">
        <div className={styles.header}>
          <span className={styles.title}>📈 Rythme de dépenses</span>
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
      <section className={styles.card} aria-label="Rythme de dépenses">
        <div className={styles.header}>
          <span className={styles.title}>📈 Rythme de dépenses</span>
        </div>
        <p className={styles.errorState}>Impossible de charger le rythme de dépenses.</p>
      </section>
    )
  }

  const maxAmount = data.burnPoints.length > 0
    ? data.burnPoints[data.burnPoints.length - 1].amount
    : 1

  return (
    <section className={styles.card} aria-label="Rythme de dépenses">
      <div className={styles.header}>
        <span className={styles.title}>📈 Rythme de dépenses</span>
        <span className={`${styles.statusBadge} ${statusBadgeClass(data.status)}`}>
          {STATUS_LABELS[data.status] ?? data.status}
        </span>
      </div>

      <div className={styles.metrics}>
        <div className={styles.metric}>
          <span className={styles.metricLabel}>Dépensé / jour</span>
          <span className={styles.metricValue}>{formatCurrency(data.dailyRate, currency)}</span>
        </div>
        <div className={styles.metric}>
          <span className={styles.metricLabel}>Projection 30j</span>
          <span className={styles.metricValue}>{formatCurrency(data.projectedTotal, currency)}</span>
        </div>
        <div className={styles.metric}>
          <span className={styles.metricLabel}>Jours écoulés</span>
          <span className={styles.metricValue}>{data.daysElapsed}j</span>
        </div>
      </div>

      {data.burnPoints.length === 0 ? (
        <p className={styles.empty}>Aucun achat enregistré pour l'instant.</p>
      ) : (
        <div className={styles.burnBars}>
          <span className={styles.burnBarsLabel}>Cumul des dépenses</span>
          {data.burnPoints.map((pt) => {
            const pct = maxAmount > 0 ? Math.min((pt.amount / maxAmount) * 100, 100) : 0
            const isOver = pt.amount > data.totalBudget
            return (
              <div key={pt.date} className={styles.barItem}>
                <span className={styles.barDate}>{pt.date}</span>
                <div className={styles.barTrack}>
                  <div
                    className={styles.barFill}
                    style={{ '--pct': pct } as React.CSSProperties}
                    data-over={isOver ? 'true' : 'false'}
                  />
                </div>
                <span className={styles.barAmount}>{formatCurrency(pt.amount, currency)}</span>
              </div>
            )
          })}
        </div>
      )}
    </section>
  )
}
