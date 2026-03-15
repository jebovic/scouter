import { useBudgetRebalancer } from '../../hooks/useRebalancer'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import type { EnvelopeAdjustment } from '../../api/rebalancer'
import styles from './BudgetRebalancer.module.css'

function DeltaBadge({ delta }: { delta: number }) {
  if (Math.abs(delta) < 0.5) {
    return <span className={`${styles.deltaBadge} ${styles.deltaOptimal}`}>Optimal</span>
  }
  if (delta > 0) {
    return (
      <span className={`${styles.deltaBadge} ${styles.deltaIncrease}`}>
        +{delta.toFixed(1)} %
      </span>
    )
  }
  return (
    <span className={`${styles.deltaBadge} ${styles.deltaDecrease}`}>
      {delta.toFixed(1)} %
    </span>
  )
}

function AdjustmentCard({ adj, fmt }: { adj: EnvelopeAdjustment; fmt: (n: number) => string }) {
  const isOptimal = Math.abs(adj.deltaPercent) < 0.5

  return (
    <div className={styles.card}>
      <div className={styles.cardHeader}>
        <span className={styles.envelopeName}>{adj.envelopeName}</span>
        <DeltaBadge delta={adj.deltaPercent} />
      </div>

      {isOptimal ? (
        <div className={styles.amounts}>
          <span className={styles.amountCurrent}>{fmt(adj.currentMonthly)}/mois</span>
          <span style={{ color: 'var(--green)', fontSize: '0.8rem' }}>Allocation optimale</span>
        </div>
      ) : (
        <div className={styles.amounts}>
          <span className={styles.amountCurrent}>{fmt(adj.currentMonthly)}</span>
          <span className={styles.arrow}>→</span>
          <span className={styles.amountSuggested}>{fmt(adj.suggestedMonthly)}/mois</span>
        </div>
      )}

      <div className={styles.reason}>{adj.reason}</div>
    </div>
  )
}

export function BudgetRebalancer() {
  const { plan, isLoading, isError, refetch } = useBudgetRebalancer()
  const { fmt } = useFormatCurrency()

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <span className={styles.title}>Rééquilibrage du budget</span>
        <button
          className={styles.refetchBtn}
          onClick={() => refetch()}
          disabled={isLoading}
          aria-label="Rééquilibrer le budget"
        >
          Rééquilibrer le budget
        </button>
      </div>

      {isLoading && (
        <div className={styles.spinner}>
          <span className={styles.spinnerDot} />
          Analyse en cours…
        </div>
      )}

      {isError && !isLoading && (
        <div className={styles.empty}>
          Impossible de charger les suggestions. Veuillez réessayer.
        </div>
      )}

      {!isLoading && !isError && plan && plan.adjustments.length === 0 && (
        <div className={styles.empty}>
          Créez des enveloppes budgétaires pour obtenir des suggestions de rééquilibrage.
        </div>
      )}

      {!isLoading && !isError && plan && plan.adjustments.length > 0 && (
        <>
          {plan.summary && <p className={styles.summary}>{plan.summary}</p>}

          <div className={styles.adjustmentList}>
            {plan.adjustments.map((adj) => (
              <AdjustmentCard key={adj.envelopeId} adj={adj} fmt={fmt} />
            ))}
          </div>

          <div className={styles.footer}>
            <div className={styles.footerTotal}>
              <span className={styles.footerLabel}>Total actuel :</span>
              <span className={styles.footerValue}>{fmt(plan.totalCurrent)}/mois</span>
            </div>
            <div className={styles.footerTotal}>
              <span className={styles.footerLabel}>Total suggéré :</span>
              <span className={styles.footerValue}>{fmt(plan.totalSuggested)}/mois</span>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
