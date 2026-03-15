import type { Option, Constraint } from '../../types'
import styles from './ConstraintChecker.module.css'

interface ConstraintCheckerProps {
  option: Option
  constraints: Constraint[]
}

function getConstraintVars(pass: boolean | null, type: 'hard' | 'soft'): React.CSSProperties {
  if (pass === true) {
    return {
      '--constraint-bg': 'var(--green-dim)',
      '--constraint-border': 'var(--green)30',
    } as React.CSSProperties
  }
  if (pass === false) {
    return type === 'hard'
      ? ({
          '--constraint-bg': 'var(--coral-dim)',
          '--constraint-border': 'var(--coral-dim)',
        } as React.CSSProperties)
      : ({
          '--constraint-bg': 'var(--gold-dim)',
          '--constraint-border': 'var(--gold-dim)',
        } as React.CSSProperties)
  }
  return {
    '--constraint-bg': 'var(--raised)',
    '--constraint-border': 'var(--border)',
  } as React.CSSProperties
}

export function ConstraintChecker({ option, constraints }: ConstraintCheckerProps) {
  if (constraints.length === 0) return null

  const results = constraints.map((c) => {
    const attr = option.attributes.find((a) => a.key === c.key)
    const pass = attr?.pass ?? null
    return { constraint: c, pass, attrValue: attr?.value }
  })

  const hardFails = results.filter((r) => r.constraint.type === 'hard' && r.pass === false)
  const softFails = results.filter((r) => r.constraint.type === 'soft' && r.pass === false)

  return (
    <div className={styles.root}>
      <div className={styles.heading}>CONSTRAINTS</div>
      {results.map(({ constraint, pass, attrValue }) => (
        <div
          key={constraint.key}
          className={styles.constraintRow}
          style={getConstraintVars(pass, constraint.type)}
        >
          <span className={styles.icon}>
            {pass === true ? '✓' : pass === false ? '✗' : '?'}
          </span>
          <span className={styles.constraintLabel}>{constraint.label}</span>
          {attrValue !== undefined && (
            <span className={styles.attrValue}>{String(attrValue)}</span>
          )}
          {constraint.type === 'hard' && (
            <span className={styles.hardBadge}>HARD</span>
          )}
        </div>
      ))}
      {hardFails.length > 0 && (
        <div className={styles.hardFailSummary}>
          ⚠ {hardFails.length} hard constraint{hardFails.length > 1 ? 's' : ''} failed
        </div>
      )}
      {softFails.length > 0 && hardFails.length === 0 && (
        <div className={styles.softFailSummary}>
          {softFails.length} soft constraint{softFails.length > 1 ? 's' : ''} not met
        </div>
      )}
    </div>
  )
}
