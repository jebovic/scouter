type BadgeVariant =
  | 'buy'
  | 'flash-sale'
  | 'preorder'
  | 'defer'
  | 'watch'
  | 'crisis'
  | 'recommended'
  | 'rejected'
  | 'alternative'

interface BadgeProps {
  variant: BadgeVariant
  label?: string
}

const VARIANT_STYLES: Record<BadgeVariant, { color: string; bg: string }> = {
  buy:         { color: 'var(--badge-buy)',         bg: 'var(--badge-buy-bg)' },
  'flash-sale':{ color: 'var(--badge-flash-sale)',  bg: 'var(--badge-flash-bg)' },
  preorder:    { color: 'var(--badge-preorder)',    bg: 'var(--badge-preorder-bg)' },
  defer:       { color: 'var(--badge-defer)',       bg: 'var(--badge-defer-bg)' },
  watch:       { color: 'var(--badge-watch)',       bg: 'var(--badge-watch-bg)' },
  crisis:      { color: 'var(--badge-crisis)',      bg: 'var(--badge-crisis-bg)' },
  recommended: { color: 'var(--badge-recommended)', bg: 'var(--badge-recommended-bg)' },
  rejected:    { color: 'var(--badge-rejected)',    bg: 'var(--badge-rejected-bg)' },
  alternative: { color: 'var(--text-mid)',          bg: 'transparent' },
}

const LABELS: Record<BadgeVariant, string> = {
  buy: 'BUY',
  'flash-sale': 'FLASH',
  preorder: 'PRE',
  defer: 'DEFER',
  watch: 'WATCH',
  crisis: 'CRISIS',
  recommended: 'REC',
  rejected: 'REJECT',
  alternative: 'ALT',
}

export function Badge({ variant, label }: BadgeProps) {
  const { color, bg } = VARIANT_STYLES[variant] ?? VARIANT_STYLES.defer
  const isFlash = variant === 'flash-sale'

  return (
    <span
      style={{
        display: 'inline-block',
        padding: '2px 8px',
        borderRadius: 'var(--radius-sm)',
        fontSize: '0.7rem',
        fontFamily: 'var(--font-mono)',
        fontWeight: 600,
        letterSpacing: '0.08em',
        color,
        background: bg,
        border: `1px solid ${color}40`,
        animation: isFlash ? 'pulse-badge 1.5s ease-in-out infinite' : undefined,
      }}
    >
      {label ?? LABELS[variant]}
    </span>
  )
}
