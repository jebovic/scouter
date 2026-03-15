import { useState } from 'react'
import { useCarbon } from '../../hooks/useCarbon'
import styles from './CarbonBadge.module.css'

interface CarbonBadgeProps {
  itemName: string
  category?: string
}

const LABEL_CLASS: Record<string, string> = {
  'Très faible': 'veryLow',
  'Faible': 'low',
  'Moyen': 'medium',
  'Élevé': 'high',
  'Très élevé': 'veryHigh',
}

export function CarbonBadge({ itemName, category = '' }: CarbonBadgeProps) {
  const [expanded, setExpanded] = useState(false)
  const { data, isLoading, isError } = useCarbon(itemName, category)

  if (isLoading) return <span className={styles.loading}>🌱 …</span>
  if (isError || !data) return null

  const cls = LABEL_CLASS[data.label] ?? 'medium'

  return (
    <div className={styles.wrap}>
      <button
        className={`${styles.pill} ${styles[cls]}`}
        onClick={() => setExpanded((v) => !v)}
        title={`${data.kgCO2} kg CO₂ — ${data.confidence} confidence`}
        aria-expanded={expanded}
      >
        🌱 {data.kgCO2.toFixed(1)} kg CO₂
      </button>
      {expanded && data.tips.length > 0 && (
        <ul className={styles.tips}>
          {data.tips.map((tip, i) => (
            <li key={i}>{tip}</li>
          ))}
        </ul>
      )}
    </div>
  )
}
