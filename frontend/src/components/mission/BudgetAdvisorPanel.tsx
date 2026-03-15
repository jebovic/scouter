import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { LoadingPulse } from '../scouter'
import { useBudgetAdvisor } from '../../hooks'
import styles from './BudgetAdvisorPanel.module.css'

interface BudgetAdvisorPanelProps {
  missionId: string
  currency: string
  locale: string
}

export function BudgetAdvisorPanel({
  missionId,
  currency,
  locale,
}: BudgetAdvisorPanelProps) {
  const { t } = useTranslation()
  const { data: advice, isLoading } = useBudgetAdvisor(missionId)
  const [isCollapsed, setIsCollapsed] = useState(false)

  if (isLoading) {
    return <LoadingPulse label={t('budgetAdvisor.loading')} />
  }

  if (!advice) {
    return null
  }

  const fmt = (v: number) =>
    new Intl.NumberFormat(locale, { style: 'currency', currency }).format(v)

  const surplusColor = advice.surplus >= 0 ? 'positive' : 'negative'

  return (
    <div className={styles.panel}>
      <button
        className={styles.header}
        onClick={() => setIsCollapsed(!isCollapsed)}
        aria-expanded={!isCollapsed}
      >
        <span className={styles.icon}>💡</span>
        <span className={styles.title}>{t('budgetAdvisor.title')}</span>
        <span className={styles.toggle}>{isCollapsed ? '▶' : '▼'}</span>
      </button>

      {!isCollapsed && (
        <div className={styles.content}>
          {/* Summary bar */}
          <div className={styles.summaryBar}>
            <div className={styles.summaryItem}>
              <span className={styles.label}>{t('budgetAdvisor.total')}</span>
              <span className={styles.value}>{fmt(advice.totalBudget)}</span>
            </div>
            <div className={styles.summaryItem}>
              <span className={styles.label}>{t('budgetAdvisor.spent')}</span>
              <span className={`${styles.value} ${styles.spent}`}>
                {fmt(advice.spent)}
              </span>
            </div>
            <div className={styles.summaryItem}>
              <span className={styles.label}>{t('budgetAdvisor.remaining')}</span>
              <span className={styles.value}>{fmt(advice.remaining)}</span>
            </div>
            <div className={styles.summaryItem}>
              <span className={styles.label}>{t('budgetAdvisor.surplus')}</span>
              <span className={`${styles.value} ${styles[surplusColor]}`}>
                {fmt(advice.surplus)}
              </span>
            </div>
          </div>

          {/* Advice text */}
          <div className={styles.adviceBox}>
            <p className={styles.adviceText}>{advice.advice}</p>
          </div>

          {/* Allocations list */}
          {advice.allocations.length > 0 && (
            <div className={styles.allocationsContainer}>
              <h3 className={styles.allocationsTitle}>
                {t('budgetAdvisor.allocations')}
              </h3>
              <ol className={styles.allocationsList}>
                {advice.allocations.map((item) => (
                  <li key={item.itemId} className={styles.allocationItem}>
                    <div className={styles.itemHeader}>
                      <span className={styles.priorityBadge}>
                        #{item.priority}
                      </span>
                      <span className={styles.itemName}>{item.name}</span>
                      <span className={styles.feasibility}>
                        {item.feasible ? (
                          <span className={styles.feasibleIcon}>✓</span>
                        ) : (
                          <span className={styles.infeasibleIcon}>✗</span>
                        )}
                      </span>
                    </div>
                    <div className={styles.itemDetails}>
                      <span className={styles.allocatedBudget}>
                        {fmt(item.allocatedBudget)}
                      </span>
                      <span
                        className={`${styles.reason} ${
                          item.feasible ? styles.feasible : styles.infeasible
                        }`}
                      >
                        {item.reason}
                      </span>
                    </div>
                  </li>
                ))}
              </ol>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
