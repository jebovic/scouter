import { useBudgetPlan } from '../../hooks/useBudgetPlan'
import type { PlannedItem } from '../../api/budgetplanner'
import styles from './BudgetPlannerPanel.module.css'

interface Props {
  missionId: string
  currency?: string
  onClose: () => void
}

const STATUS_ICONS: Record<string, string> = {
  'flash-sale': '⚡',
  buy: '✅',
  watch: '👁',
  defer: '⏸',
  crisis: '🚨',
  preorder: '📦',
}

const BUDGET_STATUS_LABELS: Record<string, { label: string; icon: string }> = {
  under: { label: 'Budget confortable', icon: '✅' },
  tight: { label: 'Budget serré', icon: '⚠️' },
  over: { label: 'Budget dépassé', icon: '🔴' },
}

function fmt(value: number, currency = 'EUR') {
  return new Intl.NumberFormat('fr-FR', { style: 'currency', currency, maximumFractionDigits: 0 }).format(value)
}

function SequenceItem({ item, rank, currency }: { item: PlannedItem; rank: number; currency: string }) {
  const icon = STATUS_ICONS[item.status] ?? '•'
  return (
    <li className={styles.itemRow}>
      <span className={styles.rank}>{rank}</span>
      <div className={styles.itemBody}>
        <p className={styles.itemName}>
          {icon} {item.itemName}
        </p>
        <div className={styles.itemMeta}>
          <span className={`${styles.statusBadge} ${styles[item.status] ?? ''}`}>{item.status}</span>
          <span style={{ fontSize: '0.6875rem', color: 'var(--text-dim)' }}>Priorité {item.priority}</span>
        </div>
        <p className={styles.itemReason}>{item.reason}</p>
      </div>
      <div className={styles.itemPrice}>
        <span className={styles.priceValue}>{fmt(item.price, currency)}</span>
        <span className={styles.cumulative}>Cumulé {fmt(item.cumulativeCost, currency)}</span>
      </div>
    </li>
  )
}

function DeferredItem({ item, currency }: { item: PlannedItem; currency: string }) {
  const icon = STATUS_ICONS[item.status] ?? '•'
  return (
    <div className={styles.deferredRow}>
      <div style={{ flex: 1, minWidth: 0 }}>
        <p className={styles.deferredName}>
          {icon} {item.itemName}
        </p>
        <p className={styles.deferredReason}>{item.reason}</p>
      </div>
      <span className={styles.deferredPrice}>{fmt(item.price, currency)}</span>
    </div>
  )
}

export function BudgetPlannerPanel({ missionId, currency = 'EUR', onClose }: Props) {
  const { data: plan, isLoading, isError } = useBudgetPlan(missionId)

  const budgetStatusInfo = plan ? (BUDGET_STATUS_LABELS[plan.budgetStatus] ?? BUDGET_STATUS_LABELS.tight) : null

  return (
    <div className={styles.panel}>
      <div className={styles.header}>
        <h3 className={styles.title}>Plan d'Achat</h3>
        <button className={styles.closeBtn} onClick={onClose} aria-label="Fermer le planificateur">
          Fermer
        </button>
      </div>

      {isLoading && <p className={styles.loading}>Calcul du plan en cours…</p>}
      {isError && <p className={styles.error}>Impossible de charger le plan. Réessayez.</p>}

      {plan && budgetStatusInfo && (
        <>
          {/* Budget status bar */}
          <div className={`${styles.statusBar} ${styles[plan.budgetStatus]}`}>
            <span className={styles.statusIcon}>{budgetStatusInfo.icon}</span>
            <div className={styles.statusText}>
              <span className={styles.statusLabel}>{budgetStatusInfo.label}</span>
              <span className={styles.statusMeta}>
                Estimé {fmt(plan.totalEstimate, currency)} sur {fmt(plan.totalBudget, currency)} de budget
              </span>
            </div>
            {plan.savings > 0 && (
              <span className={styles.savings}>Économies potentielles {fmt(plan.savings, currency)}</span>
            )}
          </div>

          {/* Sequence */}
          {plan.sequence.length > 0 ? (
            <>
              <p className={styles.sectionTitle}>
                Ordre d'achat recommandé ({plan.sequence.length} article{plan.sequence.length > 1 ? 's' : ''})
              </p>
              <ul className={styles.itemList}>
                {plan.sequence.map((item, i) => (
                  <SequenceItem key={item.itemId} item={item} rank={i + 1} currency={currency} />
                ))}
              </ul>
            </>
          ) : (
            <p className={styles.empty}>Aucun article dans le budget disponible.</p>
          )}

          {/* Deferred */}
          {plan.deferred.length > 0 && (
            <div className={styles.deferredSection}>
              <p className={styles.sectionTitle}>
                Articles reportés ({plan.deferred.length}) — hors budget
              </p>
              {plan.deferred.map((item) => (
                <DeferredItem key={item.itemId} item={item} currency={currency} />
              ))}
            </div>
          )}
        </>
      )}
    </div>
  )
}
