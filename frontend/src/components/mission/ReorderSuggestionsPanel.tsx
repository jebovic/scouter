import { useReorderSuggestions } from '../../hooks/useReorderSuggestions'
import { formatCurrency } from '../../utils/format'
import styles from './ReorderSuggestionsPanel.module.css'

interface Props {
  missionId: string
  currency?: string
}

const PRIORITY_CLASS: Record<string, string> = {
  'immédiat': 'critical',
  'cette semaine': 'warning',
  'ce mois': 'info',
  'flexible': 'neutral',
}

export function ReorderSuggestionsPanel({ missionId, currency = 'EUR' }: Props) {
  const { data, isLoading } = useReorderSuggestions(missionId)
  const fmt = (n: number) => formatCurrency(n, currency)

  if (isLoading) return null
  if (!data || data.suggestions.length === 0) return null

  return (
    <div className={styles.panel}>
      <div className={styles.title}>🛍️ Ordre d'achat recommandé</div>
      <div className={styles.total}>Total à acheter : <strong>{fmt(data.totalToBuy)}</strong></div>
      <ol className={styles.list}>
        {data.suggestions.map((s) => (
          <li key={s.itemId} className={styles.item}>
            <span className={`${styles.priority} ${styles[PRIORITY_CLASS[s.priority] ?? 'neutral']}`}>
              {s.priority}
            </span>
            <span className={styles.name}>{s.name}</span>
            <span className={styles.price}>{fmt(s.price)}</span>
            <span className={styles.reason}>{s.reasoning}</span>
          </li>
        ))}
      </ol>
      {data.merchantBatches.map((b) => (
        <div key={b.merchant} className={styles.batch}>
          💡 {b.batchNote}
        </div>
      ))}
    </div>
  )
}
