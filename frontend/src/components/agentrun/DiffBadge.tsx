import type { AgentDiff } from '../../api'

interface DiffBadgeProps {
  diff: AgentDiff | null
}

export function DiffBadge({ diff }: DiffBadgeProps) {
  if (!diff) return null

  const added = diff.added.length
  const removed = diff.removed.length
  const changed = diff.changed.length

  if (added === 0 && removed === 0 && changed === 0) {
    return (
      <span
        style={{
          fontSize: '0.65rem',
          fontFamily: 'var(--font-mono)',
          color: 'var(--text-dim)',
          padding: '1px 6px',
          border: '1px solid var(--border)',
          borderRadius: 4,
        }}
      >
        no changes
      </span>
    )
  }

  return (
    <span style={{ display: 'inline-flex', gap: 4, alignItems: 'center' }}>
      {added > 0 && (
        <span
          style={{
            fontSize: '0.65rem',
            fontFamily: 'var(--font-mono)',
            color: 'var(--green)',
            padding: '1px 6px',
            background: 'var(--green-dim)',
            border: '1px solid var(--green-dim)',
            borderRadius: 4,
          }}
        >
          +{added}
        </span>
      )}
      {removed > 0 && (
        <span
          style={{
            fontSize: '0.65rem',
            fontFamily: 'var(--font-mono)',
            color: 'var(--coral)',
            padding: '1px 6px',
            background: 'var(--coral-dim)',
            border: '1px solid var(--coral-dim)',
            borderRadius: 4,
          }}
        >
          -{removed}
        </span>
      )}
      {changed > 0 && (
        <span
          style={{
            fontSize: '0.65rem',
            fontFamily: 'var(--font-mono)',
            color: 'var(--gold)',
            padding: '1px 6px',
            background: 'var(--gold-dim)',
            border: '1px solid var(--gold-dim)',
            borderRadius: 4,
          }}
        >
          ~{changed}
        </span>
      )}
    </span>
  )
}
