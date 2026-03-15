import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useWeights, useMatrix, useSetWeight, useDeleteWeight } from '../../hooks/useComparison'
import styles from './ComparisonMatrix.module.css'

const PREDEFINED_CRITERIA = ['price', 'quality', 'warranty', 'brand', 'availability'] as const

interface Props {
  missionId: string
}

export function ComparisonMatrix({ missionId }: Props) {
  const { t } = useTranslation()
  const { weights } = useWeights(missionId)
  const { matrix, refetch } = useMatrix(missionId)
  const { mutate: doSetWeight } = useSetWeight(missionId)
  const { mutate: doDeleteWeight } = useDeleteWeight(missionId)
  const [customInput, setCustomInput] = useState('')

  // Build a map of criteria → weight for quick lookup
  const weightMap = new Map(weights.map((w) => [w.criteria, w.weight]))

  // All criteria to display: predefined + any custom ones from saved weights
  const customCriteria = weights
    .map((w) => w.criteria)
    .filter((c) => !(PREDEFINED_CRITERIA as readonly string[]).includes(c))

  const allCriteria = [...PREDEFINED_CRITERIA, ...customCriteria]

  function handleSliderChange(criteria: string, value: number) {
    doSetWeight({ criteria, weight: value })
  }

  function handleDelete(criteria: string) {
    doDeleteWeight(criteria)
  }

  function handleAddCustom() {
    const trimmed = customInput.trim().toLowerCase().replace(/\s+/g, '_')
    if (!trimmed) return
    doSetWeight({ criteria: trimmed, weight: 50 })
    setCustomInput('')
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter') handleAddCustom()
  }

  const scores = matrix?.scores ?? []

  return (
    <section className={styles.section}>
      <h2 className={styles.title}>{t('comparison.title')}</h2>

      {/* Weight sliders */}
      <div className={styles.weightList}>
        {allCriteria.map((criteria) => {
          const weight = weightMap.get(criteria) ?? 0
          const label =
            (PREDEFINED_CRITERIA as readonly string[]).includes(criteria)
              ? t(`comparison.predefined.${criteria}`, { defaultValue: criteria })
              : criteria
          return (
            <div key={criteria} className={styles.weightRow}>
              <span className={styles.criteriaLabel} title={label}>
                {label}
              </span>
              <input
                type="range"
                min={0}
                max={100}
                value={weight}
                className={styles.slider}
                onChange={(e) => handleSliderChange(criteria, Number(e.target.value))}
                aria-label={`${label} ${t('comparison.weight')}`}
              />
              <span className={styles.weightValue}>{weight}</span>
              <button
                className={styles.deleteBtn}
                onClick={() => handleDelete(criteria)}
                aria-label={`${t('common.delete')} ${label}`}
                title={t('common.delete')}
              >
                ×
              </button>
            </div>
          )
        })}
      </div>

      {/* Add custom criteria */}
      <div className={styles.addRow}>
        <input
          type="text"
          className={styles.addInput}
          placeholder={t('comparison.addCriteria')}
          value={customInput}
          onChange={(e) => setCustomInput(e.target.value)}
          onKeyDown={handleKeyDown}
          maxLength={50}
        />
        <button
          className={styles.addBtn}
          onClick={handleAddCustom}
          disabled={!customInput.trim()}
        >
          {t('common.add')}
        </button>
      </div>

      {/* Matrix results */}
      <div className={styles.matrixSection}>
        <div className={styles.matrixHeader}>
          <span className={styles.matrixTitle}>{t('comparison.score')}</span>
          <button
            className={styles.refreshBtn}
            onClick={() => void refetch()}
          >
            {t('comparison.refresh')}
          </button>
        </div>

        {scores.length === 0 ? (
          <p className={styles.noScores}>{t('comparison.noScores')}</p>
        ) : (
          <table className={styles.table}>
            <thead>
              <tr>
                <th>{t('comparison.rank')}</th>
                <th>{t('comparison.option')}</th>
                <th>{t('comparison.score')}</th>
              </tr>
            </thead>
            <tbody>
              {scores.map((s) => (
                <tr
                  key={s.optionId}
                  className={s.rank === 1 ? styles.rankOne : undefined}
                >
                  <td className={styles.rankCell}>#{s.rank}</td>
                  <td className={styles.optionNameCell} title={s.optionName}>
                    {s.optionName}
                  </td>
                  <td className={styles.scoreCell}>
                    <div className={styles.scoreBar}>
                      <div className={styles.scoreBarTrack}>
                        <div
                          className={styles.scoreBarFill}
                          style={{ '--score-pct': `${s.score}%` } as React.CSSProperties}
                        />
                      </div>
                      <span className={styles.scoreText}>{s.score.toFixed(1)}%</span>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </section>
  )
}
