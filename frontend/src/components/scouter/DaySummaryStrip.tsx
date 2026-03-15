import { Link } from 'react-router-dom'
import { useMissions } from '../../hooks'
import styles from './DaySummaryStrip.module.css'

export function DaySummaryStrip() {
  const { missions } = useMissions()
  const activeMissions = missions.filter(m => m.phase === 'researching' || m.phase === 'comparing').length
  const inProgress = missions.filter(m => m.phase !== 'done' && !m.archivedAt).length

  return (
    <div className={styles.strip}>
      <Link to="/envelopes" className={styles.cell}>
        <span className={styles.value}>{activeMissions}</span>
        <span className={styles.label}>missions actives</span>
      </Link>
      <div className={styles.divider} />
      <Link to="/wishlist" className={styles.cell}>
        <span className={styles.value}>{inProgress}</span>
        <span className={styles.label}>en cours</span>
      </Link>
      <div className={styles.divider} />
      <Link to="/deal-calendar" className={styles.cell}>
        <span className={styles.value}>📅</span>
        <span className={styles.label}>Calendrier promos</span>
      </Link>
    </div>
  )
}
