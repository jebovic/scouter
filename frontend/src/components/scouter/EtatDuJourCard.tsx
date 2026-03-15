import { Link } from 'react-router-dom'
import { useMissions, useUnreadCount } from '../../hooks'
import styles from './EtatDuJourCard.module.css'

export function EtatDuJourCard() {
  const { missions } = useMissions()
  const unreadCount = useUnreadCount()
  const activeMissions = missions.filter(m => m.phase === 'researching' || m.phase === 'comparing').length

  const items = [
    { count: activeMissions, label: 'missions actives', to: '/', show: activeMissions > 0 },
    { count: unreadCount, label: 'alertes prix', to: '/notifications', show: unreadCount > 0 },
  ].filter(i => i.show)

  if (items.length === 0) return null

  return (
    <div className={styles.card}>
      <span className={styles.title}>Aujourd'hui</span>
      <div className={styles.items}>
        {items.map((item, i) => (
          <span key={item.to} className={styles.item}>
            {i > 0 && <span className={styles.sep}>·</span>}
            <Link to={item.to} className={styles.link}>
              <strong>{item.count}</strong> {item.label}
            </Link>
          </span>
        ))}
      </div>
    </div>
  )
}
