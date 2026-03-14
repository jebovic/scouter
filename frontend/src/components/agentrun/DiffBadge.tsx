import type { AgentDiff } from '../../api'
import styles from './DiffBadge.module.css'

interface DiffBadgeProps {
  diff: AgentDiff | null
}

export function DiffBadge({ diff }: DiffBadgeProps) {
  if (!diff) return null

  const added = diff.added.length
  const removed = diff.removed.length
  const changed = diff.changed.length

  if (added === 0 && removed === 0 && changed === 0) {
    return <span className={`${styles.badge} ${styles.noChanges}`}>no changes</span>
  }

  return (
    <span className={styles.wrapper}>
      {added > 0 && (
        <span className={`${styles.badge} ${styles.added}`}>+{added}</span>
      )}
      {removed > 0 && (
        <span className={`${styles.badge} ${styles.removed}`}>-{removed}</span>
      )}
      {changed > 0 && (
        <span className={`${styles.badge} ${styles.changed}`}>~{changed}</span>
      )}
    </span>
  )
}
