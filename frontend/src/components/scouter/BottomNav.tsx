import { NavLink, useNavigate } from 'react-router-dom'
import styles from './BottomNav.module.css'

interface BottomNavProps {
  missionSlug: string
}

export function BottomNav({ missionSlug }: BottomNavProps) {
  const navigate = useNavigate()
  const base = `/missions/${missionSlug}`

  return (
    <nav className={styles.bottomNav} aria-label="Mobile mission navigation">
      <NavLink
        to={base}
        end
        className={({ isActive }) =>
          `${styles.tab} ${isActive ? styles.tabActive : ''}`
        }
      >
        <span className={styles.tabIcon} aria-hidden="true">🏠</span>
        <span className={styles.tabLabel}>Overview</span>
      </NavLink>

      <NavLink
        to={`${base}/options`}
        className={({ isActive }) =>
          `${styles.tab} ${isActive ? styles.tabActive : ''}`
        }
      >
        <span className={styles.tabIcon} aria-hidden="true">🔍</span>
        <span className={styles.tabLabel}>Options</span>
      </NavLink>

      <NavLink
        to={`${base}/shopping`}
        className={({ isActive }) =>
          `${styles.tab} ${isActive ? styles.tabActive : ''}`
        }
      >
        <span className={styles.tabIcon} aria-hidden="true">🛒</span>
        <span className={styles.tabLabel}>Shopping</span>
      </NavLink>

      <button
        className={styles.tab}
        onClick={() => navigate(`${base}?research=1`)}
        type="button"
        aria-label="Trigger research"
      >
        <span className={styles.tabIcon} aria-hidden="true">⚡</span>
        <span className={styles.tabLabel}>Research</span>
      </button>
    </nav>
  )
}
