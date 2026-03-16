import { NavLink, useLocation } from 'react-router-dom'
import { Home, ShoppingCart, Plus, Tag, BarChart2 } from 'lucide-react'
import { useState } from 'react'
import styles from './GlobalBottomNav.module.css'

interface QuickAddMenuProps {
  onClose: () => void
}

function QuickAddMenu({ onClose }: QuickAddMenuProps) {
  return (
    <div className={styles.quickAddOverlay} onClick={onClose} role="presentation">
      <div className={styles.quickAddSheet} onClick={e => e.stopPropagation()} role="dialog" aria-modal="true" aria-label="Action rapide">
        <div className={styles.sheetHandle} />
        <h3 className={styles.sheetTitle}>Que voulez-vous faire ?</h3>
        <nav className={styles.sheetActions}>
          <NavLink to="/?new=1" className={styles.sheetAction} onClick={onClose}>
            <span className={styles.sheetActionIcon}><ShoppingCart size={20} /></span>
            <span>Nouveau projet d'achat</span>
          </NavLink>
          <NavLink to="/envelopes" className={styles.sheetAction} onClick={onClose}>
            <span className={styles.sheetActionIcon}><Plus size={20} /></span>
            <span>Enregistrer une dépense</span>
          </NavLink>
          <NavLink to="/wishlist" className={styles.sheetAction} onClick={onClose}>
            <span className={styles.sheetActionIcon}>❤</span>
            <span>Ajouter à ma liste de souhaits</span>
          </NavLink>
          <NavLink to="/notifications" className={styles.sheetAction} onClick={onClose}>
            <span className={styles.sheetActionIcon}><Tag size={20} /></span>
            <span>Signaler une alerte prix</span>
          </NavLink>
        </nav>
      </div>
    </div>
  )
}

export function GlobalBottomNav() {
  const location = useLocation()
  const [quickAddOpen, setQuickAddOpen] = useState(false)

  // Don't render inside mission pages — they have their own BottomNav
  if (location.pathname.startsWith('/missions/')) return null
  // Don't render on shared wishlist
  if (location.pathname.startsWith('/wishlist/shared')) return null

  const isAchats = ['/kanban', '/wishlist', '/notifications'].some(p =>
    location.pathname === p
  )
  const isPlans = ['/deal-calendar', '/insights', '/digest'].some(p =>
    location.pathname === p
  )
  const isStats = ['/stats', '/performance'].some(p =>
    location.pathname === p
  )

  return (
    <>
      {quickAddOpen && <QuickAddMenu onClose={() => setQuickAddOpen(false)} />}
      <nav
        className={styles.bottomNav}
        aria-label="Navigation globale"
        role="navigation"
      >
        <NavLink
          to="/"
          end
          className={({ isActive }) =>
            `${styles.tab} ${isActive ? styles.tabActive : ''}`
          }
        >
          <Home size={20} aria-hidden="true" />
          <span className={styles.tabLabel}>Accueil</span>
        </NavLink>

        <NavLink
          to="/kanban"
          className={`${styles.tab} ${isAchats ? styles.tabActive : ''}`}
        >
          <ShoppingCart size={20} aria-hidden="true" />
          <span className={styles.tabLabel}>Achats</span>
        </NavLink>

        <button
          className={`${styles.tab} ${styles.fabTab}`}
          onClick={() => setQuickAddOpen(true)}
          type="button"
          aria-label="Action rapide"
        >
          <span className={styles.fab} aria-hidden="true">
            <Plus size={22} />
          </span>
        </button>

        <NavLink
          to="/deal-calendar"
          className={`${styles.tab} ${isPlans ? styles.tabActive : ''}`}
        >
          <Tag size={20} aria-hidden="true" />
          <span className={styles.tabLabel}>Bons Plans</span>
        </NavLink>

        <NavLink
          to="/stats"
          className={`${styles.tab} ${isStats ? styles.tabActive : ''}`}
        >
          <BarChart2 size={20} aria-hidden="true" />
          <span className={styles.tabLabel}>Bilan</span>
        </NavLink>
      </nav>
    </>
  )
}
