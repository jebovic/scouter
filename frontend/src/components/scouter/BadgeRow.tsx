import { useMemo } from 'react'
import type { Badge } from '../../utils/badges'
import styles from './BadgeRow.module.css'

interface BadgeRowProps {
  badges: Badge[]
}

export function BadgeRow({ badges }: BadgeRowProps) {
  const sorted = useMemo(() => {
    // Sort: unlocked first (by date desc), then locked
    return [...badges].sort((a, b) => {
      // Unlocked badges come first
      const aUnlocked = a.unlockedAt ? 1 : 0
      const bUnlocked = b.unlockedAt ? 1 : 0
      if (aUnlocked !== bUnlocked) {
        return bUnlocked - aUnlocked
      }
      // Within same unlock status, sort by date desc
      if (a.unlockedAt && b.unlockedAt) {
        return b.unlockedAt.getTime() - a.unlockedAt.getTime()
      }
      return 0
    })
  }, [badges])

  const formatDate = (date: Date): string => {
    return new Intl.DateTimeFormat('fr-FR', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    }).format(date)
  }

  return (
    <div className={styles.container} role="list">
      {sorted.map((badge) => (
        <div
          key={badge.id}
          className={`${styles.badge} ${badge.unlockedAt ? styles.unlocked : styles.locked}`}
          role="listitem"
          title={
            badge.unlockedAt
              ? `${badge.name} — ${badge.description}\nDébloqué le ${formatDate(badge.unlockedAt)}`
              : `🔒 ${badge.name} — Verrouillé`
          }
        >
          <div className={styles.icon}>{badge.icon}</div>
          <div className={styles.info}>
            <div className={styles.name}>{badge.name}</div>
            {badge.unlockedAt && (
              <div className={styles.date}>{formatDate(badge.unlockedAt)}</div>
            )}
            {!badge.unlockedAt && <div className={styles.locked_text}>Verrouillé</div>}
          </div>
        </div>
      ))}
    </div>
  )
}
