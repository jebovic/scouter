import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import type { Option, OptionBadge } from '../../types'
import styles from './OptionEditModal.module.css'

const BADGES: OptionBadge[] = ['recommended', 'alternative', 'watch', 'rejected']

interface OptionEditModalProps {
  option: Option
  onSave: (updates: Partial<Option>) => void
  onClose: () => void
  loading?: boolean
  error?: string
}

export function OptionEditModal({ option, onSave, onClose, loading, error }: OptionEditModalProps) {
  const { t } = useTranslation()
  const [name, setName] = useState(option.name)
  const [badge, setBadge] = useState<OptionBadge>(option.badge)
  const [priceMin, setPriceMin] = useState(String(option.priceRange?.min ?? ''))
  const [priceMax, setPriceMax] = useState(String(option.priceRange?.max ?? ''))
  const [notes, setNotes] = useState(option.notes ?? '')
  const [warnings, setWarnings] = useState(
    (option.warnings ?? []).join(', ')
  )

  useEffect(() => {
    function handleKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', handleKey)
    return () => document.removeEventListener('keydown', handleKey)
  }, [onClose])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    onSave({
      name,
      badge,
      priceRange: {
        min: parseFloat(priceMin) || 0,
        max: parseFloat(priceMax) || 0,
        best: option.priceRange?.best ?? (parseFloat(priceMin) || 0),
      },
      notes,
      warnings: warnings ? warnings.split(',').map((w) => w.trim()).filter(Boolean) : [],
    })
  }

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.container} onClick={(e) => e.stopPropagation()}>
        <div className={styles.header}>
          <h2 className={styles.title}>{t('option.actions.edit')}</h2>
          <button className={styles.closeBtn} onClick={onClose} aria-label={t('common.close')}>×</button>
        </div>

        <form aria-label="form" onSubmit={handleSubmit} className={styles.form}>
          {error && <p className={styles.error}>{error}</p>}

          <label className={styles.label}>{t('common.name')}</label>
          <input
            className={styles.input}
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />

          <label className={styles.label}>{t('option.actions.badgeLabel')}</label>
          <select className={styles.select} value={badge} onChange={(e) => setBadge(e.target.value as OptionBadge)}>
            {BADGES.map((b) => (
              <option key={b} value={b}>{t(`options.badge.${b}`)}</option>
            ))}
          </select>

          <div className={styles.priceRow}>
            <div>
              <label className={styles.label}>{t('common.priceMin')}</label>
              <input
                className={styles.input}
                type="number"
                value={priceMin}
                onChange={(e) => setPriceMin(e.target.value)}
              />
            </div>
            <div>
              <label className={styles.label}>{t('common.priceMax')}</label>
              <input
                className={styles.input}
                type="number"
                value={priceMax}
                onChange={(e) => setPriceMax(e.target.value)}
              />
            </div>
          </div>

          <label className={styles.label}>{t('common.notes')}</label>
          <textarea className={styles.textarea} value={notes} onChange={(e) => setNotes(e.target.value)} rows={3} />

          <label className={styles.label}>{t('common.warnings')}</label>
          <textarea className={styles.textarea} value={warnings} onChange={(e) => setWarnings(e.target.value)} rows={2} />

          {option.attributes.length > 0 && (
            <div className={styles.attributesSection}>
              <p className={styles.attributesLabel}>{t('common.attributes')} ({t('common.readOnly')})</p>
              <ul className={styles.attributeList}>
                {option.attributes.map((attr) => (
                  <li key={attr.key} className={styles.attributeItem}>
                    <span className={styles.attrLabel}>{attr.label}</span>
                    <span className={styles.attrValue}>{String(attr.value)}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          <div className={styles.actions}>
            <button type="button" onClick={onClose}>{t('common.cancel')}</button>
            <button type="submit" className={styles.saveBtn} disabled={loading}>
              {loading ? t('common.saving') : t('common.save')}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
