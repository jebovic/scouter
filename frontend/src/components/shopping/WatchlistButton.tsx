import { useState } from 'react'
import { useWatchlist } from '../../hooks/useWatchlist'
import styles from './WatchlistButton.module.css'

interface WatchlistButtonProps {
  itemId: string
  itemName: string
  currentPrice: number
}

export function WatchlistButton({ itemId, itemName, currentPrice }: WatchlistButtonProps) {
  const { data: entries, add, remove } = useWatchlist()
  const [threshold, setThreshold] = useState(10)
  const [showForm, setShowForm] = useState(false)

  const existing = entries?.find((e) => e.itemId === itemId)

  if (existing) {
    return (
      <button
        className={`${styles.btn} ${existing.triggeredAt ? styles.triggered : styles.watching}`}
        onClick={() => remove.mutate(existing.id)}
        title={existing.triggeredAt ? `Prix baissé! Cliquer pour supprimer` : `Surveiller (-${existing.thresholdPct}%) — cliquer pour supprimer`}
      >
        {existing.triggeredAt ? '🔔 Prix baissé!' : `👁 -${existing.thresholdPct}%`}
      </button>
    )
  }

  if (showForm) {
    return (
      <div className={styles.form}>
        <label className={styles.label}>Alerte si baisse de</label>
        <input
          type="number"
          min={1}
          max={90}
          value={threshold}
          onChange={(e) => setThreshold(Number(e.target.value))}
          className={styles.input}
        />
        <span className={styles.pct}>%</span>
        <button
          className={styles.confirmBtn}
          onClick={() => {
            add.mutate({ itemId, itemName, baselinePrice: currentPrice, thresholdPct: threshold })
            setShowForm(false)
          }}
        >
          ✓
        </button>
        <button className={styles.cancelBtn} onClick={() => setShowForm(false)}>✕</button>
      </div>
    )
  }

  return (
    <button className={styles.btn} onClick={() => setShowForm(true)} title="Ajouter à la liste de surveillance">
      👁 Surveiller
    </button>
  )
}
