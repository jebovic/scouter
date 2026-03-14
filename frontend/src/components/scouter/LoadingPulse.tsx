interface LoadingPulseProps {
  label?: string
  color?: string
}

export function LoadingPulse({ label = 'Loading...', color = 'var(--cyan)' }: LoadingPulseProps) {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 14,
        color: 'var(--text-mid)',
        fontFamily: 'var(--font-mono)',
        fontSize: '0.8rem',
        letterSpacing: '0.1em',
        padding: '2rem',
      }}
    >
      {/* Radar scan bar */}
      <div
        style={{
          position: 'relative',
          width: 48,
          height: 4,
          borderRadius: 2,
          background: 'var(--raised)',
          overflow: 'hidden',
          flexShrink: 0,
        }}
      >
        <div
          style={{
            position: 'absolute',
            top: 0,
            left: 0,
            height: '100%',
            width: 14,
            background: `linear-gradient(90deg, transparent, ${color}, transparent)`,
            borderRadius: 2,
            animation: 'scan-line 1.4s ease-in-out infinite',
            boxShadow: `0 0 8px ${color}`,
          }}
        />
      </div>
      <span style={{ color: 'var(--text-dim)', textTransform: 'uppercase' }}>
        {label}
      </span>
    </div>
  )
}
