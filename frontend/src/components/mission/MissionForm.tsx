import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Card } from '../scouter'
import type { MissionCreateRequest, MissionCategory } from '../../types'
import styles from './MissionForm.module.css'

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

  return (
    <Card className={styles.card}>
      <h3 className={styles.title}>{t('mission.create')}</h3>
      <form onSubmit={handleSubmit} className={styles.form}>
        <div>
          <label className={styles.label}>MISSION NAME</label>
          <input
            className={styles.input}
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. MacBook Pro upgrade"
            required
            autoFocus
          />
        </div>

        <div>
          <label className={styles.label}>CATEGORY</label>
          <div className={styles.categoryRow}>
            {CATEGORIES.map((c) => (
              <button
                key={c.value}
                type="button"
                onClick={() => {
                  setCategory(c.value)
                  setIcon(c.emoji)
                }}
                className={`${styles.categoryBtn} ${
                  category === c.value ? styles.categoryBtnActive : styles.categoryBtnInactive
                }`}
                title={t(`mission.categories.${c.value}`)}
              >
                {c.emoji}
              </button>
            ))}
          </div>
        </div>

        <div className={styles.budgetRow}>
          <div>
            <label className={styles.label}>BUDGET</label>
            <input
              className={styles.input}
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
            <label className={styles.label}>CURRENCY</label>
            <select
              className={styles.selectNarrow}
              value={currency}
              onChange={(e) => setCurrency(e.target.value)}
            >
              {CURRENCIES.map((c) => (
                <option key={c} value={c}>{c}</option>
              ))}
            </select>
          </div>
        </div>

        {error && <div className={styles.error}>{error}</div>}

        <div className={styles.actions}>
          <button
            type="submit"
            disabled={loading}
            className={`${styles.submitBtn} ${loading ? styles.submitBtnDisabled : styles.submitBtnActive}`}
          >
            {loading ? 'LAUNCHING...' : 'LAUNCH MISSION'}
          </button>
          <button
            type="button"
            onClick={onCancel}
            className={styles.cancelBtn}
          >
            Cancel
          </button>
        </div>
      </form>
    </Card>
  )
}
