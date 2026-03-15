import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Card } from '../scouter'
import type { MissionCreateRequest, MissionCategory } from '../../types'
import { EnvelopeSelector } from './EnvelopeSelector'
import { VoiceInputButton } from './VoiceInputButton'
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
  onSubmit: (req: MissionCreateRequest & { envelopeId?: string | null }) => void
  onCancel: () => void
  loading?: boolean
  error?: string
  initialValues?: Partial<MissionCreateRequest & { envelopeId?: string | null }>
}

export function MissionForm({ onSubmit, onCancel, loading, error, initialValues }: MissionFormProps) {
  const { t } = useTranslation()
  const [name, setName] = useState(initialValues?.name ?? '')
  const [icon, setIcon] = useState(initialValues?.icon ?? '🎯')
  const [category, setCategory] = useState<MissionCategory>(initialValues?.category ?? 'custom')
  const [budget, setBudget] = useState(initialValues?.budget != null ? String(initialValues.budget) : '')
  const [currency, setCurrency] = useState(initialValues?.currency ?? 'USD')
  const [envelopeId, setEnvelopeId] = useState<string | null>(initialValues?.envelopeId ?? null)

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
      constraints: initialValues?.constraints ?? [],
      costCategories: initialValues?.costCategories ?? [],
      envelopeId,
    })
  }

  return (
    <Card className={styles.card}>
      <h3 className={styles.title}>{t('mission.create')}</h3>
      <form onSubmit={handleSubmit} className={styles.form}>
        <div>
          <div className={styles.nameRow}>
            <div className={styles.nameFlex}>
              <label className={styles.label}>{t('mission.missionName').toUpperCase()}</label>
              <input
                className={styles.input}
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. MacBook Pro upgrade"
                required
                autoFocus
              />
            </div>
            <VoiceInputButton
              onTranscript={(text) => {
                setName((prev) => (prev.trim() ? `${prev.trim()} ${text}` : text))
              }}
            />
          </div>
        </div>

        <div>
          <label className={styles.label}>{t('mission.category').toUpperCase()}</label>
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
            <label className={styles.label}>{t('mission.budget').toUpperCase()}</label>
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
            <label className={styles.label}>{t('mission.currency').toUpperCase()}</label>
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

        <div>
          <EnvelopeSelector value={envelopeId} onChange={setEnvelopeId} />
        </div>

        {error && <div className={styles.error}>{error}</div>}

        <div className={styles.actions}>
          <button
            type="submit"
            disabled={loading}
            className={`${styles.submitBtn} ${loading ? styles.submitBtnDisabled : styles.submitBtnActive}`}
          >
            {loading ? t('mission.launching').toUpperCase() : t('mission.launchMission').toUpperCase()}
          </button>
          <button
            type="button"
            onClick={onCancel}
            className={styles.cancelBtn}
          >
            {t('common.cancel')}
          </button>
        </div>
      </form>
    </Card>
  )
}
