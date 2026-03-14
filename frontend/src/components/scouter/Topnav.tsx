import { Link, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

interface TopnavProps {
  missionSlug?: string
  missionName?: string
}

export function Topnav({ missionSlug, missionName }: TopnavProps) {
  const { t } = useTranslation()
  const location = useLocation()

  const isActive = (path: string) =>
    location.pathname === path || location.pathname.startsWith(path + '/')

  const isMissionRoot = (slug: string) =>
    location.pathname === `/missions/${slug}` ||
    (!location.pathname.startsWith(`/missions/${slug}/options`) &&
     !location.pathname.startsWith(`/missions/${slug}/shopping`))

  return (
    <nav
      style={{
        position: 'sticky',
        top: 0,
        zIndex: 100,
        background: 'rgba(10, 14, 26, 0.92)',
        backdropFilter: 'blur(12px)',
        WebkitBackdropFilter: 'blur(12px)',
        borderBottom: '1px solid var(--border)',
        boxShadow: '0 1px 0 rgba(0, 229, 255, 0.08), 0 4px 20px rgba(0,0,0,0.4)',
        display: 'flex',
        alignItems: 'center',
        gap: '1.5rem',
        padding: '0 1.5rem',
        height: 52,
      }}
    >
      {/* Logo */}
      <Link
        to="/"
        style={{
          fontFamily: 'var(--font-display)',
          color: 'var(--cyan)',
          letterSpacing: '0.15em',
          fontSize: '0.95rem',
          fontWeight: 700,
          flexShrink: 0,
          textShadow: '0 0 12px var(--cyan-dim)',
          textDecoration: 'none',
        }}
      >
        {t('nav.hq')}
      </Link>

      {missionSlug && (
        <>
          {/* Breadcrumb separator — comms style */}
          <span
            style={{
              color: 'var(--border-glow)',
              fontFamily: 'var(--font-mono)',
              fontSize: '0.7rem',
              letterSpacing: '0.1em',
              flexShrink: 0,
            }}
          >
            //
          </span>

          <Link
            to={`/missions/${missionSlug}`}
            style={{
              color: isMissionRoot(missionSlug) ? 'var(--text)' : 'var(--text-mid)',
              fontSize: '0.85rem',
              whiteSpace: 'nowrap',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              maxWidth: 200,
              fontFamily: 'var(--font-display)',
              letterSpacing: '0.05em',
              textDecoration: 'none',
              transition: 'color 0.15s',
            }}
          >
            {missionName ?? missionSlug}
          </Link>

          <div style={{ display: 'flex', gap: '0.5rem', marginLeft: 'auto', height: '100%', alignItems: 'center' }}>
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
                  className={active ? 'nav-link-active' : ''}
                  style={{
                    padding: '4px 12px',
                    borderRadius: 'var(--radius-sm)',
                    fontSize: '0.78rem',
                    fontFamily: 'var(--font-display)',
                    letterSpacing: '0.08em',
                    color: active ? 'var(--cyan)' : 'var(--text-dim)',
                    background: active ? 'var(--cyan-dim)' : 'transparent',
                    border: `1px solid ${active ? 'rgba(0,229,255,0.25)' : 'transparent'}`,
                    transition: 'all 0.15s',
                    textDecoration: 'none',
                    display: 'inline-block',
                    position: 'relative',
                  }}
                >
                  {label}
                </Link>
              )
            })}
          </div>
        </>
      )}
    </nav>
  )
}
