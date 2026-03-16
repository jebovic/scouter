import { useState, useEffect, useLayoutEffect } from 'react'
import { NavLink, Link } from 'react-router-dom'
import {
  Home, ShoppingCart, Heart, Bell, Package,
  Wallet, History, RotateCcw, CreditCard,
  CalendarDays, TrendingDown, FileText,
  BarChart2, Award,
  Settings, PanelLeftClose, PanelLeft,
} from 'lucide-react'
import type { Mission } from '../../types'
import styles from './NavRail.module.css'

interface NavRailProps {
  missions: Mission[]
  onNewMission?: () => void
}

const SECTIONS = [
  {
    key: 'achats',
    label: 'MES ACHATS',
    items: [
      { label: 'Accueil',         path: '/',           icon: Home,          end: true },
      { label: 'Tableau achats',  path: '/kanban',     icon: ShoppingCart },
      { label: 'Liste de souhaits', path: '/wishlist', icon: Heart },
      { label: 'Alertes',         path: '/notifications', icon: Bell },
    ],
  },
  {
    key: 'budget',
    label: 'MON BUDGET',
    items: [
      { label: 'Budget par catégorie', path: '/envelopes', icon: Wallet },
      { label: 'Historique',     path: '/history',   icon: History },
      { label: 'Cashback',       path: '/cashback',  icon: RotateCcw },
      { label: 'Cartes fidélité', path: '/loyalty',  icon: CreditCard },
    ],
  },
  {
    key: 'bons-plans',
    label: 'BONS PLANS',
    items: [
      { label: 'Calendrier promos', path: '/deal-calendar', icon: CalendarDays },
      { label: 'Analyse',        path: '/insights',  icon: TrendingDown },
      { label: 'Résumé hebdo',   path: '/digest',    icon: FileText },
    ],
  },
  {
    key: 'stats',
    label: 'MON BILAN',
    items: [
      { label: 'Bilan général',  path: '/stats',       icon: BarChart2 },
      { label: 'Performances',   path: '/performance', icon: Award },
    ],
  },
]

const PHASE_DOT: Record<string, string> = {
  researching: 'var(--cyan)',
  comparing:   'var(--gold)',
  buying:      'var(--green)',
  done:        'var(--text-dim)',
}

function missionDotColor(mission: Mission): string {
  if (mission.archivedAt) return 'var(--text-dim)'
  return PHASE_DOT[mission.phase] ?? 'var(--text-dim)'
}

export function NavRail({ missions, onNewMission }: NavRailProps) {
  const [collapsed, setCollapsed] = useState(() => {
    try { return localStorage.getItem('railCollapsed') === 'true' } catch { return false }
  })
  const [simpleMode, setSimpleMode] = useState(() => {
    try { return localStorage.getItem('simpleMode') === 'true' } catch { return false }
  })

  useEffect(() => {
    function handleStorage(e: StorageEvent) {
      if (e.key === 'simpleMode') setSimpleMode(e.newValue === 'true')
    }
    window.addEventListener('storage', handleStorage)
    return () => window.removeEventListener('storage', handleStorage)
  }, [])

  useLayoutEffect(() => {
    try { localStorage.setItem('railCollapsed', String(collapsed)) } catch {}
    document.documentElement.style.setProperty(
      '--rail-offset',
      collapsed ? 'var(--rail-collapsed)' : 'var(--rail-width)'
    )
  }, [collapsed])

  // In simple mode, hide advanced tools
  const HIDDEN_SIMPLE = ['/kanban', '/performance', '/digest', '/insights', '/cashback', '/loyalty']

  const visibleSections = SECTIONS.map(section => ({
    ...section,
    items: simpleMode
      ? section.items.filter(item => !HIDDEN_SIMPLE.includes(item.path))
      : section.items,
  })).filter(section => section.items.length > 0)

  return (
    <nav
      className={`${styles.rail} ${collapsed ? styles.collapsed : ''}`}
      aria-label="Navigation principale"
    >
      {/* Logo */}
      <Link to="/" className={styles.logo} title="Accueil">
        <span className={styles.logoText}>
          {collapsed ? 'S' : 'SCOUTER'}
        </span>
      </Link>

      <div className={styles.sections}>
        {visibleSections.map(section => (
          <div key={section.key} className={styles.section}>
            {!collapsed && (
              <span className={styles.sectionLabel}>{section.label}</span>
            )}
            {section.items.map(({ label, path, icon: Icon, end }) => (
              <NavLink
                key={path}
                to={path}
                end={end}
                className={({ isActive }) =>
                  `${styles.navItem} ${isActive ? styles.active : ''}`
                }
                title={collapsed ? label : undefined}
              >
                <Icon className={styles.icon} size={18} aria-hidden="true" />
                {!collapsed && <span className={styles.label}>{label}</span>}
              </NavLink>
            ))}
          </div>
        ))}

        {/* Missions section */}
        <div className={styles.section}>
          {!collapsed && <span className={styles.sectionLabel}>MES MISSIONS</span>}
          <button
            className={`${styles.navItem} ${styles.newMissionBtn}`}
            onClick={onNewMission}
            title="Nouvelle mission"
            type="button"
          >
            <span className={styles.plusIcon} aria-hidden="true">+</span>
            {!collapsed && <span className={styles.label}>Nouvelle mission</span>}
          </button>
          {missions.slice(0, 8).map(m => (
            <NavLink
              key={m.id}
              to={`/missions/${m.slug}`}
              className={({ isActive }) =>
                `${styles.navItem} ${styles.missionItem} ${isActive ? styles.active : ''}`
              }
              title={m.name}
            >
              <Package className={styles.icon} size={16} aria-hidden="true" />
              {!collapsed && (
                <>
                  <span className={styles.missionName}>{m.name}</span>
                  <span
                    className={styles.statusDot}
                    style={{ background: missionDotColor(m) }}
                    aria-hidden="true"
                  />
                </>
              )}
              {collapsed && (
                <span
                  className={styles.statusDotCollapsed}
                  style={{ background: missionDotColor(m) }}
                  aria-hidden="true"
                />
              )}
            </NavLink>
          ))}
          {missions.length > 8 && !collapsed && (
            <Link to="/" className={styles.moreLink}>
              +{missions.length - 8} autres missions
            </Link>
          )}
        </div>
      </div>

      {/* Bottom: settings + collapse toggle */}
      <div className={styles.footer}>
        <NavLink
          to="/settings"
          className={({ isActive }) =>
            `${styles.navItem} ${isActive ? styles.active : ''}`
          }
          title={collapsed ? 'Paramètres' : undefined}
        >
          <Settings className={styles.icon} size={18} aria-hidden="true" />
          {!collapsed && <span className={styles.label}>Paramètres</span>}
        </NavLink>
        <button
          className={styles.collapseBtn}
          onClick={() => setCollapsed(c => !c)}
          aria-label={collapsed ? 'Agrandir la navigation' : 'Réduire la navigation'}
          title={collapsed ? 'Agrandir' : 'Réduire'}
          type="button"
        >
          {collapsed
            ? <PanelLeft size={16} aria-hidden="true" />
            : <PanelLeftClose size={16} aria-hidden="true" />
          }
          {!collapsed && <span className={styles.label}>Réduire</span>}
        </button>
      </div>
    </nav>
  )
}
