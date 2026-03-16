import { useState, useEffect, useLayoutEffect } from 'react'
import { NavLink, Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import {
  Home, ShoppingCart, Heart, Bell, Package,
  Wallet, History, RotateCcw, CreditCard,
  CalendarDays, TrendingDown, FileText,
  BarChart2, Award,
  Settings, PanelLeftClose, PanelLeft,
} from 'lucide-react'
import type { Mission } from '../../types'
import { Logo } from './Logo'
import styles from './NavRail.module.css'

interface NavRailProps {
  missions: Mission[]
  onNewMission?: () => void
}

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
  const { t } = useTranslation()
  const [collapsed, setCollapsed] = useState(() => {
    try { return localStorage.getItem('railCollapsed') === 'true' } catch { return false }
  })
  const [simpleMode, setSimpleMode] = useState(() => {
    try { return localStorage.getItem('simpleMode') === 'true' } catch { return false }
  })

  const SECTIONS = [
    {
      key: 'achats',
      label: t('nav.rail.sectionAchats'),
      items: [
        { label: t('nav.rail.itemHq'),            path: '/',              icon: Home,          end: true },
        { label: t('nav.rail.itemKanban'),         path: '/kanban',        icon: ShoppingCart },
        { label: t('nav.rail.itemWishlist'),       path: '/wishlist',      icon: Heart },
        { label: t('nav.rail.itemNotifications'),  path: '/notifications', icon: Bell },
      ],
    },
    {
      key: 'budget',
      label: t('nav.rail.sectionBudget'),
      items: [
        { label: t('nav.rail.itemEnvelopes'), path: '/envelopes', icon: Wallet },
        { label: t('nav.rail.itemHistory'),   path: '/history',   icon: History },
        { label: t('nav.rail.itemCashback'),  path: '/cashback',  icon: RotateCcw },
        { label: t('nav.rail.itemLoyalty'),   path: '/loyalty',   icon: CreditCard },
      ],
    },
    {
      key: 'bons-plans',
      label: t('nav.rail.sectionBonsPlans'),
      items: [
        { label: t('nav.rail.itemDealCalendar'), path: '/deal-calendar', icon: CalendarDays },
        { label: t('nav.rail.itemInsights'),     path: '/insights',      icon: TrendingDown },
        { label: t('nav.rail.itemDigest'),       path: '/digest',        icon: FileText },
      ],
    },
    {
      key: 'stats',
      label: t('nav.rail.sectionStats'),
      items: [
        { label: t('nav.rail.itemStats'),       path: '/stats',       icon: BarChart2 },
        { label: t('nav.rail.itemPerformance'), path: '/performance', icon: Award },
      ],
    },
  ]

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
      aria-label={t('nav.mainNavLabel')}
    >
      {/* Logo */}
      <Link to="/" className={styles.logo} title={t('nav.logoTitle')}>
        {collapsed
          ? <Logo size="sm" iconOnly aria-label={t('nav.logoTitle')} />
          : <Logo size="sm" />
        }
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
          {!collapsed && <span className={styles.sectionLabel}>{t('nav.rail.sectionMissions')}</span>}
          <button
            className={`${styles.navItem} ${styles.newMissionBtn}`}
            onClick={onNewMission}
            title={t('nav.rail.newMission')}
            type="button"
          >
            <span className={styles.plusIcon} aria-hidden="true">+</span>
            {!collapsed && <span className={styles.label}>{t('nav.rail.newMission')}</span>}
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
              {t('nav.rail.moreMissions', { count: missions.length - 8 })}
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
          title={collapsed ? t('nav.rail.itemSettings') : undefined}
        >
          <Settings className={styles.icon} size={18} aria-hidden="true" />
          {!collapsed && <span className={styles.label}>{t('nav.rail.itemSettings')}</span>}
        </NavLink>
        <button
          className={`${styles.navItem} ${styles.collapseBtn}`}
          onClick={() => setCollapsed(c => !c)}
          aria-label={collapsed ? t('nav.rail.expand') : t('nav.rail.collapse')}
          title={collapsed ? t('nav.rail.expand') : t('nav.rail.collapse')}
          type="button"
        >
          {collapsed
            ? <PanelLeft size={16} aria-hidden="true" />
            : <PanelLeftClose size={16} aria-hidden="true" />
          }
          {!collapsed && <span className={styles.label}>{t('nav.rail.collapse')}</span>}
        </button>
      </div>
    </nav>
  )
}
