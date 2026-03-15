import { Link } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { useMissions } from '../../hooks'
import styles from './LastVisitCard.module.css'

export const LAST_MISSION_KEY = 'lastVisitedMissionSlug'

export function setLastVisitedMission(slug: string) {
  try { localStorage.setItem(LAST_MISSION_KEY, slug) } catch {}
}

export function LastVisitCard() {
  const { missions } = useMissions()
  const [lastSlug, setLastSlug] = useState<string | null>(null)

  useEffect(() => {
    try {
      setLastSlug(localStorage.getItem(LAST_MISSION_KEY))
    } catch {}
  }, [])

  if (!lastSlug) return null
  const mission = missions.find(m => m.slug === lastSlug)
  if (!mission) return null

  return (
    <Link to={`/missions/${mission.slug}`} className={styles.card}>
      <span className={styles.icon} aria-hidden="true">{mission.icon ?? '📦'}</span>
      <div className={styles.content}>
        <span className={styles.label}>Reprendre où vous en étiez</span>
        <span className={styles.name}>{mission.name}</span>
      </div>
      <span className={styles.arrow} aria-hidden="true">→</span>
    </Link>
  )
}
