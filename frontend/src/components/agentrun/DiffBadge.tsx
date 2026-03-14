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

  const label = [
    added > 0 ? `${added} added` : '',
    removed > 0 ? `${removed} removed` : '',
    changed > 0 ? `${changed} changed` : '',
  ].filter(Boolean).join(', ')

  return (
    <span className={styles.wrapper} aria-label={label}>
      {added > 0 && (
        <span className={`${styles.badge} ${styles.added}`} aria-hidden="true">+{added}</span>
      )}
      {removed > 0 && (
        <span className={`${styles.badge} ${styles.removed}`} aria-hidden="true">-{removed}</span>
      )}
      {changed > 0 && (
        <span className={`${styles.badge} ${styles.changed}`} aria-hidden="true">~{changed}</span>
      )}
    </span>
  )
}
