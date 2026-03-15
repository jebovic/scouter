import { useState, useMemo } from 'react'
import { CATEGORY_SEASONS, FRENCH_MONTHS } from '../../utils/seasonalPricing'
import styles from './SeasonalCalendar.module.css'

interface SeasonalCalendarProps {
  compact?: boolean
}

export function SeasonalCalendar({ compact = false }: SeasonalCalendarProps) {
  const [selectedCategoryIndex, setSelectedCategoryIndex] = useState(0)
  const selectedCategory = CATEGORY_SEASONS[selectedCategoryIndex]

  const currentMonth = useMemo(() => new Date().getMonth() + 1, [])

  const getMonthColor = (month: number): string => {
    if (selectedCategory.bestMonths.includes(month)) {
      return 'low'
    }
    if (selectedCategory.worstMonths.includes(month)) {
      return 'high'
    }
    return 'medium'
  }

  const getMonthIndicator = (month: number): string => {
    if (selectedCategory.bestMonths.includes(month)) {
      return '🟢'
    }
    if (selectedCategory.worstMonths.includes(month)) {
      return '🔴'
    }
    return ''
  }

  if (compact) {
    const startMonth = currentMonth === 1 ? 12 : currentMonth - 1
    const months = [
      (startMonth === 12 ? 12 : startMonth) % 12 || 12,
      currentMonth,
      (currentMonth % 12) + 1,
    ]

    return (
      <section className={styles.compactWidget} role="region" aria-label={`Saisonnalité - ${selectedCategory.category}`}>
        <div className={styles.compactHeader}>
          <h3 className={styles.compactTitle}>
            {selectedCategory.icon} {selectedCategory.category}
          </h3>
        </div>

        <div className={styles.compactGrid}>
          {months.map((month) => (
            <div
              key={month}
              className={`${styles.compactCell} ${styles[`level${getMonthColor(month)}`]}`}
              title={`${FRENCH_MONTHS[month - 1]} ${getMonthColor(month) === 'low' ? '✓ Bon mois' : getMonthColor(month) === 'high' ? '✗ Mauvais mois' : '~ Normal'}`}
            >
              <div className={styles.compactMonth}>{FRENCH_MONTHS[month - 1]}</div>
              {month === currentMonth && <div className={styles.currentBorder} />}
            </div>
          ))}
        </div>

        <p className={styles.compactTip}>{selectedCategory.tip}</p>
      </section>
    )
  }

  return (
    <section className={styles.widget} role="region" aria-label={`Saisonnalité - ${selectedCategory.category}`}>
      <div className={styles.header}>
        <h2 className={styles.title}>📅 Saisonnalité des Prix</h2>
      </div>

      <div className={styles.tabs} role="tablist">
        {CATEGORY_SEASONS.map((cat, idx) => (
          <button
            key={cat.category}
            role="tab"
            aria-selected={idx === selectedCategoryIndex}
            className={`${styles.tab} ${idx === selectedCategoryIndex ? styles.tabActive : ''}`}
            onClick={() => setSelectedCategoryIndex(idx)}
            title={cat.category}
          >
            <span className={styles.tabIcon}>{cat.icon}</span>
            <span className={styles.tabLabel}>{cat.category}</span>
          </button>
        ))}
      </div>

      <div className={styles.grid}>
        {Array.from({ length: 12 }, (_, i) => i + 1).map((month) => {
          const priceLevel = getMonthColor(month)
          const isCurrent = month === currentMonth
          return (
            <div
              key={month}
              role="cell"
              className={`${styles.cell} ${styles[`level${priceLevel}`]} ${isCurrent ? styles.cellCurrent : ''}`}
              title={`${FRENCH_MONTHS[month - 1]} - ${priceLevel === 'low' ? '✓ Bon mois' : priceLevel === 'high' ? '✗ Mauvais mois' : '~ Normal'}`}
            >
              <div className={styles.cellHeader}>
                <span className={styles.monthLabel}>{FRENCH_MONTHS[month - 1]}</span>
                {getMonthIndicator(month) && (
                  <span className={styles.indicator}>{getMonthIndicator(month)}</span>
                )}
              </div>
              {isCurrent && <div className={styles.currentHighlight} />}
            </div>
          )
        })}
      </div>

      <div className={styles.callout}>
        <div className={styles.calloutIcon}>{selectedCategory.icon}</div>
        <div className={styles.calloutContent}>
          <p className={styles.calloutTitle}>{selectedCategory.category}</p>
          <p className={styles.calloutTip}>{selectedCategory.tip}</p>
        </div>
      </div>
    </section>
  )
}
