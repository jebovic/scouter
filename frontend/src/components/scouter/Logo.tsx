import { useId } from 'react'
import styles from './Logo.module.css'

export interface LogoProps {
  size?: 'sm' | 'md' | 'lg'
  showTagline?: boolean
  iconOnly?: boolean
  variant?: 'dark' | 'light'
  gradientId?: string
  className?: string
  'aria-label'?: string
}

const ICON_SIZE: Record<NonNullable<LogoProps['size']>, number> = {
  sm: 28,
  md: 40,
  lg: 52,
}

export function Logo({
  size = 'md',
  showTagline = false,
  iconOnly = false,
  variant,
  gradientId,
  className,
  'aria-label': ariaLabel = 'SCOUTER',
}: LogoProps) {
  const autoId = useId()
  const prefix = gradientId ?? autoId.replace(/:/g, 'id')
  const gradId = `${prefix}-grad`
  const clipId = `${prefix}-lens`
  const px = ICON_SIZE[size]

  // Detect effective variant: explicit prop > prefers-color-scheme > dark
  const effectiveVariant: 'dark' | 'light' =
    variant ??
    (typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: light)').matches
      ? 'light'
      : 'dark')

  const c1 = effectiveVariant === 'light' ? '#0099bb' : '#00e5ff'
  const c2 = effectiveVariant === 'light' ? '#7c3aed' : '#a855f7'

  const svgSharedProps = iconOnly
    ? { role: 'img' as const, 'aria-label': ariaLabel }
    : { 'aria-hidden': true as const }

  const icon =
    size === 'sm' ? (
      // Simplified mark — no trend line (too fine at ≤28px)
      <svg
        width={px}
        height={px}
        viewBox="0 0 80 80"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        {...svgSharedProps}
      >
        <defs>
          <linearGradient id={gradId} x1="8" y1="8" x2="72" y2="72" gradientUnits="userSpaceOnUse">
            <stop offset="0%" stopColor={c1} />
            <stop offset="100%" stopColor={c2} />
          </linearGradient>
        </defs>
        <circle cx="34" cy="34" r="22" stroke={`url(#${gradId})`} strokeWidth="5" fill="none" />
        <line x1="51" y1="51" x2="67" y2="67" stroke={`url(#${gradId})`} strokeWidth="7" strokeLinecap="round" />
        <circle cx="34" cy="34" r="6" fill={`url(#${gradId})`} />
      </svg>
    ) : (
      // Full detail mark with price trend line inside lens
      <svg
        width={px}
        height={px}
        viewBox="0 0 80 80"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        {...svgSharedProps}
      >
        <defs>
          <linearGradient id={gradId} x1="8" y1="8" x2="72" y2="72" gradientUnits="userSpaceOnUse">
            <stop offset="0%" stopColor={c1} />
            <stop offset="100%" stopColor={c2} />
          </linearGradient>
          <clipPath id={clipId}>
            <circle cx="34" cy="34" r="21" />
          </clipPath>
        </defs>
        {/* Outer lens ring */}
        <circle cx="34" cy="34" r="22" stroke={`url(#${gradId})`} strokeWidth="3" fill="none" />
        {/* Price trend line clipped to lens interior */}
        <g clipPath={`url(#${clipId})`}>
          <polyline
            points="16,42 22,38 27,41 32,33 37,30 42,24 48,22"
            stroke={c1}
            strokeWidth="2.5"
            strokeLinecap="round"
            strokeLinejoin="round"
            fill="none"
            opacity={0.85}
          />
          {/* Trend tip dot */}
          <circle cx="48" cy="22" r="2.5" fill={c1} />
        </g>
        {/* Handle */}
        <line x1="51" y1="51" x2="67" y2="67" stroke={`url(#${gradId})`} strokeWidth="4" strokeLinecap="round" />
      </svg>
    )

  const gradientVars = {
    '--logo-start': c1,
    '--logo-end': c2,
  } as React.CSSProperties

  if (iconOnly) {
    return (
      <span
        className={[styles.root, styles[size], className].filter(Boolean).join(' ')}
        style={gradientVars}
      >
        {icon}
      </span>
    )
  }

  return (
    <span
      className={[styles.root, styles[size], className].filter(Boolean).join(' ')}
      style={gradientVars}
    >
      {icon}
      <span className={styles.text}>
        <span className={styles.wordmark}>SCOUTER</span>
        {size === 'lg' && showTagline && (
          <span className={styles.tagline}>spending intelligence</span>
        )}
      </span>
    </span>
  )
}
