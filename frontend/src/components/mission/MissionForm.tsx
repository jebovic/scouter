import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Card } from '../scouter'
import type { MissionCreateRequest, MissionCategory } from '../../types'

const CATEGORIES: { value: MissionCategory; emoji: string }[] = [
  { value: 'travel', emoji: '✈️' },
  { value: 'electronics', emoji: '📱' },
  { value: 'computing', emoji: '💻' },
  { value: 'renovation', emoji: '🏠' },
  { value: 'custom', emoji: '🎯' },
]

const CURRENCIES = ['USD', 'EUR', 'GBP', 'CAD', 'AUD', 'JPY']

interface MissionFormProps {
  onSubmit: (req: MissionCreateRequest) => void
  onCancel: () => void
  loading?: boolean
  error?: string
}

export function MissionForm({ onSubmit, onCancel, loading, error }: MissionFormProps) {
  const { t } = useTranslation()
  const [name, setName] = useState('')
  const [icon, setIcon] = useState('🎯')
  const [category, setCategory] = useState<MissionCategory>('custom')
  const [budget, setBudget] = useState('')
  const [currency, setCurrency] = useState('USD')

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim() || !budget) return
    onSubmit({
      name: name.trim(),
      icon,
      category,
      budget: parseFloat(budget),
      currency,
      locale: navigator.language,
      constraints: [],
      costCategories: [],
    })
  }

  const inputStyle: React.CSSProperties = {
    background: 'var(--raised)',
    border: '1px solid var(--border)',
    borderRadius: 'var(--radius-sm)',
    color: 'var(--text)',
    padding: '8px 12px',
    fontSize: '0.875rem',
    fontFamily: 'var(--font-body)',
    width: '100%',
    outline: 'none',
  }

  const labelStyle: React.CSSProperties = {
    fontSize: '0.75rem',
    color: 'var(--text-dim)',
    display: 'block',
    marginBottom: 4,
    fontFamily: 'var(--font-mono)',
    letterSpacing: '0.05em',
  }

  return (
    <Card style={{ maxWidth: 480, width: '100%' }}>
      <h3
        style={{
          fontFamily: 'var(--font-display)',
          color: 'var(--cyan)',
          marginBottom: '1.5rem',
          letterSpacing: '0.05em',
        }}
      >
        {t('mission.create')}
      </h3>
      <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        <div>
          <label style={labelStyle}>MISSION NAME</label>
          <input
            style={inputStyle}
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. MacBook Pro upgrade"
            required
            autoFocus
          />
        </div>

        <div>
          <label style={labelStyle}>CATEGORY</label>
          <div style={{ display: 'flex', gap: 8 }}>
            {CATEGORIES.map((c) => (
              <button
                key={c.value}
                type="button"
                onClick={() => {
                  setCategory(c.value)
                  setIcon(c.emoji)
                }}
                style={{
                  flex: 1,
                  padding: '8px 4px',
                  borderRadius: 'var(--radius-sm)',
                  border: `1px solid ${category === c.value ? 'var(--cyan)' : 'var(--border)'}`,
                  background: category === c.value ? 'var(--cyan-dim)' : 'var(--raised)',
                  color: category === c.value ? 'var(--cyan)' : 'var(--text-mid)',
                  cursor: 'pointer',
                  fontSize: '1rem',
                  transition: 'all 0.15s',
                }}
                title={t(`mission.categories.${c.value}`)}
              >
                {c.emoji}
              </button>
            ))}
          </div>
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: '1fr auto', gap: 8 }}>
          <div>
            <label style={labelStyle}>BUDGET</label>
            <input
              style={inputStyle}
              type="number"
              min="0"
              step="0.01"
              value={budget}
              onChange={(e) => setBudget(e.target.value)}
              placeholder="5000"
              required
            />
          </div>
          <div>
            <label style={labelStyle}>CURRENCY</label>
            <select
              style={{ ...inputStyle, width: 'auto' }}
              value={currency}
              onChange={(e) => setCurrency(e.target.value)}
            >
              {CURRENCIES.map((c) => (
                <option key={c} value={c}>{c}</option>
              ))}
            </select>
          </div>
        </div>

        {error && (
          <div style={{ color: 'var(--coral)', fontSize: '0.8rem', fontFamily: 'var(--font-mono)', padding: '6px 0' }}>
            {error}
          </div>
        )}

        <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
          <button
            type="submit"
            disabled={loading}
            style={{
              flex: 1,
              padding: '10px',
              borderRadius: 'var(--radius-sm)',
              border: '1px solid var(--cyan)',
              background: loading ? 'var(--raised)' : 'var(--cyan-dim)',
              color: 'var(--cyan)',
              cursor: loading ? 'not-allowed' : 'pointer',
              fontFamily: 'var(--font-display)',
              fontWeight: 700,
              fontSize: '0.875rem',
              letterSpacing: '0.05em',
              transition: 'all 0.15s',
            }}
          >
            {loading ? 'LAUNCHING...' : 'LAUNCH MISSION'}
          </button>
          <button
            type="button"
            onClick={onCancel}
            style={{
              padding: '10px 16px',
              borderRadius: 'var(--radius-sm)',
              border: '1px solid var(--border)',
              background: 'transparent',
              color: 'var(--text-dim)',
              cursor: 'pointer',
              fontFamily: 'var(--font-body)',
              fontSize: '0.875rem',
            }}
          >
            Cancel
          </button>
        </div>
      </form>
    </Card>
  )
}
