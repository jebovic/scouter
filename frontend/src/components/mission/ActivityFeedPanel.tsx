import { Link } from 'react-router-dom'
import { useActivityFeed } from '../../hooks/useActivityFeed'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import type { Activity } from '../../api/activityfeed'
import styles from './ActivityFeedPanel.module.css'

const TYPE_ICON: Record<string, string> = {
  price_drop: '📉',
  price_rise: '📈',
  item_added: '➕',
  mission_created: '🎯',
}

function timeAgo(isoDate: string): string {
  const diff = Date.now() - new Date(isoDate).getTime()
  const mins = Math.floor(diff / 60_000)
  if (mins < 1) return "à l'instant"
  if (mins < 60) return `il y a ${mins}min`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `il y a ${hrs}h`
  const days = Math.floor(hrs / 24)
  return `il y a ${days}j`
}

function StatCard({ label, value }: { label: string; value: string | number }) {
  return (
    <div className={styles.statCard}>
      <span className={styles.statValue}>{value}</span>
      <span className={styles.statLabel}>{label}</span>
    </div>
  )
}

function ActivityRow({ activity }: { activity: Activity }) {
  const icon = TYPE_ICON[activity.type] ?? '🔔'
  const isDrop = activity.type === 'price_drop'
  const isRise = activity.type === 'price_rise'

  return (
    <Link
      to={`/missions/${activity.missionSlug}`}
      className={styles.activityRow}
    >
      <span className={styles.activityIcon}>{icon}</span>
      <div className={styles.activityBody}>
        <span
          className={styles.activityDesc}
          data-type={activity.type}
          style={
            isDrop
              ? ({ color: 'var(--green)' } as React.CSSProperties)
              : isRise
                ? ({ color: 'var(--coral)' } as React.CSSProperties)
                : undefined
          }
        >
          {activity.description}
        </span>
        <span className={styles.activityMeta}>
          {activity.missionName} · {timeAgo(activity.occurredAt)}
        </span>
      </div>
    </Link>
  )
}

export function ActivityFeedPanel() {
  const { feed, isLoading } = useActivityFeed()
  const { fmt } = useFormatCurrency()

  return (
    <div className={styles.panel}>
      <div className={styles.statsBar}>
        {isLoading ? (
          <>
            <div className={`${styles.statCard} ${styles.skeleton}`} />
            <div className={`${styles.statCard} ${styles.skeleton}`} />
            <div className={`${styles.statCard} ${styles.skeleton}`} />
            <div className={`${styles.statCard} ${styles.skeleton}`} />
          </>
        ) : (
          <>
            <StatCard label="Missions" value={feed?.totalMissions ?? 0} />
            <StatCard label="Articles" value={feed?.totalItems ?? 0} />
            <StatCard label="Alertes actives" value={feed?.activeAlerts ?? 0} />
            <StatCard
              label="Budget total"
              value={feed ? fmt(feed.totalBudget) : '—'}
            />
          </>
        )}
      </div>

      <div className={styles.feedSection}>
        <span className={styles.feedLabel}>ACTIVITÉ RÉCENTE</span>

        {isLoading ? (
          <div className={styles.feedList}>
            {Array.from({ length: 4 }).map((_, i) => (
              <div key={i} className={`${styles.activityRow} ${styles.skeleton}`} style={{ height: 48 }} />
            ))}
          </div>
        ) : !feed || feed.activities.length === 0 ? (
          <div className={styles.empty}>
            <span className={styles.emptyIcon}>📭</span>
            <span className={styles.emptyText}>Aucune activité récente</span>
          </div>
        ) : (
          <div className={styles.feedList}>
            {feed.activities.map((a, i) => (
              <ActivityRow key={`${a.missionId}-${a.occurredAt}-${i}`} activity={a} />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
