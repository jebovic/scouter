import type { MissionPhase } from '../../types'

interface StatusBadgeProps {
  phase: MissionPhase
}

const PHASE_COLORS: Record<MissionPhase, string> = {
  researching: 'var(--cyan)',
  comparing:   'var(--gold)',
  buying:      'var(--green)',
  done:        'var(--text-dim)',
}

const PHASE_LABELS: Record<MissionPhase, string> = {
  researching: 'RESEARCH',
  comparing:   'COMPARE',
  buying:      'BUYING',
  done:        'DONE',
}

export function StatusBadge({ phase }: StatusBadgeProps) {
  const color = PHASE_COLORS[phase]

  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 6,
        padding: '3px 10px',
        borderRadius: 'var(--radius-sm)',
        fontSize: '0.7rem',
        fontFamily: 'var(--font-mono)',
        fontWeight: 600,
        letterSpacing: '0.1em',
        color,
        background: `${color}18`,
        border: `1px solid ${color}40`,
      }}
    >
      <span
        style={{
          width: 6,
          height: 6,
          borderRadius: '50%',
          background: color,
          boxShadow: `0 0 6px ${color}`,
          flexShrink: 0,
        }}
      />
      {PHASE_LABELS[phase]}
    </span>
  )
}
