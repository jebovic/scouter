import { useState } from 'react'
import { useCoachTips, useRefreshCoach } from '../../hooks/useCoach'
import { SkeletonGrid } from '../scouter'
import styles from './CoachPanel.module.css'

interface CoachPanelProps {
  missionSlug: string | undefined
}

const categoryIcons: Record<string, string> = {
  research: '🔬',
  budget: '💰',
  timing: '⏰',
  decision: '⚖️',
}

const priorityEmojis: Record<string, string> = {
  high: '🔴',
  medium: '🟡',
  low: '🟢',
}

export function CoachPanel({ missionSlug }: CoachPanelProps) {
  const [expanded, setExpanded] = useState(true)
  const { data: response, isLoading, error } = useCoachTips(missionSlug)
  const { mutate: refresh, isPending: isRefreshing } = useRefreshCoach(missionSlug)

  if (!missionSlug) return null

  return (
    <div className={styles.panel}>
      <div className={styles.header} onClick={() => setExpanded(!expanded)}>
        <div className={styles.titleArea}>
          <button className={styles.expandBtn} aria-label={expanded ? 'Collapse' : 'Expand'}>
            {expanded ? '▼' : '▶'}
          </button>
          <h3 className={styles.title}>AI Coach 🧠</h3>
        </div>
        {!isLoading && (
          <button
            className={styles.refreshBtn}
            onClick={(e) => {
              e.stopPropagation()
              refresh()
            }}
            disabled={isRefreshing}
            title={isRefreshing ? 'Refreshing...' : 'Refresh tips'}
          >
            {isRefreshing ? '⟳' : '🔄'}
          </button>
        )}
      </div>

      {expanded && (
        <div className={styles.content}>
          {isLoading && (
            <div className={styles.skeletons}>
              <SkeletonGrid count={2} />
            </div>
          )}

          {error && (
            <div className={styles.error}>
              <p>Failed to load coaching tips. Try again later.</p>
            </div>
          )}

          {response && response.tips.length === 0 && (
            <div className={styles.empty}>
              <p>No tips available at the moment.</p>
            </div>
          )}

          {response && response.tips.length > 0 && (
            <div className={styles.tipsContainer}>
              {response.tips.map((tip) => (
                <div key={tip.id} className={styles.tipCard}>
                  <div className={styles.tipHeader}>
                    <span className={styles.categoryIcon}>{categoryIcons[tip.category] || '💡'}</span>
                    <span className={styles.tipTitle}>{tip.title}</span>
                    <span className={styles.priorityBadge} title={`Priority: ${tip.priority}`}>
                      {priorityEmojis[tip.priority] || '●'}
                    </span>
                  </div>
                  <p className={styles.tipBody}>{tip.body}</p>
                  {tip.action && (
                    <div className={styles.action}>
                      <button className={styles.actionBtn}>{tip.action}</button>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}

          <div className={styles.footer}>
            <small>Generated {response ? new Date(response.generatedAt).toLocaleTimeString() : 'now'}</small>
          </div>
        </div>
      )}
    </div>
  )
}
