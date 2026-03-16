import { Link, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Menu, Sun, Moon } from 'lucide-react'
import { useSidebar } from '../../contexts/sidebar'
import { useTheme } from '../../hooks/useTheme'
import { LLMStatus } from './LLMStatus'
import { NotificationBell } from './NotificationBell'
import { SearchDropdown } from './SearchDropdown'
import styles from './Topnav.module.css'
import { Logo } from './Logo'

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
      {/* Sidebar toggle — mobile only (NavRail handles desktop) */}
      <button className={styles.menuBtn} onClick={openSidebar} aria-label={t('common.openSidebar')}>
        <Menu size={18} aria-hidden="true" />
      </button>

      {/* Logo */}
      <Link to="/" className={styles.logo} aria-label="SCOUTER — Home">
        <Logo size="md" />
      </Link>

      {/* Mission breadcrumb + sub-nav — only when inside a mission */}
      {missionSlug && (
        <>
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
          aria-label={isDark ? t('common.switchToLight') : t('common.switchToDark')}
          title={isDark ? t('common.switchToLight') : t('common.switchToDark')}
        >
          {isDark ? <Sun size={15} aria-hidden="true" /> : <Moon size={15} aria-hidden="true" />}
        </button>
        <LLMStatus />
        <NotificationBell />
      </div>
    </nav>
  )
}
