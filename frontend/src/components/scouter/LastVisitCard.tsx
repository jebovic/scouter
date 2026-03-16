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
  if (!mission || mission.archivedAt) return null

  const label = mission.phase === 'done'
    ? 'Voir votre dernier achat'
    : 'Reprendre où vous en étiez'

  return (
    <Link to={`/missions/${mission.slug}`} className={styles.card}>
      <span className={styles.icon} aria-hidden="true">{mission.icon ?? '📦'}</span>
      <div className={styles.content}>
        <span className={styles.label}>{label}</span>
        <span className={styles.name}>{mission.name}</span>
      </div>
      <span className={styles.arrow} aria-hidden="true">→</span>
    </Link>
  )
}
