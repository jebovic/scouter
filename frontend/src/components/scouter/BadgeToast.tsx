import { useEffect, useState } from 'react'
import type { Badge } from '../../utils/badges'
import styles from './BadgeToast.module.css'

interface BadgeToastProps {
  badge: Badge
  onDismiss?: () => void
  duration?: number // ms, default 4000
}

export function BadgeToast({ badge, onDismiss, duration = 4000 }: BadgeToastProps) {
  const [isVisible, setIsVisible] = useState(true)

  useEffect(() => {
    const timer = setTimeout(() => {
      setIsVisible(false)
      onDismiss?.()
    }, duration)

    return () => clearTimeout(timer)
  }, [duration, onDismiss])

  if (!isVisible) return null

  return (
    <div className={styles.toast}>
      <div className={styles.icon}>{badge.icon}</div>
      <div className={styles.content}>
        <div className={styles.title}>{badge.name}</div>
        <div className={styles.subtitle}>Badge débloqué !</div>
      </div>
    </div>
  )
}
