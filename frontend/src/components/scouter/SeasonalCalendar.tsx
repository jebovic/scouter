import { useState, useMemo } from 'react'
import { Calendar } from 'lucide-react'
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

  const getMonthLevel = (month: number): 'good' | 'bad' | null => {
    if (selectedCategory.bestMonths.includes(month)) return 'good'
    if (selectedCategory.worstMonths.includes(month)) return 'bad'
    return null
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
        <h2 className={styles.title}>
          <Calendar size={18} aria-hidden="true" className={styles.titleIcon} />
          Saisonnalité des Prix
        </h2>
      </div>

      <div className={styles.tabs} role="tablist">
        {CATEGORY_SEASONS.map((cat, idx) => (
          <button
            key={cat.category}
            role="tab"
            aria-selected={idx === selectedCategoryIndex}
            aria-label={cat.category}
            className={`${styles.tab} ${idx === selectedCategoryIndex ? styles.tabActive : ''}`}
            onClick={() => setSelectedCategoryIndex(idx)}
            title={cat.category}
          >
            <span className={styles.tabIcon} aria-hidden="true">{cat.icon}</span>
            <span className={styles.tabLabel}>{cat.category}</span>
          </button>
        ))}
      </div>

      <div className={styles.grid}>
        {Array.from({ length: 12 }, (_, i) => i + 1).map((month) => {
          const priceLevel = getMonthColor(month)
          const level = getMonthLevel(month)
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
                {level === 'good' && <span className={styles.dotGood} aria-hidden="true" />}
                {level === 'bad' && <span className={styles.dotBad} aria-hidden="true" />}
              </div>
              {isCurrent && <div className={styles.currentHighlight} />}
            </div>
          )
        })}
      </div>

      <div className={styles.callout}>
        <div className={styles.calloutIcon} aria-hidden="true">{selectedCategory.icon}</div>
        <div className={styles.calloutContent}>
          <p className={styles.calloutTitle}>{selectedCategory.category}</p>
          <p className={styles.calloutTip}>{selectedCategory.tip}</p>
        </div>
      </div>
    </section>
  )
}
