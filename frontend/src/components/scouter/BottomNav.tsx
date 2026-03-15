import { NavLink, Link, useNavigate } from 'react-router-dom'
import { ClipboardList, GitCompare, ShoppingBag, Zap, ChevronLeft } from 'lucide-react'
import styles from './BottomNav.module.css'

interface BottomNavProps {
  missionSlug: string
}

export function BottomNav({ missionSlug }: BottomNavProps) {
  const navigate = useNavigate()
  const base = `/missions/${missionSlug}`

  return (
    <nav className={styles.bottomNav} aria-label="Navigation mission">
      <Link to="/" className={styles.backLink} aria-label="Toutes mes missions">
        <ChevronLeft size={16} aria-hidden="true" />
        <span className={styles.backLabel}>Missions</span>
      </Link>

      <NavLink
        to={base}
        end
        className={({ isActive }) =>
          `${styles.tab} ${isActive ? styles.tabActive : ''}`
        }
      >
        <ClipboardList size={20} aria-hidden="true" />
        <span className={styles.tabLabel}>Résumé</span>
      </NavLink>

      <NavLink
        to={`${base}/options`}
        className={({ isActive }) =>
          `${styles.tab} ${isActive ? styles.tabActive : ''}`
        }
      >
        <GitCompare size={20} aria-hidden="true" />
        <span className={styles.tabLabel}>Comparatif</span>
      </NavLink>

      <NavLink
        to={`${base}/shopping`}
        className={({ isActive }) =>
          `${styles.tab} ${isActive ? styles.tabActive : ''}`
        }
      >
        <ShoppingBag size={20} aria-hidden="true" />
        <span className={styles.tabLabel}>Mes achats</span>
      </NavLink>

      <button
        className={styles.tab}
        onClick={() => navigate(`${base}?research=1`)}
        type="button"
        aria-label="Lancer la recherche"
      >
        <Zap size={20} aria-hidden="true" />
        <span className={styles.tabLabel}>Recherche</span>
      </button>
    </nav>
  )
}
