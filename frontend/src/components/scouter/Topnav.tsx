import { Link, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useSidebar } from '../../contexts/sidebar'
import { useTheme } from '../../hooks/useTheme'
import { LLMStatus } from './LLMStatus'
import { NotificationBell } from './NotificationBell'
import { SearchDropdown } from './SearchDropdown'
import styles from './Topnav.module.css'

interface TopnavProps {
  missionSlug?: string
  missionName?: string
}

export function Topnav({ missionSlug, missionName }: TopnavProps) {
  const { t } = useTranslation()
  const location = useLocation()
  const { openSidebar } = useSidebar()
  const { isDark, toggleTheme } = useTheme()

  const isActive = (path: string) =>
    location.pathname === path || location.pathname.startsWith(path + '/')

  const isMissionRoot = (slug: string) =>
    location.pathname === `/missions/${slug}`

  return (
    <nav className={styles.nav}>
      {/* Sidebar toggle */}
      <button className={styles.menuBtn} onClick={openSidebar} aria-label="Open missions sidebar">
        ☰
      </button>

      {/* Logo */}
      <Link to="/" className={styles.logo}>
        {t('nav.hq')}
      </Link>

      {/* Global nav links — only show when not in a mission sub-nav */}
      {!missionSlug && (
        <div className={styles.subNav}>
          {[
            { label: 'Kanban', path: '/kanban' },
            { label: 'Cashback', path: '/cashback' },
            { label: 'Wishlist', path: '/wishlist' },
            { label: 'Envelopes', path: '/envelopes' },
            { label: 'Bons Plans', path: '/deal-calendar' },
            { label: 'Fidélité', path: '/loyalty' },
            { label: 'History', path: '/history' },
            { label: 'Stats', path: '/stats' },
            { label: 'Digest', path: '/digest' },
            { label: 'Perfs', path: '/performance' },
            { label: 'Settings', path: '/settings' },
          ].map(({ label, path }) => (
            <Link
              key={path}
              to={path}
              className={isActive(path) ? `${styles.navLink} ${styles.navLinkActive}` : styles.navLink}
            >
              {label}
            </Link>
          ))}
        </div>
      )}

      {missionSlug && (
        <>
          {/* Breadcrumb separator — comms style */}
          <span className={styles.separator}>//</span>

          <Link
            to={`/missions/${missionSlug}`}
            className={styles.missionLink}
            style={{
              '--mission-link-color': isMissionRoot(missionSlug)
                ? 'var(--text)'
                : 'var(--text-mid)',
            } as React.CSSProperties}
          >
            {missionName ?? missionSlug}
          </Link>

          <div className={styles.subNav}>
            {[
              { label: t('nav.overview'), path: `/missions/${missionSlug}` },
              { label: t('nav.options'),  path: `/missions/${missionSlug}/options` },
              { label: t('nav.shopping'), path: `/missions/${missionSlug}/shopping` },
            ].map(({ label, path }) => {
              const active = isActive(path)
              return (
                <Link
                  key={path}
                  to={path}
                  className={active
                    ? `${styles.navLink} ${styles.navLinkActive}`
                    : styles.navLink
                  }
                >
                  {label}
                </Link>
              )
            })}
          </div>
        </>
      )}

      <div className={styles.actions}>
        <SearchDropdown />
        <button
          className={styles.themeBtn}
          onClick={toggleTheme}
          aria-label="Toggle theme"
          title={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
        >
          {isDark ? '☀' : '🌙'}
        </button>
        <LLMStatus />
        <NotificationBell />
      </div>
    </nav>
  )
}
