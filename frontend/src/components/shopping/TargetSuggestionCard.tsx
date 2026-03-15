import { useTargetSuggestion, useApplyTargetSuggestion } from '../../hooks/useTargetSuggestion'
import { useUpdateShoppingItem } from '../../hooks/useShopping'
import { formatCurrency } from '../../utils/format'
import type { Strategy } from '../../api/targetsuggestion'
import styles from './TargetSuggestionCard.module.css'

interface TargetSuggestionCardProps {
  missionId: string
  itemId: string
  currency?: string
}

const TIER_CLASS: Record<string, string> = {
  conservative: styles.tierConservative,
  moderate: styles.tierModerate,
  aggressive: styles.tierAggressive,
}

function StrategyButton({
  strategy,
  isApplying,
  currency,
  onApply,
}: {
  strategy: Strategy
  isApplying: boolean
  currency: string
  onApply: (type: Strategy['type']) => void
}) {
  const fmt = (n: number) => formatCurrency(n, currency)
  const tierCls = TIER_CLASS[strategy.type] ?? ''

  return (
    <div className={`${styles.tier} ${tierCls}`}>
      <div className={styles.tierLeft}>
        <span className={styles.tierLabel}>{strategy.label}</span>
        <span className={styles.tierRationale}>{strategy.rationale}</span>
      </div>
      <div className={styles.tierRight}>
        <span className={styles.tierPrice}>{fmt(strategy.price)}</span>
        <button
          className={`${styles.applyBtn} ${tierCls ? styles.applyBtnColored : ''}`}
          disabled={isApplying}
          onClick={() => onApply(strategy.type)}
          aria-label={`Appliquer ${strategy.label}: ${fmt(strategy.price)}`}
        >
          Appliquer
        </button>
      </div>
    </div>
  )
}

export function TargetSuggestionCard({
  missionId,
  itemId,
  currency = 'EUR',
}: TargetSuggestionCardProps) {
  const { suggestion, isLoading } = useTargetSuggestion(missionId, itemId)
  const { apply, isApplying } = useApplyTargetSuggestion(missionId, itemId)
  const { updateItem } = useUpdateShoppingItem(missionId)

  function handleApply(type: Strategy['type']) {
    const strategy = suggestion?.strategies.find((s) => s.type === type)
    if (!strategy) return
    // Use the apply endpoint (POST target-suggestion/apply) as primary,
    // and also update via PATCH shopping item for immediate UI refresh.
    apply(type)
    updateItem({ itemId, req: { targetPrice: strategy.price } })
  }

  if (isLoading) {
    return (
      <div className={styles.card}>
        <div className={styles.title}>🎯 Suggestions de Prix Cible</div>
        <div className={styles.noData}>Chargement…</div>
      </div>
    )
  }

  if (!suggestion) {
    return (
      <div className={styles.card}>
        <div className={styles.title}>🎯 Suggestions de Prix Cible</div>
        <div className={styles.noData}>Pas assez de données (min. 3 relevés de prix requis).</div>
      </div>
    )
  }

  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <div className={styles.title}>🎯 Suggestions de Prix Cible</div>
      </div>

      <div className={styles.tiers}>
        {suggestion.strategies.map((strategy) => (
          <StrategyButton
            key={strategy.type}
            strategy={strategy}
            isApplying={isApplying}
            currency={currency}
            onApply={handleApply}
          />
        ))}
      </div>
    </div>
  )
}
