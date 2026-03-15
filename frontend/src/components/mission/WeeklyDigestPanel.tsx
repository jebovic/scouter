import { Link } from 'react-router-dom'
import { useWeeklyDigest } from '../../hooks/useWeeklyDigest'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import type { DigestAlert } from '../../api/weeklydigest'
import styles from './WeeklyDigestPanel.module.css'

function formatDateRange(start: string, end: string): string {
  const fmt = new Intl.DateTimeFormat('fr-FR', { day: 'numeric', month: 'long' })
  return `${fmt.format(new Date(start))} — ${fmt.format(new Date(end))}`
}

const ALERT_LABELS: Record<DigestAlert['alertType'], string> = {
  big_drop: 'Grandes baisses',
  new_low: 'Nouveaux planchers',
  target_hit: 'Objectifs atteints',
}

const ALERT_COLORS: Record<DigestAlert['alertType'], string> = {
  big_drop: 'var(--green)',
  new_low: 'var(--cyan)',
  target_hit: 'var(--gold)',
}

function ChangeBadge({ pct, alertType }: { pct: number; alertType: DigestAlert['alertType'] }) {
  const color = ALERT_COLORS[alertType]
  const sign = pct <= 0 ? '' : '+'
  return (
    <span className={styles.changeBadge} style={{ color } as React.CSSProperties}>
      {sign}{pct.toFixed(1)}%
    </span>
  )
}

function AlertRow({ alert }: { alert: DigestAlert }) {
  const { fmt } = useFormatCurrency()
  return (
    <Link to={`/missions/${alert.missionSlug}`} className={styles.alertRow}>
      <div className={styles.alertInfo}>
        <span className={styles.alertName}>{alert.itemName}</span>
        <span className={styles.alertMission}>{alert.missionName}</span>
      </div>
      <div className={styles.alertPrices}>
        <span className={styles.oldPrice}>{fmt(alert.oldPrice)}</span>
        <span className={styles.priceArrow}>→</span>
        <span className={styles.newPrice}>{fmt(alert.newPrice)}</span>
        <ChangeBadge pct={alert.changePct} alertType={alert.alertType} />
      </div>
    </Link>
  )
}

function AlertSection({
  title,
  alerts,
  alertType,
}: {
  title: string
  alerts: DigestAlert[]
  alertType: DigestAlert['alertType']
}) {
  if (alerts.length === 0) return null
  const color = ALERT_COLORS[alertType]
  return (
    <div className={styles.section}>
      <div className={styles.sectionHeader}>
        <span className={styles.sectionTitle} style={{ color } as React.CSSProperties}>
          {title}
        </span>
        <span className={styles.sectionCount}>{alerts.length}</span>
      </div>
      <div className={styles.alertList}>
        {alerts.map((a) => (
          <AlertRow key={`${a.itemId}-${a.alertType}`} alert={a} />
        ))}
      </div>
    </div>
  )
}

export function WeeklyDigestPanel() {
  const { digest, isLoading } = useWeeklyDigest()

  if (isLoading) {
    return (
      <div className={styles.panel}>
        <div className={styles.header}>
          <span className={styles.headerTitle}>RESUME DE LA SEMAINE</span>
        </div>
        <div className={styles.skeletonBody}>
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className={styles.skeletonRow} />
          ))}
        </div>
      </div>
    )
  }

  if (!digest) return null

  const hasAlerts = digest.totalAlerts > 0

  return (
    <div className={styles.panel}>
      <div className={styles.header}>
        <div className={styles.headerLeft}>
          <span className={styles.headerTitle}>RESUME DE LA SEMAINE</span>
          <span className={styles.dateRange}>
            {formatDateRange(digest.weekStart, digest.weekEnd)}
          </span>
        </div>
        {digest.totalAlerts > 0 && (
          <span className={styles.totalBadge}>{digest.totalAlerts} alertes</span>
        )}
      </div>

      {!hasAlerts ? (
        <div className={styles.empty}>
          <span className={styles.emptyIcon}>📭</span>
          <span className={styles.emptyText}>{digest.summary}</span>
        </div>
      ) : (
        <>
          <div className={styles.sections}>
            <AlertSection
              title={ALERT_LABELS.big_drop}
              alerts={digest.bigDrops}
              alertType="big_drop"
            />
            <AlertSection
              title={ALERT_LABELS.new_low}
              alerts={digest.newLows}
              alertType="new_low"
            />
            <AlertSection
              title={ALERT_LABELS.target_hit}
              alerts={digest.targetHits}
              alertType="target_hit"
            />
          </div>
          <div className={styles.summaryBar}>{digest.summary}</div>
        </>
      )}
    </div>
  )
}
