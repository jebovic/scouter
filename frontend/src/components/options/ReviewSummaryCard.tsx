import { useTranslation } from 'react-i18next'
import { useOptionReviews, useRefreshReviews } from '../../hooks'
import styles from './ReviewSummaryCard.module.css'

interface Props {
  optionId: string
}

function relativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(diff / 60_000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  return `${Math.floor(hrs / 24)}d ago`
}

export function ReviewSummaryCard({ optionId }: Props) {
  const { t } = useTranslation()
  const { summary, isLoading } = useOptionReviews(optionId)
  const { refresh, isPending } = useRefreshReviews(optionId)

  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <span className={styles.title}>{t('reviews.title')}</span>
        <button
          className={styles.refreshBtn}
          disabled={isPending || isLoading}
          onClick={() => refresh()}
        >
          {isPending ? t('reviews.refreshing') : t('reviews.refresh')}
        </button>
      </div>

      {isLoading && !summary && (
        <div style={{ fontSize: '0.8rem', color: 'var(--text-dim)', padding: '0.5rem 0' }}>
          {t('common.loading')}
        </div>
      )}

      {summary && (
        <>
          <div className={styles.gauge}>
            <div
              className={styles.gaugePos}
              style={{ width: `${summary.positivePct}%` }}
            />
            <div
              className={styles.gaugeNeu}
              style={{ width: `${summary.neutralPct}%` }}
            />
            <div
              className={styles.gaugeNeg}
              style={{ width: `${summary.negativePct}%` }}
            />
          </div>

          <div className={styles.legend}>
            <span className={styles.legendItem}>
              <span className={`${styles.dot} ${styles.dotPos}`} />
              {summary.positivePct}% {t('reviews.positive')}
            </span>
            <span className={styles.legendItem}>
              <span className={`${styles.dot} ${styles.dotNeu}`} />
              {summary.neutralPct}% {t('reviews.neutral')}
            </span>
            <span className={styles.legendItem}>
              <span className={`${styles.dot} ${styles.dotNeg}`} />
              {summary.negativePct}% {t('reviews.negative')}
            </span>
          </div>

          <div className={styles.bullets}>
            <div className={styles.section}>
              <div className={styles.sectionTitle}>{t('reviews.pros')}</div>
              {summary.pros.map((p, i) => (
                <div key={i} className={`${styles.bullet} ${styles.bulletPro}`}>{p}</div>
              ))}
            </div>
            <div className={styles.section}>
              <div className={styles.sectionTitle}>{t('reviews.cons')}</div>
              {summary.cons.map((c, i) => (
                <div key={i} className={`${styles.bullet} ${styles.bulletCon}`}>{c}</div>
              ))}
            </div>
          </div>

          <div className={styles.meta}>
            <span>{summary.totalReviews} {t('reviews.reviews')}</span>
            <span>{t('reviews.via')} {summary.source}</span>
            <span>{t('reviews.updated')} {relativeTime(summary.refreshedAt)}</span>
          </div>
        </>
      )}
    </div>
  )
}
