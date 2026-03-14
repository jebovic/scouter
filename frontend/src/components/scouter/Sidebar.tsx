import { Link } from 'react-router-dom'
import type { Mission } from '../../types'
import styles from './Sidebar.module.css'

interface SidebarProps {
  open: boolean
  missions: Mission[]
  onClose: () => void
}

export function Sidebar({ open, missions, onClose }: SidebarProps) {
  if (!open) return null

  return (
    <>
      <div role="presentation" className={styles.backdrop} onClick={onClose} />
      <nav aria-label="Missions" className={styles.panel}>
        <div className={styles.header}>
          <span className={styles.title}>Missions</span>
          <button className={styles.closeBtn} onClick={onClose} aria-label="Close sidebar">
            ✕
          </button>
        </div>
        {missions.length === 0 ? (
          <p className={styles.empty}>No missions yet</p>
        ) : (
          <ul className={styles.list}>
            {missions.map((m) => (
              <li key={m.id}>
                <Link to={`/missions/${m.slug}`} className={styles.item} onClick={onClose}>
                  <span className={styles.icon}>{m.icon}</span>
                  <span>{m.name}</span>
                </Link>
              </li>
            ))}
          </ul>
        )}
      </nav>
    </>
  )
}
