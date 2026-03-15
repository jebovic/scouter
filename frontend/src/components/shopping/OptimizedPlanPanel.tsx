import { useOptimizeShoppingList } from '../../hooks/useOptimizer'
import type { OptimizedPlan, ShoppingAction } from '../../api/optimizer'
import styles from './OptimizedPlanPanel.module.css'

interface Props {
  missionSlug: string
}

function actionType(action: string): 'buy' | 'compare' | 'wait' | 'other' {
  const lower = action.toLowerCase()
  if (lower.includes('acheter') || lower.includes('buy')) return 'buy'
  if (lower.includes('compar')) return 'compare'
  if (lower.includes('attend') || lower.includes('wait') || lower.includes('defer')) return 'wait'
  return 'other'
}

function ActionRow({ action, currency }: { action: ShoppingAction; currency: string }) {
  const type = actionType(action.action)
  return (
    <li className={styles.actionItem}>
      <span className={styles.priority}>{action.priority}</span>
      <div className={styles.actionBody}>
        <div className={styles.itemNames}>{action.itemNames.join(', ')}</div>
        <div className={styles.actionMeta}>
          <span className={styles.actionLabel} data-type={type}>
            {action.action}
          </span>
          <span className={styles.merchant}>{action.merchant}</span>
        </div>
        {action.reason && <div className={styles.reason}>{action.reason}</div>}
      </div>
      {action.estSavings > 0 && (
        <span className={styles.savingsBadge}>
          -{action.estSavings.toFixed(0)} {currency}
        </span>
      )}
    </li>
  )
}

function PlanView({
  plan,
  currency,
  onRecalc,
  isPending,
}: {
  plan: OptimizedPlan
  currency: string
  onRecalc: () => void
  isPending: boolean
}) {
  return (
    <>
      <div className={styles.header}>
        <h3 className={styles.title}>Plan d'achat optimise</h3>
        <button className={styles.recalcBtn} onClick={onRecalc} disabled={isPending}>
          {isPending ? 'Calcul...' : 'Recalculer'}
        </button>
      </div>
      <ul className={styles.actionList}>
        {plan.actions.map((a) => (
          <ActionRow key={`${a.priority}-${a.itemIds.join(',')}`} action={a} currency={currency} />
        ))}
      </ul>
      {plan.summary && (
        <div className={styles.callout}>
          <p className={styles.calloutText}>{plan.summary}</p>
        </div>
      )}
    </>
  )
}

export function OptimizedPlanPanel({ missionSlug }: Props) {
  const { mutate, data, isPending, error, reset } = useOptimizeShoppingList(missionSlug)

  if (!data && !error) {
    return (
      <div className={styles.panel}>
        <button className={styles.triggerBtn} onClick={() => mutate()} disabled={isPending}>
          {isPending ? (
            <>
              <span className={styles.spinner} aria-hidden="true" />
              Optimisation en cours...
            </>
          ) : (
            'Optimiser mes achats'
          )}
        </button>
      </div>
    )
  }

  if (error) {
    return (
      <div className={styles.panel}>
        <p className={styles.error}>
          {error instanceof Error ? error.message : 'Erreur lors de l\'optimisation.'}
        </p>
        <button className={styles.triggerBtn} onClick={() => { reset(); mutate() }} disabled={isPending}>
          Reessayer
        </button>
      </div>
    )
  }

  return (
    <div className={styles.panel}>
      <PlanView
        plan={data!}
        currency="EUR"
        onRecalc={() => mutate()}
        isPending={isPending}
      />
    </div>
  )
}
