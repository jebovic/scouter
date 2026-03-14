import { Link } from 'react-router-dom'
import { useSimilarOptions } from '../../hooks/useSearch'
import styles from './SimilarOptions.module.css'

interface SimilarOptionsProps {
  optionId: string
  limit?: number
}

export function SimilarOptions({ optionId, limit = 4 }: SimilarOptionsProps) {
  const { data: results = [], isLoading } = useSimilarOptions(optionId, limit)

  if (isLoading || results.length === 0) return null

  return (
    <div className={styles.section}>
      <div className={styles.title}>Similar options</div>
      <div className={styles.list}>
        {results.map((r) => (
          <Link
            key={r.id}
            to={`/missions/${r.missionSlug}/options`}
            className={styles.item}
          >
            <div className={styles.itemBody}>
              <div className={styles.itemName}>{r.name}</div>
              <div className={styles.itemMeta}>
                {r.missionName}
                {r.category ? ` · ${r.category}` : ''}
              </div>
            </div>
            <span className={styles.score}>{Math.round(r.similarity * 100)}%</span>
          </Link>
        ))}
      </div>
    </div>
  )
}
