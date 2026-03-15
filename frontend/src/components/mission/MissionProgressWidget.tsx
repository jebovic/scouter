import React from 'react'
import { useMissionProgress } from '../../hooks/useMissionProgress'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import styles from './MissionProgressWidget.module.css'

interface Props {
  missionId: string | undefined
}

const RADIUS = 28
const CIRCUMFERENCE = 2 * Math.PI * RADIUS

export function MissionProgressWidget({ missionId }: Props) {
  const { data, isLoading, isError } = useMissionProgress(missionId)

  if (isLoading) {
    return (
      <section className={styles.card} aria-label="Progression de la mission">
        <div className={styles.header}>
          <span className={styles.title}>Progression</span>
        </div>
        <div className={styles.hero}>
          <div className={`${styles.skeleton} ${styles.skeletonCircle}`} />
          <div className={`${styles.skeleton} ${styles.skeletonLine}`} />
        </div>
        <div className={styles.skeletonGrid}>
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className={`${styles.skeleton} ${styles.skeletonCell}`} />
          ))}
        </div>
      </section>
    )
  }

  if (isError || !data) {
    return (
      <section className={styles.card} aria-label="Progression de la mission">
        <div className={styles.header}>
          <span className={styles.title}>Progression</span>
        </div>
        <p className={styles.errorState}>Impossible de charger la progression.</p>
      </section>
    )
  }

  if (data.totalItems === 0) {
    return (
      <section className={styles.card} aria-label="Progression de la mission">
        <div className={styles.header}>
          <span className={styles.title}>Progression</span>
        </div>
        <p className={styles.empty}>Aucun article dans la liste de courses.</p>
      </section>
    )
  }

  const clampedCompletion = Math.min(data.completionPct, 100)
  const dashOffset = CIRCUMFERENCE * (1 - clampedCompletion / 100)

  const clampedBudget = Math.min(data.budgetPct, 100)
  const isOverBudget = data.budgetPct > 100

  const { fmt } = useFormatCurrency()

  return (
    <section className={styles.card} aria-label="Progression de la mission">
      <div className={styles.header}>
        <span className={styles.title}>Progression</span>
      </div>

      {/* Completion ring + budget bar */}
      <div className={styles.hero}>
        {/* SVG circular progress ring */}
        <div className={styles.ringWrap}>
          <svg width="72" height="72" viewBox="0 0 72 72" aria-hidden="true">
            <circle
              className={styles.ringBg}
              cx="36"
              cy="36"
              r={RADIUS}
              strokeWidth="6"
            />
            <circle
              className={styles.ringFill}
              cx="36"
              cy="36"
              r={RADIUS}
              strokeWidth="6"
              strokeDasharray={CIRCUMFERENCE}
              strokeDashoffset={dashOffset}
            />
          </svg>
          <div className={styles.ringLabel}>
            <span className={styles.ringPct}>{Math.round(clampedCompletion)}%</span>
            <span className={styles.ringUnit}>acheté</span>
          </div>
        </div>

        {/* Budget progress */}
        <div className={styles.budgetBlock}>
          <span className={styles.budgetLabel}>Budget</span>
          <div className={styles.budgetTrack}>
            <div
              className={styles.budgetFill}
              data-over={isOverBudget ? 'true' : undefined}
              style={{ '--pct': clampedBudget } as React.CSSProperties}
            />
          </div>
          <div className={styles.budgetMeta}>
            <span>{fmt(data.budgetUsed)} dépensé</span>
            <span>{fmt(data.budgetTotal)}</span>
          </div>
        </div>
      </div>

      {/* Status breakdown */}
      <div className={styles.breakdown}>
        <div className={styles.stat}>
          <span className={styles.statCount} data-status="buy">{data.boughtItems}</span>
          <span className={styles.statLabel}>Achetés</span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statCount} data-status="watch">{data.watchingItems}</span>
          <span className={styles.statLabel}>Surveillés</span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statCount} data-status="defer">{data.deferredItems}</span>
          <span className={styles.statLabel}>Différés</span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statCount} data-status="flash-sale">{data.flashSaleItems}</span>
          <span className={styles.statLabel}>Flash sale</span>
        </div>
      </div>

      {/* Top saving badge */}
      {data.topSaving && (
        <div className={styles.savingBadge} aria-label="Meilleure économie">
          <span className={styles.savingIcon}>💰</span>
          <div className={styles.savingBody}>
            <div className={styles.savingName}>{data.topSaving.itemName}</div>
            <div className={styles.savingAmount}>
              -{fmt(data.topSaving.saving)} vs objectif
            </div>
          </div>
        </div>
      )}
    </section>
  )
}
