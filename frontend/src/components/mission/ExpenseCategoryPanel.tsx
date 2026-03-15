import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { LoadingPulse } from '../scouter'
import { useExpenseCategorizer } from '../../hooks'
import styles from './ExpenseCategoryPanel.module.css'

interface ExpenseCategoryPanelProps {
  missionId: string
  currency: string
  locale: string
}

const categoryColors: Record<string, string> = {
  electronique: '#3b82f6',
  alimentaire: '#10b981',
  mode: '#f59e0b',
  sport: '#ec4899',
  voyage: '#06b6d4',
  maison: '#8b5cf6',
  beaute: '#f97316',
  loisirs: '#a855f7',
  autre: '#6b7280',
}

export function ExpenseCategoryPanel({
  missionId,
  currency,
  locale,
}: ExpenseCategoryPanelProps) {
  const { t } = useTranslation()
  const { data, isLoading } = useExpenseCategorizer(missionId)
  const [isOpen, setIsOpen] = useState(true)

  const fmt = (v: number) => new Intl.NumberFormat(locale, { style: 'currency', currency }).format(v)

  if (isLoading) {
    return <LoadingPulse label={t('common.loading')} />
  }

  if (!data || data.totalItems === 0) {
    return (
      <div className={styles.panel}>
        <div className={styles.header} onClick={() => setIsOpen(!isOpen)}>
          <span className={styles.icon}>📊</span>
          <span className={styles.title}>{t('expenseBreakdown.title', 'Expense Breakdown')}</span>
          <span className={`${styles.toggle} ${isOpen ? '▼' : '▶'}`} />
        </div>
        {isOpen && (
          <div className={styles.content}>
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>📦</div>
              <p className={styles.emptyText}>{t('expenseBreakdown.empty', 'No items yet')}</p>
            </div>
          </div>
        )}
      </div>
    )
  }

  return (
    <div className={styles.panel}>
      <div className={styles.header} onClick={() => setIsOpen(!isOpen)}>
        <span className={styles.icon}>📊</span>
        <span className={styles.title}>{t('expenseBreakdown.title', 'Expense Breakdown')}</span>
        <span className={`${styles.toggle} ${isOpen ? '▼' : '▶'}`} />
      </div>

      {isOpen && (
        <div className={styles.content}>
          {/* Insight box */}
          <div className={styles.insightBox}>
            <p className={styles.insightText}>{data.insight}</p>
          </div>

          {/* Donut bar visualization */}
          <div className={styles.donutBar}>
            {data.breakdowns.map((bd) => (
              <div
                key={bd.category}
                className={styles.donutSegment}
                style={{
                  flex: `${bd.percentage || 0}`,
                  background: categoryColors[bd.category] || '#6b7280',
                  minWidth: '40px',
                }}
                title={`${bd.label}: ${bd.percentage.toFixed(1)}%`}
              >
                {bd.percentage > 8 && `${bd.percentage.toFixed(0)}%`}
              </div>
            ))}
          </div>

          {/* Category list */}
          <ul className={styles.categoryList}>
            {data.breakdowns.map((bd) => (
              <li key={bd.category} className={styles.categoryItem}>
                <div className={styles.itemHeader}>
                  <span className={styles.itemIcon}>{bd.icon}</span>
                  <span className={styles.itemName}>{bd.label}</span>
                  <span className={styles.itemTotal}>{fmt(bd.totalSpent)}</span>
                </div>
                <div className={styles.itemFooter}>
                  <div className={styles.percentBar}>
                    <div
                      className={styles.percentFill}
                      style={{
                        width: `${bd.percentage}%`,
                        background: categoryColors[bd.category] || '#6b7280',
                      }}
                    />
                  </div>
                  <span className={styles.percentText}>{bd.percentage.toFixed(1)}%</span>
                  <span className={styles.itemCount}>{bd.itemCount} {t('expenseBreakdown.items', 'items')}</span>
                </div>
              </li>
            ))}
          </ul>

          {/* Summary row */}
          <div style={{ marginTop: '16px', paddingTop: '16px', borderTop: '1px solid var(--border)' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <strong>{t('expenseBreakdown.total', 'Total')}</strong>
              <span style={{ fontSize: '1.1rem', fontWeight: 700 }}>{fmt(data.totalSpent)}</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.85rem', color: 'var(--text-secondary)', marginTop: '8px' }}>
              <span>{data.totalItems} {t('expenseBreakdown.items', 'items')}</span>
              <span>{data.breakdowns.length} {t('expenseBreakdown.categories', 'categories')}</span>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
