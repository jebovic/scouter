import { DiffBadge } from './DiffBadge'
import type { AgentRunDTO } from '../../api'
import styles from './AgentRunHistory.module.css'

interface AgentRunHistoryProps {
  runs: AgentRunDTO[]
  isLoading?: boolean
}

function formatRelative(iso: string): string {
  const delta = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(delta / 60_000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}h ago`
  return `${Math.floor(hours / 24)}d ago`
}

export function AgentRunHistory({ runs, isLoading }: AgentRunHistoryProps) {
  if (isLoading) {
    return (
      <div className={styles.container}>
        <div className={styles.header}>AGENT HISTORY</div>
        <div className={styles.loading}>Loading...</div>
      </div>
    )
  }

  if (runs.length === 0) {
    return (
      <div className={styles.container}>
        <div className={styles.header}>AGENT HISTORY</div>
        <div className={styles.empty}>No runs yet</div>
      </div>
    )
  }

  return (
    <div className={styles.container}>
      <div className={styles.header}>AGENT HISTORY</div>
      <div className={styles.list}>
        {runs.map((run) => (
          <div key={run.id} className={styles.row}>
            <div className={styles.rowLeft}>
              <span className={`${styles.agentType} ${styles[run.agentType]}`}>
                {run.agentType === 'research' ? '🔍' : '💰'} {run.agentType}
              </span>
              {run.feedback && (
                <span className={styles.feedback} title={run.feedback}>
                  "{run.feedback.length > 40 ? run.feedback.slice(0, 40) + '…' : run.feedback}"
                </span>
              )}
            </div>
            <div className={styles.rowRight}>
              <DiffBadge diff={run.diff} />
              <span className={styles.timestamp}>{formatRelative(run.createdAt)}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
