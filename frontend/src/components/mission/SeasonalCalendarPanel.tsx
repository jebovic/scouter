import { useState } from 'react'
import { useSeasonalCalendar } from '../../hooks/useSeasonalCalendar'
import type { MonthRecommendation } from '../../api/seasonalcalendar'
import styles from './SeasonalCalendarPanel.module.css'

interface Props {
  missionId: string
}

function getScoreColor(score: number): string {
  if (score >= 70) return styles.scoreHigh
  if (score >= 50) return styles.scoreMedium
  return styles.scoreLow
}


function MonthCell({ month }: { month: MonthRecommendation }) {
  return (
    <div className={`${styles.monthCell}`}>
      <div className={styles.monthHeader}>
        <span className={styles.monthName}>{month.monthName}</span>
        <span className={`${styles.score} ${getScoreColor(month.score)}`}>
          {month.score}
        </span>
      </div>
      <div className={styles.scoreBar}>
        <div
          className={`${styles.scoreBarFill} ${getScoreColor(month.score)}`}
          style={{ width: `${month.score}%` }}
        />
      </div>
      {(month.events?.length ?? 0) > 0 && (
        <div className={styles.eventDots}>
          {(month.events ?? []).map((event, idx) => (
            <span key={idx} className={styles.eventDot} title={event} />
          ))}
        </div>
      )}
    </div>
  )
}

export function SeasonalCalendarPanel({ missionId }: Props) {
  const { data, isLoading, isError } = useSeasonalCalendar(missionId)
  const [isExpanded, setIsExpanded] = useState(true)

  if (isLoading) {
    return (
      <section className={styles.card} aria-label="Calendrier saisonnier">
        <div className={styles.header}>
          <span className={styles.title}>📅 Calendrier Saisonnier</span>
        </div>
        <div className={`${styles.skeleton} ${styles.skeletonLine}`} />
        <div className={styles.skeletonGrid}>
          {Array.from({ length: 12 }).map((_, i) => (
            <div key={i} className={`${styles.skeleton} ${styles.skeletonCell}`} />
          ))}
        </div>
      </section>
    )
  }

  if (isError || !data) {
    return (
      <section className={styles.card} aria-label="Calendrier saisonnier">
        <div className={styles.header}>
          <span className={styles.title}>📅 Calendrier Saisonnier</span>
        </div>
        <p className={styles.errorState}>Impossible de charger le calendrier saisonnier.</p>
      </section>
    )
  }

  return (
    <section className={styles.card} aria-label="Calendrier saisonnier">
      <div className={styles.header}>
        <button
          className={styles.toggleBtn}
          onClick={() => setIsExpanded(!isExpanded)}
          aria-expanded={isExpanded}
        >
          <span className={styles.title}>📅 Calendrier Saisonnier</span>
          <span className={`${styles.toggleIcon} ${isExpanded ? styles.expanded : ''}`}>▼</span>
        </button>
      </div>

      {isExpanded && (
        <>
          {/* Current month advice */}
          <div className={`${styles.currentMonth} ${getScoreColor(data.currentMonthScore)}`}>
            <div className={styles.currentMonthContent}>
              <div className={styles.currentMonthScore}>{data.currentMonthScore}</div>
              <div className={styles.currentMonthText}>
                <div className={styles.currentMonthLabel}>Ce mois-ci</div>
                <div className={styles.currentMonthAdvice}>{data.currentAdvice}</div>
              </div>
            </div>
          </div>

          {/* Best and worst months summary */}
          <div className={styles.summaryRow}>
            <div className={styles.summaryCard}>
              <h4 className={styles.summaryTitle}>🟢 Meilleurs mois</h4>
              <div className={styles.monthList}>
                {data.bestMonths.map((monthNum) => {
                  const rec = data.monthRecommendations.find((m) => m.month === monthNum)
                  return rec ? (
                    <div key={monthNum} className={styles.bestMonthBadge}>
                      {rec.monthName} ({rec.score})
                    </div>
                  ) : null
                })}
              </div>
            </div>
            <div className={styles.summaryCard}>
              <h4 className={styles.summaryTitle}>🔴 Pires mois</h4>
              <div className={styles.monthList}>
                {data.worstMonths.map((monthNum) => {
                  const rec = data.monthRecommendations.find((m) => m.month === monthNum)
                  return rec ? (
                    <div key={monthNum} className={styles.worstMonthBadge}>
                      {rec.monthName} ({rec.score})
                    </div>
                  ) : null
                })}
              </div>
            </div>
          </div>

          {/* 12-month grid */}
          <div className={styles.monthGrid}>
            {data.monthRecommendations.map((month) => (
              <MonthCell key={month.month} month={month} />
            ))}
          </div>

          {/* Legend */}
          <div className={styles.legend}>
            <div className={styles.legendItem}>
              <span className={`${styles.legendDot} ${styles.scoreHigh}`} />
              <span>Excellent (70+)</span>
            </div>
            <div className={styles.legendItem}>
              <span className={`${styles.legendDot} ${styles.scoreMedium}`} />
              <span>Bon (50-69)</span>
            </div>
            <div className={styles.legendItem}>
              <span className={`${styles.legendDot} ${styles.scoreLow}`} />
              <span>À éviter (&lt;50)</span>
            </div>
          </div>
        </>
      )}
    </section>
  )
}
