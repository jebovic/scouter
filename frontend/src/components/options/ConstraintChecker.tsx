import type { Option, Constraint } from '../../types'

interface ConstraintCheckerProps {
  option: Option
  constraints: Constraint[]
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
    <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
      <div style={{ fontSize: '0.7rem', color: 'var(--text-dim)', fontFamily: 'var(--font-mono)', letterSpacing: '0.05em' }}>
        CONSTRAINTS
      </div>
      {results.map(({ constraint, pass, attrValue }) => (
        <div
          key={constraint.key}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            padding: '4px 8px',
            borderRadius: 4,
            background:
              pass === true
                ? 'var(--green-dim)'
                : pass === false
                ? constraint.type === 'hard'
                  ? 'var(--coral-dim)'
                  : 'var(--gold-dim)'
                : 'var(--raised)',
            border: `1px solid ${
              pass === true
                ? 'var(--green)30'
                : pass === false
                ? constraint.type === 'hard'
                  ? 'var(--coral)30'
                  : 'var(--gold)30'
                : 'var(--border)'
            }`,
          }}
        >
          <span style={{ fontSize: '0.75rem' }}>
            {pass === true ? '✓' : pass === false ? '✗' : '?'}
          </span>
          <span style={{ fontSize: '0.75rem', color: 'var(--text-mid)', flex: 1 }}>
            {constraint.label}
          </span>
          {attrValue !== undefined && (
            <span style={{ fontSize: '0.7rem', color: 'var(--text-dim)', fontFamily: 'var(--font-mono)' }}>
              {String(attrValue)}
            </span>
          )}
          {constraint.type === 'hard' && (
            <span style={{ fontSize: '0.6rem', color: 'var(--coral)', fontFamily: 'var(--font-mono)' }}>
              HARD
            </span>
          )}
        </div>
      ))}
      {hardFails.length > 0 && (
        <div style={{ fontSize: '0.7rem', color: 'var(--coral)', marginTop: 4 }}>
          ⚠ {hardFails.length} hard constraint{hardFails.length > 1 ? 's' : ''} failed
        </div>
      )}
      {softFails.length > 0 && hardFails.length === 0 && (
        <div style={{ fontSize: '0.7rem', color: 'var(--gold)', marginTop: 4 }}>
          {softFails.length} soft constraint{softFails.length > 1 ? 's' : ''} not met
        </div>
      )}
    </div>
  )
}
