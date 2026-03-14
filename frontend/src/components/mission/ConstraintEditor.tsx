import { useState } from 'react'
import type { Constraint } from '../../types'
import styles from './ConstraintEditor.module.css'

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

  return (
    <div className={styles.root}>
      <div className={styles.sectionLabel}>CONSTRAINTS</div>

      <div className={styles.tagList}>
        {constraints.map((c) => (
          <span
            key={c.key}
            className={`${styles.tag} ${c.type === 'hard' ? styles.tagHard : styles.tagSoft}`}
          >
            {c.label}: {String(c.value)}
            <button onClick={() => remove(c.key)} className={styles.removeBtn}>
              ×
            </button>
          </span>
        ))}
      </div>

      <div className={styles.addRow}>
        <input
          className={`${styles.inputBase} ${styles.inputLabel}`}
          value={label}
          onChange={(e) => setLabel(e.target.value)}
          placeholder="Label"
          onKeyDown={(e) => e.key === 'Enter' && add()}
        />
        <input
          className={`${styles.inputBase} ${styles.inputValue}`}
          value={value}
          onChange={(e) => setValue(e.target.value)}
          placeholder="Value"
          onKeyDown={(e) => e.key === 'Enter' && add()}
        />
        <select
          className={styles.inputBase}
          value={type}
          onChange={(e) => setType(e.target.value as 'hard' | 'soft')}
        >
          <option value="hard">Hard</option>
          <option value="soft">Soft</option>
        </select>
        <button onClick={add} className={styles.addBtn}>
          +
        </button>
      </div>
    </div>
  )
}
