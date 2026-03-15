import { useReviewSummary } from '../../hooks/useReviewSummary'
import styles from './ReviewSummaryCard.module.css'

interface Props {
  itemId: string
  itemName: string
}

function scoreClass(score: number): string {
  if (score >= 8.5) return styles.excellent
  if (score >= 7.0) return styles.good
  if (score >= 5.5) return styles.average
  return styles.poor
}

export function ReviewSummaryCard({ itemId, itemName }: Props) {
  const { summary, isLoading, error } = useReviewSummary(itemId, itemName)

  if (isLoading) {
    return (
      <div className={styles.card}>
        <span className={styles.loading}>Loading review summary…</span>
      </div>
    )
  }

  if (error || !summary) {
    return (
      <div className={styles.card}>
        <span className={styles.error}>Review summary unavailable.</span>
      </div>
    )
  }

  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <span className={styles.title}>Review Summary</span>
        <div className={`${styles.scoreCircle} ${scoreClass(summary.overallScore)}`}>
          {summary.overallScore.toFixed(1)}
        </div>
        <span className={styles.reviewCount}>
          {summary.totalReviews.toLocaleString()} reviews
        </span>
      </div>

      <div className={styles.columns}>
        <div className={styles.col}>
          <span className={`${styles.colLabel} ${styles.pros}`}>Pros</span>
          {summary.pros.map((p) => (
            <span key={p} className={styles.listItem}>
              <span className={styles.bullet}>+</span>
              {p}
            </span>
          ))}
        </div>
        <div className={styles.col}>
          <span className={`${styles.colLabel} ${styles.cons}`}>Cons</span>
          {summary.cons.map((c) => (
            <span key={c} className={styles.listItem}>
              <span className={styles.bullet}>−</span>
              {c}
            </span>
          ))}
        </div>
      </div>

      <p className={styles.verdict}>{summary.verdict}</p>

      <div className={styles.tags}>
        {summary.recommendedFor.length > 0 && (
          <div className={styles.tagRow}>
            <span className={styles.tagLabel}>Best for</span>
            {summary.recommendedFor.map((r) => (
              <span key={r} className={`${styles.tag} ${styles.rec}`}>{r}</span>
            ))}
          </div>
        )}
        {summary.avoidIf.length > 0 && (
          <div className={styles.tagRow}>
            <span className={styles.tagLabel}>Avoid if</span>
            {summary.avoidIf.map((a) => (
              <span key={a} className={`${styles.tag} ${styles.avoid}`}>{a}</span>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
