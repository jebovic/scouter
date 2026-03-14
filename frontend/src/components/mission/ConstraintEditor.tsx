import { useState } from 'react'
import type { Constraint } from '../../types'

interface ConstraintEditorProps {
  constraints: Constraint[]
  onChange: (constraints: Constraint[]) => void
}

export function ConstraintEditor({ constraints, onChange }: ConstraintEditorProps) {
  const [label, setLabel] = useState('')
  const [value, setValue] = useState('')
  const [type, setType] = useState<'hard' | 'soft'>('hard')

  function add() {
    if (!label.trim()) return
    const key = label.toLowerCase().replace(/\s+/g, '_')
    onChange([...constraints, { key, label: label.trim(), value, type }])
    setLabel('')
    setValue('')
  }

  function remove(key: string) {
    onChange(constraints.filter((c) => c.key !== key))
  }

  const inputStyle: React.CSSProperties = {
    background: 'var(--raised)',
    border: '1px solid var(--border)',
    borderRadius: 'var(--radius-sm)',
    color: 'var(--text)',
    padding: '6px 10px',
    fontSize: '0.8rem',
    fontFamily: 'var(--font-body)',
    outline: 'none',
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      <div style={{ fontSize: '0.75rem', color: 'var(--text-dim)', fontFamily: 'var(--font-mono)', letterSpacing: '0.05em' }}>
        CONSTRAINTS
      </div>

      <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
        {constraints.map((c) => (
          <span
            key={c.key}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 4,
              padding: '4px 10px',
              borderRadius: 4,
              fontSize: '0.75rem',
              background: c.type === 'hard' ? 'var(--coral-dim)' : 'var(--raised)',
              color: c.type === 'hard' ? 'var(--coral)' : 'var(--text-mid)',
              border: `1px solid ${c.type === 'hard' ? 'var(--coral)30' : 'var(--border)'}`,
            }}
          >
            {c.label}: {String(c.value)}
            <button
              onClick={() => remove(c.key)}
              style={{
                background: 'none',
                border: 'none',
                color: 'inherit',
                cursor: 'pointer',
                fontSize: '0.8rem',
                padding: 0,
                lineHeight: 1,
                opacity: 0.6,
              }}
            >
              ×
            </button>
          </span>
        ))}
      </div>

      <div style={{ display: 'flex', gap: 6 }}>
        <input
          style={{ ...inputStyle, flex: 2 }}
          value={label}
          onChange={(e) => setLabel(e.target.value)}
          placeholder="Label"
          onKeyDown={(e) => e.key === 'Enter' && add()}
        />
        <input
          style={{ ...inputStyle, flex: 1 }}
          value={value}
          onChange={(e) => setValue(e.target.value)}
          placeholder="Value"
          onKeyDown={(e) => e.key === 'Enter' && add()}
        />
        <select
          style={{ ...inputStyle }}
          value={type}
          onChange={(e) => setType(e.target.value as 'hard' | 'soft')}
        >
          <option value="hard">Hard</option>
          <option value="soft">Soft</option>
        </select>
        <button
          onClick={add}
          style={{
            padding: '6px 14px',
            borderRadius: 'var(--radius-sm)',
            border: '1px solid var(--border)',
            background: 'var(--raised)',
            color: 'var(--cyan)',
            cursor: 'pointer',
            fontSize: '0.85rem',
          }}
        >
          +
        </button>
      </div>
    </div>
  )
}
