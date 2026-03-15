import { useShoppingOptimizer } from '../../hooks/useShoppingOptimizer'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import type { OptimizedItem, OptimizeResponse } from '../../api/shoppingoptimizer'
import styles from './ShoppingOptimizerPanel.module.css'

interface Props {
  missionId: string
}

const URGENCY_ICON: Record<string, string> = {
  now: '🔴',
  soon: '🟡',
  wait: '🟢',
}

function ItemRow({ item }: { item: OptimizedItem }) {
  const icon = URGENCY_ICON[item.urgencyLevel] ?? '⚪'
  const { fmt } = useFormatCurrency()
  return (
    <li className={styles.itemRow}>
      <span className={styles.urgencyBadge} aria-label={item.urgencyLevel}>
        {icon}
      </span>
      <div className={styles.itemBody}>
        <p className={styles.itemName}>{item.itemName}</p>
        <div className={styles.itemMeta}>
          {item.merchant && <span className={styles.merchant}>{item.merchant}</span>}
          <span className={styles.price}>{fmt(item.price)}</span>
        </div>
        {item.reasons.length > 0 && (
          <div className={styles.reasons}>
            {item.reasons.map((r) => (
              <span key={r} className={styles.reasonChip}>{r}</span>
            ))}
          </div>
        )}
      </div>
      {item.savingsPotential > 0 && (
        <span className={styles.savings}>
          -{fmt(item.savingsPotential)} à économiser
        </span>
      )}
    </li>
  )
}

function ResultView({
  data,
  onRecalc,
  isPending,
}: {
  data: OptimizeResponse
  onRecalc: () => void
  isPending: boolean
}) {
  return (
    <>
      <div className={styles.header}>
        <h3 className={styles.title}>Priorités d'achat</h3>
        <button className={styles.recalcBtn} onClick={onRecalc} disabled={isPending}>
          {isPending ? 'Calcul...' : 'Recalculer'}
        </button>
      </div>
      {data.summary && <p className={styles.summary}>{data.summary}</p>}
      <ul className={styles.itemList}>
        {data.items.map((item) => (
          <ItemRow key={item.itemId} item={item} />
        ))}
      </ul>
    </>
  )
}

export function ShoppingOptimizerPanel({ missionId }: Props) {
  const { mutate, data, isPending, error, reset } = useShoppingOptimizer(missionId)

  if (!data && !error) {
    return (
      <div className={styles.panel}>
        <button
          className={styles.triggerBtn}
          onClick={() => mutate()}
          disabled={isPending}
        >
          {isPending ? (
            <>
              <span className={styles.spinner} aria-hidden="true" />
              Analyse en cours...
            </>
          ) : (
            '🎯 Optimize'
          )}
        </button>
      </div>
    )
  }

  if (error) {
    return (
      <div className={styles.panel}>
        <p className={styles.error}>
          {error instanceof Error ? error.message : "Erreur lors de l'optimisation."}
        </p>
        <button
          className={styles.triggerBtn}
          onClick={() => { reset(); mutate() }}
          disabled={isPending}
        >
          Réessayer
        </button>
      </div>
    )
  }

  return (
    <div className={styles.panel}>
      <ResultView data={data!} onRecalc={() => mutate()} isPending={isPending} />
    </div>
  )
}
