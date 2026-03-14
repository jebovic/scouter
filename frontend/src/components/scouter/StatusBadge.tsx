import type { MissionPhase } from '../../types'
import styles from './StatusBadge.module.css'

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
      className={styles.badge}
      style={{
        '--phase-color': color,
        '--phase-bg': `${color}18`,
        '--phase-border': `${color}40`,
      } as React.CSSProperties}
    >
      <span className={styles.dot} />
      {PHASE_LABELS[phase]}
    </span>
  )
}
