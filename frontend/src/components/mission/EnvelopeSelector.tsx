import { useEnvelopes } from '../../hooks/useEnvelopes'
import styles from './EnvelopeSelector.module.css'

interface EnvelopeSelectorProps {
  value: string | null
  onChange: (id: string | null) => void
}

export function EnvelopeSelector({ value, onChange }: EnvelopeSelectorProps) {
  const { envelopes, isLoading } = useEnvelopes()

  function handleChange(e: React.ChangeEvent<HTMLSelectElement>) {
    const v = e.target.value
    onChange(v === '' ? null : v)
  }

  return (
    <div className={styles.wrapper}>
      <label className={styles.label}>ENVELOPE</label>
      <select
        className={styles.select}
        value={value ?? ''}
        onChange={handleChange}
        disabled={isLoading}
      >
        <option value="" className={styles.noEnvelope}>
          No envelope
        </option>
        {envelopes.map((env) => (
          <option key={env.id} value={env.id}>
            {env.name}
          </option>
        ))}
      </select>
    </div>
  )
}
